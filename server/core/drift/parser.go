// Copyright 2025 The Atlantis Authors
// SPDX-License-Identifier: Apache-2.0

// Package drift provides drift detection utilities for Atlantis.
// It enables detecting infrastructure drift outside of PR workflows.
package drift

import (
	"github.com/runatlantis/atlantis/server/events/models"
)

// Parser extracts drift information from Terraform plan output.
type Parser struct{}

// NewParser creates a new drift parser.
func NewParser() *Parser {
	return &Parser{}
}

// ParsePlanOutput extracts drift information from a PlanSuccess result.
// This leverages the existing plan output parsing infrastructure in models.PlanSuccess.
func (p *Parser) ParsePlanOutput(plan *models.PlanSuccess) models.DriftSummary {
	return models.NewDriftSummaryFromPlanSuccess(plan)
}

// ParsePlanStats creates a drift summary from plan statistics.
// This is useful when you have pre-computed stats.
func (p *Parser) ParsePlanStats(stats models.PlanSuccessStats, summary string) models.DriftSummary {
	return models.NewDriftSummaryFromPlanStats(stats, summary)
}

// HasDrift returns true if the plan output indicates infrastructure drift.
func (p *Parser) HasDrift(plan *models.PlanSuccess) bool {
	if plan == nil {
		return false
	}
	stats := plan.Stats()
	return stats.Add > 0 || stats.Change > 0 || stats.Destroy > 0 || stats.Import > 0
}
