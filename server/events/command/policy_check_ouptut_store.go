package command

import (
	"fmt"

	"github.com/runatlantis/atlantis/server/events/models"
)

const KeySeparator = "||"

type PolicyCheckOutputStore struct {
	store map[string]*models.PolicyCheckSuccess
}

func NewPolicyCheckOutputStore() *PolicyCheckOutputStore {
	return &PolicyCheckOutputStore{
		store: map[string]*models.PolicyCheckSuccess{},
	}
}

func buildKey(projectName string, workspace string) string {
	return fmt.Sprintf("%s%s%s", projectName, KeySeparator, workspace)
}

func (p *PolicyCheckOutputStore) Get(projectName string, workspace string) *models.PolicyCheckSuccess {
	key := buildKey(projectName, workspace)

	if output, ok := p.store[key]; ok {
		return output
	}
	return nil
}

func (p *PolicyCheckOutputStore) Set(projectName string, workspace string, output string) {
	key := buildKey(projectName, workspace)
	p.store[key] = &models.PolicyCheckSuccess{
		PolicyCheckOutput: output,
	}
}
