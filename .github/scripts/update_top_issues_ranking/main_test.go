// Copyright 2026 The Atlantis Authors
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"strconv"
	"strings"
	"testing"
	"time"
)

func TestBuildRankingDataFiltersIssues(t *testing.T) {
	base := testIssue(1, "feature", 20, labels("feature"))
	category := []categoryConfig{{Name: "Features", Labels: []string{"feature"}}}

	tests := []struct {
		name                string
		issue               rankedIssue
		trackingIssueNumber int
		wantIncluded        bool
	}{
		{
			name:         "includes matching issue",
			issue:        base,
			wantIncluded: true,
		},
		{
			name: "excludes pull requests",
			issue: testIssue(2, "pull request", 100, labels("feature"), func(issue *rankedIssue) {
				issue.IsPullRequest = true
			}),
		},
		{
			name:  "excludes configured labels case insensitively",
			issue: testIssue(3, "stale", 100, labels("feature", "stale")),
		},
		{
			name:  "excludes issues below threshold",
			issue: testIssue(4, "below threshold", 9, labels("feature")),
		},
		{
			name:                "excludes tracking issue",
			issue:               testIssue(5, "tracking issue", 100, labels("feature")),
			trackingIssueNumber: 5,
		},
		{
			name: "excludes closed issues",
			issue: testIssue(6, "closed", 100, labels("feature"), func(issue *rankedIssue) {
				issue.State = "closed"
			}),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data := buildRankingData([]rankedIssue{tt.issue}, testConfig(category, tt.trackingIssueNumber))
			got := data.Categories[0].Issues
			if tt.wantIncluded {
				if len(got) != 1 {
					t.Fatalf("expected issue to be included, got %#v", got)
				}
				if !strings.Contains(got[0].Issue, "#"+strconv.Itoa(tt.issue.Number)+" ") {
					t.Fatalf("expected issue #%d, got %q", tt.issue.Number, got[0].Issue)
				}
				return
			}
			if len(got) != 0 {
				t.Fatalf("expected issue to be excluded, got %#v", got)
			}
		})
	}
}

func TestBuildRankingDataGroupsByCategory(t *testing.T) {
	issues := []rankedIssue{
		testIssue(1, "feature", 20, labels("feature")),
		testIssue(2, "bug regression", 20, labels("bug", "regression")),
		testIssue(3, "discussion", 20, labels("needs discussion")),
		testIssue(4, "contributor", 20, labels("quick-win")),
	}

	data := buildRankingData(issues, testConfig(defaultCategories, 0))

	assertCategoryIssues(t, data, "Top Feature Requests", []int{1})
	assertCategoryIssues(t, data, "Top Bugs", []int{2})
	assertCategoryIssues(t, data, "Top Regressions", []int{2})
	assertCategoryIssues(t, data, "Top RFC / Needs Discussion", []int{3})
	assertCategoryIssues(t, data, "Top Contributor-Friendly Issues", []int{4})
}

func TestSortIssuesDeterministicTies(t *testing.T) {
	issues := []rankedIssue{
		testIssue(4, "lower comments", 10, labels("feature"), withComments(1), withUpdated("2024-01-01")),
		testIssue(3, "older", 10, labels("feature"), withComments(2), withUpdated("2023-01-01")),
		testIssue(2, "higher number", 10, labels("feature"), withComments(2), withUpdated("2024-01-01")),
		testIssue(1, "lowest number", 10, labels("feature"), withComments(2), withUpdated("2024-01-01")),
		testIssue(5, "most thumbs up", 11, labels("feature"), withComments(0), withUpdated("2020-01-01")),
	}

	data := buildRankingData(issues, testConfig([]categoryConfig{{Name: "Features", Labels: []string{"feature"}}}, 0))

	assertCategoryIssues(t, data, "Features", []int{5, 1, 2, 3, 4})
}

func TestBuildRankingDataLimitsRowsPerCategory(t *testing.T) {
	issues := []rankedIssue{
		testIssue(1, "one", 30, labels("feature")),
		testIssue(2, "two", 20, labels("feature")),
		testIssue(3, "three", 10, labels("feature")),
	}

	cfg := testConfig([]categoryConfig{{Name: "Features", Labels: []string{"feature"}}}, 0)
	cfg.MaxPerCategory = 2
	data := buildRankingData(issues, cfg)

	assertCategoryIssues(t, data, "Features", []int{1, 2})
}

