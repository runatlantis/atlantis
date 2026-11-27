// Copyright 2026 The Atlantis Authors
// SPDX-License-Identifier: Apache-2.0

package events

import (
	"fmt"
	"sort"
	"strings"

	"github.com/runatlantis/atlantis/server/events/models"
)

const (
	// maxDiagramNodes bounds the size of the rendered graph so that a large
	// plan still produces a readable comment.
	maxDiagramNodes = 60
	// maxDiagramEdges bounds the number of edges for the same reason.
	maxDiagramEdges = 300
)

// RenderPlanDiagramMarkdown renders a plan's resource graph as a Mermaid
// flowchart, coloured by the planned action, together with a one-line summary.
// It returns an empty string when the diagram has nothing to show: no
// resources, or no infrastructure change at all.
//
// Mermaid is used because the layout is computed by its dagre engine, so the
// same plan always renders to the same picture, and because GitHub, GitLab and
// Gitea render Mermaid inside comments without Atlantis hosting any image.
func RenderPlanDiagramMarkdown(diagram *models.PlanDiagram) string {
	if diagram.Empty() || !diagramChangesInfrastructure(diagram.Nodes) {
		return ""
	}
	nodes, edges, omitted := selectDiagram(diagram)

	// Node ids are assigned from the sorted node list, so the same plan always
	// produces the same diagram, which keeps the layout engine's output stable.
	nodeIDs := make(map[string]string, len(nodes))
	for i, node := range nodes {
		nodeIDs[node.Address] = fmt.Sprintf("dgr%d", i)
	}

	var b strings.Builder
	b.WriteString("**Planned infrastructure changes:** ")
	b.WriteString(diagramSummary(diagram.Nodes))
	b.WriteString("\n\n```mermaid\nflowchart LR\n")
	b.WriteString(diagramClassDefs)
	nodesByModule := groupNodesByModule(nodes)
	for i, module := range nodesByModule.modules {
		if module != "" {
			fmt.Fprintf(&b, "  subgraph dgm%d[\"%s\"]\n", i, mermaidLabel(module))
		}
		for _, node := range nodesByModule.nodes[module] {
			depth := "  "
			if module != "" {
				depth = "    "
			}
			fmt.Fprintf(&b, "%s%s[\"%s\"]:::%s\n", depth, nodeIDs[node.Address], diagramNodeLabel(node), diagramNodeClass(node.Action))
		}
		if module != "" {
			b.WriteString("  end\n")
		}
	}
	for _, edge := range edges {
		fmt.Fprintf(&b, "  %s --> %s\n", nodeIDs[edge.From], nodeIDs[edge.To])
	}
	b.WriteString("```\n")
	if omitted > 0 {
		fmt.Fprintf(&b, "\n_Showing %d of %d resources; %d omitted._\n", len(nodes), len(diagram.Nodes), omitted)
	}
	return b.String()
}

// diagramClassDefs is the colour key of the diagram: green adds, amber changes,
// red replaces, dashed red destroys and grey leaves the resource untouched.
const diagramClassDefs = `  classDef dgAdd fill:#2da44e,stroke:#1a7f37,color:#ffffff
  classDef dgChange fill:#bf8700,stroke:#7d4e00,color:#ffffff
  classDef dgReplace fill:#cf222e,stroke:#82071e,color:#ffffff
  classDef dgDestroy fill:#ffebe9,stroke:#cf222e,color:#82071e,stroke-dasharray:5 3
  classDef dgUnchanged fill:#f6f8fa,stroke:#8c959f,color:#57606a
`

func diagramNodeClass(action models.DiagramAction) string {
	switch action {
	case models.DiagramActionCreate:
		return "dgAdd"
	case models.DiagramActionUpdate:
		return "dgChange"
	case models.DiagramActionReplace:
		return "dgReplace"
	case models.DiagramActionDelete, models.DiagramActionForget:
		return "dgDestroy"
	default:
		return "dgUnchanged"
	}
}

func diagramNodeLabel(node models.DiagramNode) string {
	var annotation string
	switch node.Action {
	case models.DiagramActionCreate:
		annotation = "+ add"
	case models.DiagramActionUpdate:
		annotation = "~ change" + attributeCountSuffix(node)
	case models.DiagramActionReplace:
		annotation = "± replace" + attributeCountSuffix(node)
	case models.DiagramActionDelete:
		annotation = "- destroy"
	case models.DiagramActionForget:
		annotation = "- forget"
	case models.DiagramActionRead:
		annotation = "read"
	default:
		annotation = "unchanged"
	}
	return mermaidLabel(node.Address) + "<br/>" + mermaidLabel(annotation)
}

