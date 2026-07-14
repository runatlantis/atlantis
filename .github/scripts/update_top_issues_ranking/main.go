// Copyright 2026 The Atlantis Authors
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"bytes"
	"context"
	_ "embed"
	"errors"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"slices"
	"strings"
	"text/template"
	"time"

	"github.com/google/go-github/v88/github"
)

//go:embed ranking.tmpl
var rankingTemplate string

const (
	defaultMinThumbsUp    = 10
	defaultMaxPerCategory = 15
	markdownBacktick      = "\x60"
)

var (
	defaultCategories = []categoryConfig{
		{Name: "Top Feature Requests", Labels: []string{"feature"}},
		{Name: "Top Bugs", Labels: []string{"bug"}},
		{Name: "Top Regressions", Labels: []string{"regression"}},
		{Name: "Top RFC / Needs Discussion", Labels: []string{"rfc", "needs discussion"}},
		{Name: "Top Contributor-Friendly Issues", Labels: []string{"help wanted", "quick-win"}},
	}

	defaultExcludedLabels = []string{
		"Stale",
		"duplicate",
		"wont-do",
		"spam",
		"waiting-on-response",
	}
)

type options struct {
	Owner          string
	Repo           string
	IssueNumber    int
	MinThumbsUp    int
	MaxPerCategory int
	DryRun         bool
}

type categoryConfig struct {
	Name   string
	Labels []string
}

type rankedIssue struct {
	Number        int
	Title         string
	URL           string
	Labels        []string
	ThumbsUp      int
	Comments      int
	UpdatedAt     time.Time
	State         string
	IsPullRequest bool
}

type rankingConfig struct {
	Categories          []categoryConfig
	ExcludedLabels      []string
	TrackingIssueNumber int
	MinThumbsUp         int
	MaxPerCategory      int
	GeneratedAt         time.Time
}

type rankingData struct {
	GeneratedAt    string
	MinThumbsUp    int
	ExcludedLabels string
	Categories     []categoryData
}

type categoryData struct {
	Name   string
	Issues []issueRow
}

type issueRow struct {
	Rank     int
	Issue    string
	ThumbsUp int
	Comments int
	Updated  string
	Labels   string
}

func main() {
	opts, err := parseOptions(os.Args[1:])
	if err != nil {
		log.Fatal(err)
	}

	if err := run(context.Background(), opts, os.Getenv("GITHUB_TOKEN"), os.Stdout); err != nil {
		log.Fatal(err)
	}
}

