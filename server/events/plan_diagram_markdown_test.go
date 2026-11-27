// Copyright 2026 The Atlantis Authors
// SPDX-License-Identifier: Apache-2.0

package events_test

import (
	"fmt"
	"strings"
	"testing"

	"github.com/runatlantis/atlantis/server/events"
	"github.com/runatlantis/atlantis/server/events/models"
	. "github.com/runatlantis/atlantis/testing"
)

const testClassDefs = `  classDef dgAdd fill:#2da44e,stroke:#1a7f37,color:#ffffff
  classDef dgChange fill:#bf8700,stroke:#7d4e00,color:#ffffff
  classDef dgReplace fill:#cf222e,stroke:#82071e,color:#ffffff
  classDef dgDestroy fill:#ffebe9,stroke:#cf222e,color:#82071e,stroke-dasharray:5 3
  classDef dgUnchanged fill:#f6f8fa,stroke:#8c959f,color:#57606a
`

func TestRenderPlanDiagramMarkdown(t *testing.T) {
	diagram := &models.PlanDiagram{
		Nodes: []models.DiagramNode{
			{Address: "aws_subnet.public", Action: models.DiagramActionNoOp},
			{Address: "aws_vpc.main", Action: models.DiagramActionCreate},
		},
		Edges: []models.DiagramEdge{{From: "aws_vpc.main", To: "aws_subnet.public"}},
	}

	expected := "**Planned infrastructure changes:** 1 to add, 1 unchanged.\n\n" +
		"```mermaid\nflowchart LR\n" +
		testClassDefs +
		"  dgr0[\"aws_subnet.public<br/>unchanged\"]:::dgUnchanged\n" +
		"  dgr1[\"aws_vpc.main<br/>+ add\"]:::dgAdd\n" +
		"  dgr1 --> dgr0\n" +
		"```\n"

	Equals(t, expected, events.RenderPlanDiagramMarkdown(diagram))
}

func TestRenderPlanDiagramMarkdown_ActionsAndCounts(t *testing.T) {
	diagram := &models.PlanDiagram{
		Nodes: []models.DiagramNode{
			{Address: "aws_instance.a", Action: models.DiagramActionUpdate, ChangedAttributes: 2},
			{Address: "aws_instance.b", Action: models.DiagramActionUpdate, ChangedAttributes: 0},
			{Address: "aws_instance.c", Action: models.DiagramActionReplace, ChangedAttributes: 1},
			{Address: "aws_instance.d", Action: models.DiagramActionDelete},
			{Address: "data.aws_ami.e", Action: models.DiagramActionRead},
		},
	}

	rendered := events.RenderPlanDiagramMarkdown(diagram)

	Assert(t, strings.HasPrefix(rendered, "**Planned infrastructure changes:** 2 to change, 1 to replace, 1 to destroy, 1 unchanged."), "unexpected summary in %q", rendered)
	for _, expected := range []string{
		`dgr0["aws_instance.a<br/>~ change (2 attributes)"]:::dgChange`,
		`dgr1["aws_instance.b<br/>~ change"]:::dgChange`,
		`dgr2["aws_instance.c<br/>± replace (1 attribute)"]:::dgReplace`,
		`dgr3["aws_instance.d<br/>- destroy"]:::dgDestroy`,
		`dgr4["data.aws_ami.e<br/>read"]:::dgUnchanged`,
	} {
		Assert(t, strings.Contains(rendered, expected), "expected %q in\n%s", expected, rendered)
	}
}

func TestRenderPlanDiagramMarkdown_ModuleSubgraph(t *testing.T) {
	diagram := &models.PlanDiagram{
		Nodes: []models.DiagramNode{
			{Address: "aws_vpc.main", Action: models.DiagramActionCreate},
			{Address: "module.app.aws_instance.web", Module: "module.app", Action: models.DiagramActionUpdate, ChangedAttributes: 1},
		},
		Edges: []models.DiagramEdge{{From: "aws_vpc.main", To: "module.app.aws_instance.web"}},
	}

	rendered := events.RenderPlanDiagramMarkdown(diagram)

	Assert(t, strings.Contains(rendered, `  subgraph dgm1["module.app"]`), "expected a subgraph in\n%s", rendered)
	Assert(t, strings.Contains(rendered, `    dgr1["module.app.aws_instance.web<br/>~ change (1 attribute)"]:::dgChange`), "expected the node inside the subgraph in\n%s", rendered)
	Assert(t, strings.Contains(rendered, "  dgr0 --> dgr1"), "expected the edge in\n%s", rendered)
}

func TestRenderPlanDiagramMarkdown_NoChanges(t *testing.T) {
	diagram := &models.PlanDiagram{
		Nodes: []models.DiagramNode{
			{Address: "aws_vpc.main", Action: models.DiagramActionNoOp},
			{Address: "data.aws_ami.ubuntu", Action: models.DiagramActionRead},
		},
	}

	Equals(t, "", events.RenderPlanDiagramMarkdown(diagram))
}