func TestRenderRankingMarkdownTableAndEscaping(t *testing.T) {
	issue := testIssue(42, "Pipe | [brackets]\nnext", 12, labels("feature", "provider/github"), withComments(3), withUpdated("2026-06-29"))
	data := buildRankingData([]rankedIssue{issue}, testConfig([]categoryConfig{{Name: "Features", Labels: []string{"feature"}}}, 0))

	got, err := renderRanking(data)
	if err != nil {
		t.Fatalf("render ranking: %v", err)
	}

	expected := []string{
		"# Top-Ranking Atlantis Issues 📊",
		"_Last updated: 2026-06-29 12:00:00 UTC._",
		"| Rank | Issue | 👍 | Comments | Updated | Labels |",
		"| 1 | [#42 Pipe \\| \\[brackets\\] next](https://github.com/runatlantis/atlantis/issues/42) | 12 | 3 | 2026-06-29 | " + markdownBacktick + "feature" + markdownBacktick + ", " + markdownBacktick + "provider/github" + markdownBacktick + " |",
		"ranked by 👍 reactions on the original issue",
		"not a roadmap commitment, priority guarantee, or acceptance decision",
	}
	for _, want := range expected {
		if !strings.Contains(got, want) {
			t.Fatalf("expected rendered markdown to contain %q\n\n%s", want, got)
		}
	}
}

func TestRenderRankingEmptyCategory(t *testing.T) {
	data := buildRankingData(nil, testConfig([]categoryConfig{{Name: "Features", Labels: []string{"feature"}}}, 0))

	got, err := renderRanking(data)
	if err != nil {
		t.Fatalf("render ranking: %v", err)
	}
	if !strings.Contains(got, "## Features") {
		t.Fatalf("expected category heading, got:\n%s", got)
	}
	if !strings.Contains(got, "No matching issues above threshold.") {
		t.Fatalf("expected empty category message, got:\n%s", got)
	}
}

func testConfig(categories []categoryConfig, trackingIssueNumber int) rankingConfig {
	return rankingConfig{
		Categories:          categories,
		ExcludedLabels:      defaultExcludedLabels,
		TrackingIssueNumber: trackingIssueNumber,
		MinThumbsUp:         defaultMinThumbsUp,
		MaxPerCategory:      defaultMaxPerCategory,
		GeneratedAt:         time.Date(2026, 6, 29, 12, 0, 0, 0, time.UTC),
	}
}

func testIssue(number int, title string, thumbsUp int, issueLabels []string, opts ...func(*rankedIssue)) rankedIssue {
	issue := rankedIssue{
		Number:    number,
		Title:     title,
		URL:       "https://github.com/runatlantis/atlantis/issues/" + strconv.Itoa(number),
		Labels:    issueLabels,
		ThumbsUp:  thumbsUp,
		UpdatedAt: time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
		State:     "open",
	}
	for _, opt := range opts {
		opt(&issue)
	}
	return issue
}

func labels(values ...string) []string {
	return values
}

func withComments(comments int) func(*rankedIssue) {
	return func(issue *rankedIssue) {
		issue.Comments = comments
	}
}

func withUpdated(date string) func(*rankedIssue) {
	return func(issue *rankedIssue) {
		parsed, err := time.Parse("2006-01-02", date)
		if err != nil {
			panic(err)
		}
		issue.UpdatedAt = parsed
	}
}

func assertCategoryIssues(t *testing.T, data rankingData, categoryName string, want []int) {
	t.Helper()
	for _, category := range data.Categories {
		if category.Name != categoryName {
			continue
		}
		if len(category.Issues) != len(want) {
			t.Fatalf("%s: expected %d issues, got %d: %#v", categoryName, len(want), len(category.Issues), category.Issues)
		}
		for i, row := range category.Issues {
			if !strings.Contains(row.Issue, "#"+strconv.Itoa(want[i])+" ") {
				t.Fatalf("%s: row %d expected issue #%d, got %q", categoryName, i, want[i], row.Issue)
			}
		}
		return
	}
	t.Fatalf("category %q not found", categoryName)
}
