// Copyright 2017 HootSuite Media Inc.
//
// Licensed under the Apache License, Version 2.0 (the License);
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//    http://www.apache.org/licenses/LICENSE-2.0
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an AS IS BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
// Modified hereafter by contributors to runatlantis/atlantis.

package webhooks

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/pkg/errors"
)

const (
	teamsSuccessColor = "good"
	teamsFailureColor = "attention"
)

//go:generate pegomock generate --package mocks -o mocks/mock_msteams_client.go MSTeamsClient

// MSTeamsClient handles making API calls to Microsoft Teams.
type MSTeamsClient interface {
	PostMessage(webhookURL string, applyResult ApplyResult) error
}

// DefaultMSTeamsClient is the default implementation of MSTeamsClient.
type DefaultMSTeamsClient struct {
	Client *http.Client
}

// NewMSTeamsClient creates a new MS Teams client.
func NewMSTeamsClient() MSTeamsClient {
	return &DefaultMSTeamsClient{
		Client: http.DefaultClient,
	}
}

// PostMessage sends a message to MS Teams using the webhook URL.
func (d *DefaultMSTeamsClient) PostMessage(webhookURL string, applyResult ApplyResult) error {
	message := d.createMessage(applyResult)

	jsonData, err := json.Marshal(message)
	if err != nil {
		return errors.Wrap(err, "marshaling teams message")
	}

	req, err := http.NewRequest("POST", webhookURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return errors.Wrap(err, "creating request")
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := d.Client.Do(req)
	if err != nil {
		return errors.Wrap(err, "sending request")
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("teams webhook returned status %d: %s", resp.StatusCode, string(body))
	}

	return nil
}

// TeamsMessage represents the structure of a Microsoft Teams Adaptive Card message.
type TeamsMessage struct {
	Type        string                `json:"type"`
	Attachments []TeamsCardAttachment `json:"attachments"`
}

// TeamsCardAttachment represents an attachment in a Teams message.
type TeamsCardAttachment struct {
	ContentType string            `json:"contentType"`
	Content     TeamsAdaptiveCard `json:"content"`
}

// TeamsAdaptiveCard represents the Adaptive Card structure.
type TeamsAdaptiveCard struct {
	Type    string                  `json:"type"`
	Schema  string                  `json:"$schema"`
	Version string                  `json:"version"`
	Body    []TeamsAdaptiveCardBody `json:"body"`
}

// TeamsAdaptiveCardBody represents elements in the card body.
type TeamsAdaptiveCardBody struct {
	Type    string                    `json:"type"`
	Text    string                    `json:"text,omitempty"`
	Weight  string                    `json:"weight,omitempty"`
	Size    string                    `json:"size,omitempty"`
	Color   string                    `json:"color,omitempty"`
	Columns []TeamsAdaptiveCardColumn `json:"columns,omitempty"`
}

// TeamsAdaptiveCardColumn represents a column in a column set.
type TeamsAdaptiveCardColumn struct {
	Type  string                  `json:"type"`
	Width string                  `json:"width,omitempty"`
	Items []TeamsAdaptiveCardBody `json:"items,omitempty"`
}

// createMessage creates a Teams message from the apply result.
func (d *DefaultMSTeamsClient) createMessage(applyResult ApplyResult) TeamsMessage {
	var color string
	var successWord string

	if applyResult.Success {
		color = "Good"
		successWord = "succeeded"
	} else {
		color = "Attention"
		successWord = "failed"
	}

	title := fmt.Sprintf("Atlantis Apply %s", successWord)
	subtitle := fmt.Sprintf("Repository: %s", applyResult.Repo.FullName)

	directory := applyResult.Directory
	// Since "." looks weird, replace it with "/" to make it clear this is the root.
	if directory == "." {
		directory = "/"
	}

	// Create card body elements
	var bodyElements []TeamsAdaptiveCardBody

	// Title
	bodyElements = append(bodyElements, TeamsAdaptiveCardBody{
		Type:   "TextBlock",
		Text:   title,
		Weight: "Bolder",
		Size:   "Medium",
		Color:  color,
	})

	// Subtitle
	bodyElements = append(bodyElements, TeamsAdaptiveCardBody{
		Type: "TextBlock",
		Text: subtitle,
		Size: "Small",
	})

	// Facts as column sets
	facts := []struct{ name, value string }{
		{"Workspace", applyResult.Workspace},
		{"Branch", applyResult.Pull.BaseBranch},
		{"User", applyResult.User.Username},
		{"Directory", directory},
		{"Pull Request", fmt.Sprintf("[#%d](%s)", applyResult.Pull.Num, applyResult.Pull.URL)},
	}

	if applyResult.ProjectName != "" {
		facts = append(facts, struct{ name, value string }{"Project", applyResult.ProjectName})
	}

	for _, fact := range facts {
		bodyElements = append(bodyElements, TeamsAdaptiveCardBody{
			Type: "ColumnSet",
			Columns: []TeamsAdaptiveCardColumn{
				{
					Type:  "Column",
					Width: "auto",
					Items: []TeamsAdaptiveCardBody{
						{
							Type:   "TextBlock",
							Text:   fmt.Sprintf("**%s:**", fact.name),
							Weight: "Bolder",
						},
					},
				},
				{
					Type:  "Column",
					Width: "stretch",
					Items: []TeamsAdaptiveCardBody{
						{
							Type: "TextBlock",
							Text: fact.value,
						},
					},
				},
			},
		})
	}

	return TeamsMessage{
		Type: "message",
		Attachments: []TeamsCardAttachment{
			{
				ContentType: "application/vnd.microsoft.card.adaptive",
				Content: TeamsAdaptiveCard{
					Type:    "AdaptiveCard",
					Schema:  "http://adaptivecards.io/schemas/adaptive-card.json",
					Version: "1.4",
					Body:    bodyElements,
				},
			},
		},
	}
}