func TestRenderPlanDiagramMarkdown_NilDiagram(t *testing.T) {
	Equals(t, "", events.RenderPlanDiagramMarkdown(nil))
	Equals(t, "", events.RenderPlanDiagramMarkdown(&models.PlanDiagram{}))
}

func TestRenderPlanDiagramMarkdown_EscapesLabels(t *testing.T) {
	diagram := &models.PlanDiagram{
		Nodes: []models.DiagramNode{
			{Address: `aws_instance.web["<prod> & co"]`, Action: models.DiagramActionCreate},
		},
	}

	rendered := events.RenderPlanDiagramMarkdown(diagram)

	Assert(t, strings.Contains(rendered, `dgr0["aws_instance.web[#quot;#lt;prod#gt; #38; co#quot;]<br/>+ add"]:::dgAdd`), "unexpected escaping in\n%s", rendered)
	Assert(t, !strings.Contains(rendered, "<prod>"), "raw angle brackets leaked into the label:\n%s", rendered)
}

func TestRenderPlanDiagramMarkdown_Deterministic(t *testing.T) {
	first := events.RenderPlanDiagramMarkdown(parseTestDiagram(t))
	second := events.RenderPlanDiagramMarkdown(parseTestDiagram(t))

	Equals(t, first, second)
	Assert(t, strings.Contains(first, "```mermaid"), "expected a mermaid block in\n%s", first)
}

func TestRenderPlanDiagramMarkdown_BoundsNodes(t *testing.T) {
	var nodes []models.DiagramNode
	for i := range 100 {
		nodes = append(nodes, models.DiagramNode{
			Address: fmt.Sprintf("aws_instance.r%03d", i),
			Action:  models.DiagramActionCreate,
		})
	}

	rendered := events.RenderPlanDiagramMarkdown(&models.PlanDiagram{Nodes: nodes})

	Equals(t, 60, strings.Count(rendered, ":::dgAdd"))
	Assert(t, strings.Contains(rendered, "**Planned infrastructure changes:** 100 to add."), "unexpected summary in\n%s", rendered)
	Assert(t, strings.Contains(rendered, "_Showing 60 of 100 resources; 40 omitted._"), "expected the truncation note in\n%s", rendered)
}

func TestRenderPlanDiagramMarkdown_KeepsChangedResourcesAndNeighbours(t *testing.T) {
	var nodes []models.DiagramNode
	var edges []models.DiagramEdge
	// One changed resource depends on 99 unchanged ones: the changed resource
	// and as many of its neighbours as the budget allows must be drawn.
	nodes = append(nodes, models.DiagramNode{Address: "aws_instance.changed", Action: models.DiagramActionUpdate})
	for i := range 99 {
		address := fmt.Sprintf("aws_s3_bucket.b%03d", i)
		nodes = append(nodes, models.DiagramNode{Address: address, Action: models.DiagramActionNoOp})
		edges = append(edges, models.DiagramEdge{From: address, To: "aws_instance.changed"})
	}

	rendered := events.RenderPlanDiagramMarkdown(&models.PlanDiagram{Nodes: nodes, Edges: edges})

	Assert(t, strings.Contains(rendered, `aws_instance.changed<br/>~ change`), "expected the changed resource in\n%s", rendered)
	Equals(t, 60, strings.Count(rendered, ":::dg"))
	Assert(t, strings.Contains(rendered, "_Showing 60 of 100 resources; 40 omitted._"), "expected the truncation note in\n%s", rendered)
}

func TestRenderPlanDiagramMarkdown_BoundsEdges(t *testing.T) {
	var nodes []models.DiagramNode
	nodes = append(nodes, models.DiagramNode{Address: "aws_instance.first", Action: models.DiagramActionCreate})
	for i := range 50 {
		address := fmt.Sprintf("aws_s3_bucket.b%03d", i)
		nodes = append(nodes, models.DiagramNode{Address: address, Action: models.DiagramActionCreate})
	}
	var edges []models.DiagramEdge
	for i := range 50 {
		for range 20 {
			edges = append(edges, models.DiagramEdge{From: "aws_instance.first", To: fmt.Sprintf("aws_s3_bucket.b%03d", i)})
		}
	}

	rendered := events.RenderPlanDiagramMarkdown(&models.PlanDiagram{Nodes: nodes, Edges: edges})

	Equals(t, 300, strings.Count(rendered, " --> "))
}

func TestRenderPlanDiagramMarkdown_Forget(t *testing.T) {
	diagram := &models.PlanDiagram{
		Nodes: []models.DiagramNode{
			{Address: "aws_s3_bucket.old", Action: models.DiagramActionForget},
		},
	}

	rendered := events.RenderPlanDiagramMarkdown(diagram)

	Assert(t, strings.Contains(rendered, "**Planned infrastructure changes:** 1 to forget."), "unexpected summary in\n%s", rendered)
	Assert(t, strings.Contains(rendered, `aws_s3_bucket.old<br/>- forget"]:::dgDestroy`), "expected the forget node in\n%s", rendered)
}
