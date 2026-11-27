// Copyright 2026 The Atlantis Authors
// SPDX-License-Identifier: Apache-2.0

package events

import (
	"bytes"
	"encoding/json"
	"fmt"
	"sort"
	"strings"

	"github.com/runatlantis/atlantis/server/events/models"
)

// maxModuleReferenceFanOut bounds how many resources one reference to a module
// ("module.vpc") is expanded into, so that a single output reference cannot
// swamp the diagram.
const maxModuleReferenceFanOut = 25

// planDiagramJSON is the subset of `terraform show -json <planfile>` needed to
// draw a plan's resource graph: what changes, and which resource refers to
// which.
type planDiagramJSON struct {
	ResourceChanges []planDiagramResourceChange `json:"resource_changes"`
	Configuration   struct {
		RootModule json.RawMessage `json:"root_module"`
	} `json:"configuration"`
}

type planDiagramResourceChange struct {
	Address       string `json:"address"`
	ModuleAddress string `json:"module_address"`
	Change        struct {
		Actions      []string        `json:"actions"`
		Before       json.RawMessage `json:"before"`
		After        json.RawMessage `json:"after"`
		AfterUnknown json.RawMessage `json:"after_unknown"`
	} `json:"change"`
}

// ParsePlanDiagram builds the resource dependency graph of a Terraform plan.
// Nodes come from resource_changes and edges from the references of the
// plan's configuration, so an edge exists only when the plan itself knows the
// link. Plans without resource_changes (for example a Terraform Cloud plan
// stand-in) produce an empty diagram and no error.
func ParsePlanDiagram(planJSON []byte) (*models.PlanDiagram, error) {
	var parsed planDiagramJSON
	if err := json.Unmarshal(planJSON, &parsed); err != nil {
		return nil, fmt.Errorf("parsing plan JSON: %w", err)
	}
	if len(parsed.ResourceChanges) == 0 {
		return nil, nil
	}

	nodes := make([]models.DiagramNode, 0, len(parsed.ResourceChanges))
	for _, change := range parsed.ResourceChanges {
		if change.Address == "" {
			continue
		}
		node := models.DiagramNode{
			Address: change.Address,
			Module:  change.ModuleAddress,
			Action:  diagramAction(change.Change.Actions),
		}
		if node.Action == models.DiagramActionUpdate || node.Action == models.DiagramActionReplace {
			node.ChangedAttributes = changedAttributeCount(change.Change.Before, change.Change.After, change.Change.AfterUnknown)
		}
		nodes = append(nodes, node)
	}
	if len(nodes) == 0 {
		return nil, nil
	}

	collector := newDiagramCollector(nodes)
	if len(parsed.Configuration.RootModule) > 0 {
		var root map[string]any
		if err := json.Unmarshal(parsed.Configuration.RootModule, &root); err != nil {
			return nil, fmt.Errorf("parsing root module configuration: %w", err)
		}
		collector.walkModule(root, "")
	}

	sort.Slice(nodes, func(i, j int) bool { return nodes[i].Address < nodes[j].Address })

	return &models.PlanDiagram{
		Nodes: nodes,
		Edges: collector.sortedEdges(),
	}, nil
}

// diagramAction maps the actions of a Terraform plan change to the single verb
// the diagram displays.
func diagramAction(actions []string) models.DiagramAction {
	var create, update, destroy bool
	for _, action := range actions {
		switch action {
		case "create":
			create = true
		case "update":
			update = true
		case "delete":
			destroy = true
		case "read":
			return models.DiagramActionRead
		case "forget":
			return models.DiagramActionForget
		}
	}
	switch {
	case create && destroy:
		return models.DiagramActionReplace
	case create:
		return models.DiagramActionCreate
	case destroy:
		return models.DiagramActionDelete
	case update:
		return models.DiagramActionUpdate
	default:
		return models.DiagramActionNoOp
	}
}

// diagramCollector resolves references to the resources they point at.
type diagramCollector struct {
	// byNormalizedAddress maps an address with instance keys removed to the
	// addresses of every instance of that resource.
	byNormalizedAddress map[string][]string
	edges               map[[2]string]struct{}
}

