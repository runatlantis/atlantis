package terraform

import (
	"encoding/json"

	"github.com/hashicorp/terraform-json"
	"github.com/pkg/errors"
)

type ResourceSummary struct {
	Address string
}

type PlanSummary struct {
	Creations []ResourceSummary
	Deletions []ResourceSummary
	Updates   []ResourceSummary
}

func (s PlanSummary) IsEmpty() bool {
	return len(s.Creations) == 0 && len(s.Deletions) == 0 && len(s.Updates) == 0
}

// Generates a super simple plan summary with changes grouped by action
// creation, deletion, update.
// changes are only represented using addresses for now.
func NewPlanSummaryFromJSON(b []byte) (PlanSummary, error) {
	if len(b) == 0 {
		return PlanSummary{}, nil
	}
	var plan tfjson.Plan
	err := json.Unmarshal(b, &plan)

	if err != nil {
		return PlanSummary{}, errors.Wrap(err, "parsing plan json")
	}

	var creations []ResourceSummary
	var deletions []ResourceSummary
	var updates []ResourceSummary
	for _, c := range plan.ResourceChanges {
		summary := ResourceSummary{
			Address: c.Address,
		}
		actions := c.Change.Actions
		if actions.Delete() || actions.Replace() {
			deletions = append(deletions, summary)
		}

		if actions.Create() || actions.Replace() {
			creations = append(creations, summary)
		}

		if actions.Update() {
			updates = append(updates, summary)
		}
	}

	return PlanSummary{
		Creations: creations,
		Deletions: deletions,
		Updates:   updates,
	}, nil

}
