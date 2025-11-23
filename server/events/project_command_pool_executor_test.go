package events

import (
	"fmt"
	"testing"

	"github.com/runatlantis/atlantis/server/events/command"
	"github.com/runatlantis/atlantis/server/events/models"
)

func pint(i int) *int { return &i }

func TestAssignEffectiveExecutionGroups_NonDestroy(t *testing.T) {
	cmds := []command.ProjectContext{
		{ExecutionOrderGroup: 0},
		{ExecutionOrderGroup: 2, DestroyExecutionOrderGroup: pint(5)},
	}
	assignEffectiveExecutionGroups(cmds, false)
	// In non-destroy mode we use ExecutionOrderGroup for EffectiveExecutionOrderGroup.
	if cmds[0].EffectiveExecutionOrderGroup != 0 || cmds[1].EffectiveExecutionOrderGroup != 2 {
		t.Fatalf("expected EffectiveExecutionOrderGroup to match ExecutionOrderGroup when not destroy, got %d and %d",
			cmds[0].EffectiveExecutionOrderGroup, cmds[1].EffectiveExecutionOrderGroup)
	}
}

func TestAssignEffectiveExecutionGroups_Destroy_NoneDefined(t *testing.T) {
	cmds := []command.ProjectContext{
		{ExecutionOrderGroup: 1},
		{ExecutionOrderGroup: 3},
	}
	assignEffectiveExecutionGroups(cmds, true)
	if cmds[0].EffectiveExecutionOrderGroup != 1 || cmds[1].EffectiveExecutionOrderGroup != 3 {
		t.Fatalf("expected use of ExecutionOrderGroup when none defined")
	}
}

func TestAssignEffectiveExecutionGroups_Destroy_AllDefined(t *testing.T) {
	cmds := []command.ProjectContext{
		{ExecutionOrderGroup: 1, DestroyExecutionOrderGroup: pint(10)},
		{ExecutionOrderGroup: 3, DestroyExecutionOrderGroup: pint(11)},
	}
	assignEffectiveExecutionGroups(cmds, true)
	if cmds[0].EffectiveExecutionOrderGroup != 10 || cmds[1].EffectiveExecutionOrderGroup != 11 {
		t.Fatalf("expected use of all destroy groups when all defined")
	}
}

func TestAssignEffectiveExecutionGroups_Destroy_Partial(t *testing.T) {
	cmds := []command.ProjectContext{
		{ExecutionOrderGroup: 1, DestroyExecutionOrderGroup: pint(9)},
		{ExecutionOrderGroup: 3},
		{ExecutionOrderGroup: 2, DestroyExecutionOrderGroup: pint(7)},
	}
	assignEffectiveExecutionGroups(cmds, true)
	exp := []int{9, 3, 7}
	for i, c := range cmds {
		if c.EffectiveExecutionOrderGroup != exp[i] {
			t.Fatalf("index %d expected %d got %d", i, exp[i], c.EffectiveExecutionOrderGroup)
		}
	}
}

// Single test with subtests covering Destroy for Plan and Apply in Partial, AllDefined, and NoneDefined scenarios.
func TestRunProjectCmdsParallelGroups_Destroy_Scenarios(t *testing.T) {
	type scenario struct {
		name          string
		execGroups    []int
		destroyGroups []*int
		expectedOrder []int
	}
	scenarios := []scenario{
		{
			name:          "Partial",
			execGroups:    []int{0, 1, 2, 3},
			destroyGroups: []*int{pint(10), pint(1), pint(5), nil},
			// effective: 10,1,5,3 -> final order 1,3,5,10
			expectedOrder: []int{1, 3, 5, 10},
		},
		{
			name:          "AllDefined",
			execGroups:    []int{0, 1, 2},
			destroyGroups: []*int{pint(3), pint(1), pint(2)},
			// effective: 3,1,2 -> final order 1,2,3
			expectedOrder: []int{1, 2, 3},
		},
		{
			name:          "NoneDefined",
			execGroups:    []int{2, 0, 1},
			destroyGroups: []*int{nil, nil, nil},
			// effective: 2,0,1 -> final order 0,1,2
			expectedOrder: []int{0, 1, 2},
		},
	}

	commands := []command.Name{command.Plan, command.Apply}

	for _, sc := range scenarios {
		for _, cmdName := range commands {
			t.Run(fmt.Sprintf("%s_%s", sc.name, cmdName.String()), func(t *testing.T) {
				var order []int
				// Build contexts
				var contexts []command.ProjectContext
				for i := range sc.execGroups {
					contexts = append(contexts, command.ProjectContext{
						ExecutionOrderGroup:        sc.execGroups[i],
						DestroyExecutionOrderGroup: sc.destroyGroups[i],
						IsDestroy:                  true,
					})
				}
				runner := func(ctx command.ProjectContext) command.ProjectResult {
					order = append(order, ctx.EffectiveExecutionOrderGroup)
					if cmdName == command.Apply {
						return command.ProjectResult{ApplySuccess: "ok"}
					}
					return command.ProjectResult{PlanSuccess: &models.PlanSuccess{TerraformOutput: "No changes. Infrastructure is up-to-date"}}
				}
				res := runProjectCmdsParallelGroups(&command.Context{}, contexts, runner, 1)
				if res.HasErrors() {
					t.Fatalf("unexpected errors in scenario %s/%s", sc.name, cmdName.String())
				}
				if len(order) != len(sc.expectedOrder) {
					t.Fatalf("expected %d executions got %d", len(sc.expectedOrder), len(order))
				}
				for i := range sc.expectedOrder {
					if order[i] != sc.expectedOrder[i] {
						t.Fatalf("position %d expected group %d got %d; full=%v", i, sc.expectedOrder[i], order[i], order)
					}
				}
			})
		}
	}
}