func newDiagramCollector(nodes []models.DiagramNode) *diagramCollector {
	c := &diagramCollector{
		byNormalizedAddress: make(map[string][]string, len(nodes)),
		edges:               make(map[[2]string]struct{}),
	}
	for _, node := range nodes {
		normalized := normalizeAddress(node.Address)
		c.byNormalizedAddress[normalized] = append(c.byNormalizedAddress[normalized], node.Address)
	}
	return c
}

func (c *diagramCollector) walkModule(module map[string]any, prefix string) {
	rawResources, _ := module["resources"].([]any)
	for _, entry := range rawResources {
		resource, ok := entry.(map[string]any)
		if !ok {
			continue
		}
		address, _ := resource["address"].(string)
		if address == "" {
			continue
		}
		dependents := c.byNormalizedAddress[normalizeAddress(joinAddress(prefix, address))]
		if len(dependents) == 0 {
			// Counted resources are absent from the plan when they have no
			// instances, so there is nothing to draw.
			continue
		}
		for _, ref := range referencesIn(resource) {
			for _, dependency := range c.resolveReference(ref, prefix) {
				for _, dependent := range dependents {
					c.addEdge(dependency, dependent)
				}
			}
		}
	}
	moduleCalls, _ := module["module_calls"].(map[string]any)
	names := make([]string, 0, len(moduleCalls))
	for name := range moduleCalls {
		names = append(names, name)
	}
	sort.Strings(names)
	for _, name := range names {
		call, ok := moduleCalls[name].(map[string]any)
		if !ok {
			continue
		}
		childPrefix := joinAddress(prefix, "module."+name)
		// The call's own expressions feed the module's inputs, so they depend
		// on resources in the calling module. The "module" key holds the child
		// module's configuration and is walked separately.
		for _, ref := range referencesIn(withoutKey(call, "module")) {
			for _, dependency := range c.resolveReference(ref, prefix) {
				for _, dependent := range c.resourcesUnder(childPrefix) {
					c.addEdge(dependency, dependent)
				}
			}
		}
		if child, ok := call["module"].(map[string]any); ok {
			c.walkModule(child, childPrefix)
		}
	}
}

func withoutKey(m map[string]any, key string) map[string]any {
	copied := make(map[string]any, len(m))
	for k, v := range m {
		if k != key {
			copied[k] = v
		}
	}
	return copied
}

// referencesIn collects every "references" list in a configuration object,
// wherever it sits: plain arguments, nested blocks or provisioners.
func referencesIn(value any) []string {
	var refs []string
	var walk func(any)
	walk = func(v any) {
		switch typed := v.(type) {
		case map[string]any:
			if list, ok := typed["references"].([]any); ok {
				for _, entry := range list {
					if ref, ok := entry.(string); ok && ref != "" {
						refs = append(refs, ref)
					}
				}
			}
			for key, child := range typed {
				if key == "references" {
					continue
				}
				walk(child)
			}
		case []any:
			for _, child := range typed {
				walk(child)
			}
		}
	}
	walk(value)
	sort.Strings(refs)
	return refs
}

// resolveReference returns the addresses of the resources a reference points
// at. References name either a resource ("aws_vpc.main"), an attribute of one
// ("aws_vpc.main.id", "aws_subnet.main[*].id") or a module ("module.vpc").
func (c *diagramCollector) resolveReference(ref, modulePrefix string) []string {
	relative := normalizeAddress(ref)
	full := normalizeAddress(joinAddress(modulePrefix, ref))
	if addresses, ok := c.byNormalizedAddress[full]; ok {
		return addresses
	}
	// Longest prefix that is a resource address wins, so
	// "aws_subnet.main.id" resolves to "aws_subnet.main".
	parts := strings.Split(full, ".")
	for i := len(parts) - 1; i > 0; i-- {
		if addresses, ok := c.byNormalizedAddress[strings.Join(parts[:i], ".")]; ok {
			return addresses
		}
	}
	// A module reference stands for every resource the module declares. Only
	// "module." references qualify: "var.x" inside a module is not a module.
	if strings.HasPrefix(relative, "module.") {
		for _, path := range modulePaths(full) {
			if addresses := c.resourcesUnder(path); len(addresses) > 0 {
				return addresses
			}
		}
	}
	return nil
}