func attributeCountSuffix(node models.DiagramNode) string {
	switch count := node.ChangedAttributes; count {
	case 0:
		return ""
	case 1:
		return " (1 attribute)"
	default:
		return fmt.Sprintf(" (%d attributes)", count)
	}
}

// diagramSummary counts the plan's actions over every resource, including the
// ones the rendered graph leaves out.
func diagramSummary(nodes []models.DiagramNode) string {
	counts := map[models.DiagramAction]int{}
	for _, node := range nodes {
		counts[node.Action]++
	}
	var parts []string
	appendCount := func(count int, singular, plural string) {
		switch count {
		case 0:
			return
		case 1:
			parts = append(parts, "1 to "+singular)
		default:
			parts = append(parts, fmt.Sprintf("%d to %s", count, plural))
		}
	}
	appendCount(counts[models.DiagramActionCreate], "add", "add")
	appendCount(counts[models.DiagramActionUpdate], "change", "change")
	appendCount(counts[models.DiagramActionReplace], "replace", "replace")
	appendCount(counts[models.DiagramActionDelete], "destroy", "destroy")
	appendCount(counts[models.DiagramActionForget], "forget", "forget")
	if unchanged := counts[models.DiagramActionNoOp] + counts[models.DiagramActionRead]; unchanged > 0 {
		parts = append(parts, fmt.Sprintf("%d unchanged", unchanged))
	}
	if len(parts) == 0 {
		return "nothing to change."
	}
	return strings.Join(parts, ", ") + "."
}

func diagramChangesInfrastructure(nodes []models.DiagramNode) bool {
	for _, node := range nodes {
		if node.Action.ChangesInfrastructure() {
			return true
		}
	}
	return false
}

type moduleNodes struct {
	modules []string
	nodes   map[string][]models.DiagramNode
}

func groupNodesByModule(nodes []models.DiagramNode) moduleNodes {
	grouped := moduleNodes{nodes: make(map[string][]models.DiagramNode, 1)}
	for _, node := range nodes {
		if _, ok := grouped.nodes[node.Module]; !ok {
			grouped.modules = append(grouped.modules, node.Module)
		}
		grouped.nodes[node.Module] = append(grouped.nodes[node.Module], node)
	}
	// The root module first, then modules by name, so the layout does not
	// depend on map iteration order.
	sort.Strings(grouped.modules)
	return grouped
}

// selectDiagram bounds the graph. Changed resources are always drawn; the rest
// of the budget goes to their direct neighbours, which give the changes
// context without dragging in the whole state.
func selectDiagram(diagram *models.PlanDiagram) ([]models.DiagramNode, []models.DiagramEdge, int) {
	if len(diagram.Nodes) <= maxDiagramNodes {
		return diagram.Nodes, capEdges(diagram.Edges, nil), 0
	}

	adjacent := make(map[string][]string, len(diagram.Edges)*2)
	for _, edge := range diagram.Edges {
		adjacent[edge.From] = append(adjacent[edge.From], edge.To)
		adjacent[edge.To] = append(adjacent[edge.To], edge.From)
	}

	selected := make(map[string]bool, maxDiagramNodes)
	var order []string
	add := func(address string) {
		if len(order) >= maxDiagramNodes || selected[address] {
			return
		}
		selected[address] = true
		order = append(order, address)
	}
	for _, node := range diagram.Nodes {
		if node.Action.ChangesInfrastructure() {
			add(node.Address)
		}
	}
	for i := 0; i < len(order); i++ {
		for _, neighbour := range adjacent[order[i]] {
			add(neighbour)
		}
	}

	var nodes []models.DiagramNode
	for _, node := range diagram.Nodes {
		if selected[node.Address] {
			nodes = append(nodes, node)
		}
	}
	return nodes, capEdges(diagram.Edges, selected), len(diagram.Nodes) - len(nodes)
}

// capEdges keeps at most maxDiagramEdges edges. When keep is non-nil, edges
// with an endpoint outside the selection are dropped as well.
func capEdges(edges []models.DiagramEdge, keep map[string]bool) []models.DiagramEdge {
	capped := make([]models.DiagramEdge, 0, min(len(edges), maxDiagramEdges))
	for _, edge := range edges {
		if keep != nil && (!keep[edge.From] || !keep[edge.To]) {
			continue
		}
		if len(capped) >= maxDiagramEdges {
			break
		}
		capped = append(capped, edge)
	}
	return capped
}

// mermaidLabel escapes text that is not safe inside a quoted Mermaid node
// label. Mermaid entity escapes are used so the characters survive rendering.
func mermaidLabel(text string) string {
	return strings.NewReplacer(
		"&", "#38;",
		`"`, "#quot;",
		"<", "#lt;",
		">", "#gt;",
		"#", "#35;",
	).Replace(text)
}