func parseOptions(args []string) (options, error) {
	var opts options
	fs := flag.NewFlagSet("update-top-issues-ranking", flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	fs.StringVar(&opts.Owner, "owner", "", "GitHub repository owner")
	fs.StringVar(&opts.Repo, "repo", "", "GitHub repository name")
	fs.IntVar(&opts.IssueNumber, "issue-number", 0, "tracking issue number to update")
	fs.IntVar(&opts.MinThumbsUp, "min-thumbs-up", defaultMinThumbsUp, "minimum thumbs-up reactions required")
	fs.IntVar(&opts.MaxPerCategory, "max-per-category", defaultMaxPerCategory, "maximum issues to render per category")
	fs.BoolVar(&opts.DryRun, "dry-run", false, "print generated Markdown instead of updating GitHub")

	if err := fs.Parse(args); err != nil {
		return options{}, err
	}
	if opts.Owner == "" {
		return options{}, errors.New("--owner is required")
	}
	if opts.Repo == "" {
		return options{}, errors.New("--repo is required")
	}
	if opts.MinThumbsUp < 0 {
		return options{}, errors.New("--min-thumbs-up must be greater than or equal to 0")
	}
	if opts.MaxPerCategory <= 0 {
		return options{}, errors.New("--max-per-category must be greater than 0")
	}
	if !opts.DryRun && opts.IssueNumber <= 0 {
		return options{}, errors.New("--issue-number must be set to the tracking issue number in write mode")
	}
	return opts, nil
}

func run(ctx context.Context, opts options, token string, out io.Writer) error {
	if !opts.DryRun && token == "" {
		return errors.New("GITHUB_TOKEN is required in write mode")
	}

	clientOptions := []github.ClientOptionsFunc{}
	if token != "" {
		clientOptions = append(clientOptions, github.WithAuthToken(token))
	}
	client, err := github.NewClient(clientOptions...)
	if err != nil {
		return fmt.Errorf("creating GitHub client: %w", err)
	}

	issues, err := listOpenIssues(ctx, client, opts.Owner, opts.Repo)
	if err != nil {
		return err
	}

	data := buildRankingData(issues, rankingConfig{
		Categories:          defaultCategories,
		ExcludedLabels:      defaultExcludedLabels,
		TrackingIssueNumber: opts.IssueNumber,
		MinThumbsUp:         opts.MinThumbsUp,
		MaxPerCategory:      opts.MaxPerCategory,
		GeneratedAt:         time.Now().UTC(),
	})

	body, err := renderRanking(data)
	if err != nil {
		return err
	}

	if opts.DryRun {
		_, err := fmt.Fprint(out, body)
		return err
	}

	_, _, err = client.Issues.Edit(ctx, opts.Owner, opts.Repo, opts.IssueNumber, &github.IssueRequest{
		Body: github.Ptr(body),
	})
	if err != nil {
		return fmt.Errorf("updating issue #%d: %w", opts.IssueNumber, err)
	}
	return nil
}

func listOpenIssues(ctx context.Context, client *github.Client, owner string, repo string) ([]rankedIssue, error) {
	opts := &github.IssueListByRepoOptions{
		State: "open",
		ListOptions: github.ListOptions{
			PerPage: 100,
		},
	}

	var issues []rankedIssue
	for {
		page, resp, err := client.Issues.ListByRepo(ctx, owner, repo, opts)
		if err != nil {
			return nil, fmt.Errorf("listing open issues: %w", err)
		}
		for _, issue := range page {
			issues = append(issues, issueFromGitHub(issue))
		}
		if resp == nil || resp.NextPage == 0 {
			break
		}
		opts.ListOptions.Page = resp.NextPage
	}

	return issues, nil
}

func issueFromGitHub(issue *github.Issue) rankedIssue {
	labels := make([]string, 0, len(issue.Labels))
	for _, label := range issue.Labels {
		labels = append(labels, label.GetName())
	}
	slices.SortFunc(labels, func(a string, b string) int {
		return strings.Compare(strings.ToLower(a), strings.ToLower(b))
	})

	return rankedIssue{
		Number:        issue.GetNumber(),
		Title:         issue.GetTitle(),
		URL:           issue.GetHTMLURL(),
		Labels:        labels,
		ThumbsUp:      issue.GetReactions().GetPlusOne(),
		Comments:      issue.GetComments(),
		UpdatedAt:     issue.GetUpdatedAt().Time,
		State:         issue.GetState(),
		IsPullRequest: issue.IsPullRequest(),
	}
}

func buildRankingData(issues []rankedIssue, cfg rankingConfig) rankingData {
	filtered := make([]rankedIssue, 0, len(issues))
	for _, issue := range issues {
		if includeIssue(issue, cfg) {
			filtered = append(filtered, issue)
		}
	}

	data := rankingData{
		GeneratedAt:    cfg.GeneratedAt.UTC().Format("2006-01-02 15:04:05 UTC"),
		MinThumbsUp:    cfg.MinThumbsUp,
		ExcludedLabels: formatInlineCodeList(cfg.ExcludedLabels),
		Categories:     make([]categoryData, 0, len(cfg.Categories)),
	}

	for _, category := range cfg.Categories {
		categoryIssues := make([]rankedIssue, 0, len(filtered))
		for _, issue := range filtered {
			if hasAnyLabel(issue, category.Labels) {
				categoryIssues = append(categoryIssues, issue)
			}
		}
		sortIssues(categoryIssues)
		if len(categoryIssues) > cfg.MaxPerCategory {
			categoryIssues = categoryIssues[:cfg.MaxPerCategory]
		}

		categoryData := categoryData{Name: category.Name}
		for i, issue := range categoryIssues {
			categoryData.Issues = append(categoryData.Issues, issueRow{
				Rank:     i + 1,
				Issue:    issueMarkdownLink(issue),
				ThumbsUp: issue.ThumbsUp,
				Comments: issue.Comments,
				Updated:  issue.UpdatedAt.UTC().Format("2006-01-02"),
				Labels:   formatLabelList(issue.Labels),
			})
		}
		data.Categories = append(data.Categories, categoryData)
	}

	return data
}

func includeIssue(issue rankedIssue, cfg rankingConfig) bool {
	if issue.IsPullRequest {
		return false
	}
	if !strings.EqualFold(issue.State, "open") {
		return false
	}
	if cfg.TrackingIssueNumber > 0 && issue.Number == cfg.TrackingIssueNumber {
		return false
	}
	if issue.ThumbsUp < cfg.MinThumbsUp {
		return false
	}
	return !hasAnyLabel(issue, cfg.ExcludedLabels)
}

func sortIssues(issues []rankedIssue) {
	slices.SortFunc(issues, func(a rankedIssue, b rankedIssue) int {
		if a.ThumbsUp != b.ThumbsUp {
			return b.ThumbsUp - a.ThumbsUp
		}
		if a.Comments != b.Comments {
			return b.Comments - a.Comments
		}
		if !a.UpdatedAt.Equal(b.UpdatedAt) {
			if a.UpdatedAt.After(b.UpdatedAt) {
				return -1
			}
			return 1
		}
		return a.Number - b.Number
	})
}

func renderRanking(data rankingData) (string, error) {
	tmpl, err := template.New("ranking").Parse(rankingTemplate)
	if err != nil {
		return "", err
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", err
	}
	return buf.String(), nil
}

func hasAnyLabel(issue rankedIssue, labels []string) bool {
	issueLabels := make(map[string]struct{}, len(issue.Labels))
	for _, label := range issue.Labels {
		issueLabels[strings.ToLower(label)] = struct{}{}
	}
	for _, label := range labels {
		if _, ok := issueLabels[strings.ToLower(label)]; ok {
			return true
		}
	}
	return false
}

func issueMarkdownLink(issue rankedIssue) string {
	return fmt.Sprintf("[#%d %s](%s)", issue.Number, escapeMarkdownTableText(issue.Title), issue.URL)
}

func formatLabelList(labels []string) string {
	if len(labels) == 0 {
		return ""
	}
	escaped := make([]string, 0, len(labels))
	for _, label := range labels {
		escaped = append(escaped, markdownBacktick+escapeMarkdownTableText(label)+markdownBacktick)
	}
	return strings.Join(escaped, ", ")
}

func formatInlineCodeList(values []string) string {
	escaped := make([]string, 0, len(values))
	for _, value := range values {
		escaped = append(escaped, markdownBacktick+escapeMarkdownTableText(value)+markdownBacktick)
	}
	return strings.Join(escaped, ", ")
}

func escapeMarkdownTableText(s string) string {
	replacer := strings.NewReplacer(
		"\\", "\\\\",
		"|", "\\|",
		"\r", " ",
		"\n", " ",
		"[", "\\[",
		"]", "\\]",
	)
	return strings.Join(strings.Fields(replacer.Replace(s)), " ")
}