// modulePaths lists the module address candidates of a reference, longest
// first: "module.vpc.module.db.x" gives "module.vpc.module.db", then
// "module.vpc". A module address is a repeated "module.<name>" pair, so only
// even-length prefixes ending in a name qualify.
func modulePaths(ref string) []string {
	parts := strings.Split(ref, ".")
	var paths []string
	for length := len(parts); length >= 2; length-- {
		if parts[length-2] != "module" {
			continue
		}
		paths = append(paths, strings.Join(parts[:length], "."))
	}
	return paths
}

// resourcesUnder returns the addresses of the resources declared in a module,
// bounded by maxModuleReferenceFanOut.
func (c *diagramCollector) resourcesUnder(moduleAddress string) []string {
	normalized := make([]string, 0, len(c.byNormalizedAddress))
	for address := range c.byNormalizedAddress {
		if strings.HasPrefix(address, moduleAddress+".") {
			normalized = append(normalized, address)
		}
	}
	sort.Strings(normalized)
	var addresses []string
	for _, address := range normalized {
		addresses = append(addresses, c.byNormalizedAddress[address]...)
		if len(addresses) >= maxModuleReferenceFanOut {
			return addresses[:maxModuleReferenceFanOut]
		}
	}
	return addresses
}

func (c *diagramCollector) addEdge(from, to string) {
	if from == "" || to == "" || from == to {
		return
	}
	c.edges[[2]string{from, to}] = struct{}{}
}

func (c *diagramCollector) sortedEdges() []models.DiagramEdge {
	edges := make([]models.DiagramEdge, 0, len(c.edges))
	for edge := range c.edges {
		edges = append(edges, models.DiagramEdge{From: edge[0], To: edge[1]})
	}
	sort.Slice(edges, func(i, j int) bool {
		if edges[i].From != edges[j].From {
			return edges[i].From < edges[j].From
		}
		return edges[i].To < edges[j].To
	})
	return edges
}

// joinAddress joins two Terraform addresses without introducing a leading or
// trailing dot for the root module.
func joinAddress(prefix, address string) string {
	if prefix == "" {
		return address
	}
	return prefix + "." + address
}

// normalizeAddress removes instance keys ("[0]", `["blue"]`) so that a
// reference can be matched against every instance of the resource it names.
func normalizeAddress(address string) string {
	var b strings.Builder
	depth := 0
	for _, r := range address {
		switch r {
		case '[':
			depth++
		case ']':
			if depth > 0 {
				depth--
			}
		default:
			if depth == 0 {
				b.WriteRune(r)
			}
		}
	}
	return b.String()
}

// changedAttributeCount counts the top-level attributes the plan changes. A
// value that becomes known only after apply counts as changed.
func changedAttributeCount(before, after, afterUnknown json.RawMessage) int {
	var beforeMap, afterMap, unknownMap map[string]json.RawMessage
	if len(before) == 0 || json.Unmarshal(before, &beforeMap) != nil || beforeMap == nil {
		return 0
	}
	_ = json.Unmarshal(after, &afterMap)
	_ = json.Unmarshal(afterUnknown, &unknownMap)

	changed := 0
	for key, previous := range beforeMap {
		if unknown, ok := unknownMap[key]; ok && !isFalseJSON(unknown) {
			changed++
			continue
		}
		if !jsonValuesEqual(previous, afterMap[key]) {
			changed++
		}
	}
	for key := range afterMap {
		if _, existed := beforeMap[key]; existed {
			continue
		}
		if unknown, ok := unknownMap[key]; ok && !isFalseJSON(unknown) {
			continue
		}
		changed++
	}
	return changed
}

func isFalseJSON(raw json.RawMessage) bool {
	trimmed := strings.TrimSpace(string(raw))
	return trimmed == "false" || trimmed == "null" || trimmed == ""
}

func jsonValuesEqual(a, b json.RawMessage) bool {
	compactA, errA := compactJSON(a)
	compactB, errB := compactJSON(b)
	if errA != nil || errB != nil {
		return bytes.Equal(bytes.TrimSpace(a), bytes.TrimSpace(b))
	}
	return bytes.Equal(compactA, compactB)
}

func compactJSON(raw json.RawMessage) ([]byte, error) {
	var buffer bytes.Buffer
	if err := json.Compact(&buffer, raw); err != nil {
		return nil, err
	}
	return buffer.Bytes(), nil
}
