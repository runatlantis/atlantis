package webhooks_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/runatlantis/atlantis/server/events/models"
	"github.com/runatlantis/atlantis/server/events/webhooks"
	. "github.com/runatlantis/atlantis/testing"
)

func TestDefaultMSTeamsClient_PostMessage(t *testing.T) {
	t.Run("should send correct message format", func(t *testing.T) {
		var receivedMessage webhooks.TeamsMessage

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Method != "POST" {
				t.Errorf("Expected POST request, got %s", r.Method)
			}

			if r.Header.Get("Content-Type") != "application/json" {
				t.Errorf("Expected Content-Type application/json, got %s", r.Header.Get("Content-Type"))
			}

			err := json.NewDecoder(r.Body).Decode(&receivedMessage)
			if err != nil {
				t.Errorf("Failed to decode message: %v", err)
			}

			w.WriteHeader(http.StatusOK)
		}))
		defer server.Close()

		client := webhooks.NewMSTeamsClient()

		applyResult := webhooks.ApplyResult{
			Workspace: "production",
			Repo: models.Repo{
				FullName: "owner/repo",
			},
			Pull: models.PullRequest{
				BaseBranch: "main",
				Num:        1,
				URL:        "https://github.com/owner/repo/pull/1",
			},
			User: models.User{
				Username: "atlantis",
			},
			Success:     true,
			Directory:   ".",
			ProjectName: "test-project",
		}

		err := client.PostMessage(server.URL, applyResult)
		Ok(t, err)

		// Verify message structure
		Equals(t, "message", receivedMessage.Type)
		Assert(t, len(receivedMessage.Attachments) == 1, "Expected 1 attachment")

		attachment := receivedMessage.Attachments[0]
		Equals(t, "application/vnd.microsoft.card.adaptive", attachment.ContentType)

		card := attachment.Content
		Equals(t, "AdaptiveCard", card.Type)
		Equals(t, "http://adaptivecards.io/schemas/adaptive-card.json", card.Schema)
		Equals(t, "1.4", card.Version)

		// Verify body elements
		Assert(t, len(card.Body) > 0, "Expected card body elements")

		// Check title element
		titleElement := card.Body[0]
		Equals(t, "TextBlock", titleElement.Type)
		Equals(t, "Atlantis Apply succeeded", titleElement.Text)
		Equals(t, "Good", titleElement.Color)

		// Check subtitle element
		subtitleElement := card.Body[1]
		Equals(t, "TextBlock", subtitleElement.Type)
		Equals(t, "Repository: owner/repo", subtitleElement.Text)
	})

	t.Run("should handle failure case", func(t *testing.T) {
		var receivedMessage webhooks.TeamsMessage

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			json.NewDecoder(r.Body).Decode(&receivedMessage)
			w.WriteHeader(http.StatusOK)
		}))
		defer server.Close()

		client := webhooks.NewMSTeamsClient()

		applyResult := webhooks.ApplyResult{
			Workspace: "production",
			Repo: models.Repo{
				FullName: "owner/repo",
			},
			Pull: models.PullRequest{
				BaseBranch: "main",
				Num:        1,
				URL:        "https://github.com/owner/repo/pull/1",
			},
			User: models.User{
				Username: "atlantis",
			},
			Success:   false,
			Directory: "terraform/",
		}

		err := client.PostMessage(server.URL, applyResult)
		Ok(t, err)

		// Verify failure message
		Assert(t, len(receivedMessage.Attachments) == 1, "Expected 1 attachment")
		card := receivedMessage.Attachments[0].Content

		// Check title element for failure
		titleElement := card.Body[0]
		Equals(t, "Atlantis Apply failed", titleElement.Text)
		Equals(t, "Attention", titleElement.Color)
	})

	t.Run("should handle server error", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("Internal Server Error"))
		}))
		defer server.Close()

		client := webhooks.NewMSTeamsClient()

		applyResult := webhooks.ApplyResult{
			Workspace: "production",
			Success:   true,
		}

		err := client.PostMessage(server.URL, applyResult)
		Assert(t, err != nil, "Expected error for server error response")
	})
}
