// Copyright 2026 The Atlantis Authors
// SPDX-License-Identifier: Apache-2.0

package models

// DiagramAction is the planned action for one resource in a plan diagram.
type DiagramAction string

const (
	// DiagramActionCreate creates a resource that does not exist yet.
	DiagramActionCreate DiagramAction = "create"
	// DiagramActionUpdate changes an existing resource in place.
	DiagramActionUpdate DiagramAction = "update"
	// DiagramActionReplace destroys and recreates a resource.
	DiagramActionReplace DiagramAction = "replace"
	// DiagramActionDelete destroys an existing resource.
	DiagramActionDelete DiagramAction = "delete"
	// DiagramActionForget removes a resource from state without destroying it.
	DiagramActionForget DiagramAction = "forget"
	// DiagramActionRead reads a data source during apply.
	DiagramActionRead DiagramAction = "read"
	// DiagramActionNoOp leaves an existing resource untouched.
	DiagramActionNoOp DiagramAction = "no-op"
)

// ChangesInfrastructure reports whether the action adds, changes or removes
// infrastructure. Reading a data source and doing nothing do not.
func (a DiagramAction) ChangesInfrastructure() bool {
	switch a {
	case DiagramActionCreate, DiagramActionUpdate, DiagramActionReplace, DiagramActionDelete, DiagramActionForget:
		return true
	}
	return false
}

// PlanDiagram is the resource dependency graph of a Terraform plan: the
// resources the plan touches (and, for context, the unchanged resources they
// depend on), the planned action for each, and the edges between them.
// Edges point from a resource to the resource that depends on it.
type PlanDiagram struct {
	// Nodes are sorted by address.
	Nodes []DiagramNode
	// Edges are sorted by From and then To, and never repeat.
	Edges []DiagramEdge
}

// DiagramNode is one resource in a plan diagram. Address is the Terraform
// resource address, including module prefix and instance key.
type DiagramNode struct {
	Address string
	// Module is the module address the resource belongs to, for example
	// "module.vpc". It is empty for the root module.
	Module string
	Action DiagramAction
	// ChangedAttributes is the number of top-level attributes the plan
	// changes. It is zero for create and delete, where every attribute is
	// implied.
	ChangedAttributes int
}

// DiagramEdge is a dependency between two resources, from the resource that is
// depended on to the one that refers to it.
type DiagramEdge struct {
	From string
	To   string
}

// Empty reports whether the diagram has nothing to draw.
func (d *PlanDiagram) Empty() bool {
	return d == nil || len(d.Nodes) == 0
}
