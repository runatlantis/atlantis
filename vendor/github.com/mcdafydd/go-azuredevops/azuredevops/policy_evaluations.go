package azuredevops

import (
	"context"
	"fmt"
	"net/http"
)

// PolicyEvaluationsService handles communication with the evaluations methods on the API
// utilising https://docs.microsoft.com/en-us/rest/api/azure/devops/policy/evaluations
type PolicyEvaluationsService struct {
	client *Client
}

// PolicyEvaluationsListOptions describes what the request to the API should look like
type PolicyEvaluationsListOptions struct{}

// PolicyEvaluationsListResponse describes a pull requests list response
type PolicyEvaluationsListResponse struct {
	Count             int                       `json:"count"`
	PolicyEvaluations []*PolicyEvaluationRecord `json:"value"`
}

// PolicyEvaluationRecord encapsulates the current state of a policy as it applies to one specific pull request.
type PolicyEvaluationRecord struct {
	Links         *map[string]Link     `json:"_links,omitempty"`
	ArtifactID    *string              `json:"artifactId,omitempty"`
	CompletedDate *string              `json:"completedDate,omitempty"`
	Configuration *PolicyConfiguration `json:"configuration,omitempty"`
	Context       interface{}          `json:"context,omitempty"`
	EvaluationID  *string              `json:"evaluationId,omitempty"`
	StartedDate   *string              `json:"startedDate,omitempty"`
	Status        *string              `json:"status,omitempty"`
}

// PolicyConfiguration is the full policy configuration with settings.
type PolicyConfiguration struct {
	Links       interface{}    `json:"_links,omitempty"`
	CreatedBy   *IdentityRef   `json:"createdBy,omitempty"`
	CreatedDate *string        `json:"createdDate,omitempty"`
	ID          *int           `json:"id,omitempty"`
	IsBlocking  *bool          `json:"isBlocking,omitempty"`
	IsDeleted   *bool          `json:"isDeleted,omitempty"`
	IsEnabled   *bool          `json:"isEnabled,omitempty"`
	Revision    *int           `json:"revision,omitempty"`
	Settings    interface{}    `json:"settings,omitempty"`
	Type        *PolicyTypeRef `json:"type,omitempty"`
	Url         *string        `json:"url,omitempty"`
}

// PolicyTypeRef is the policy type reference.
type PolicyTypeRef struct {
	DisplayName *string `json:"displayName,omitempty"`
	ID          *string `json:"id,omitempty"`
	Url         *string `json:"url,omitempty"`
}

// GetPullRequestArtifactID gets the Artifact ID of a pull request.
// ex: vstfs:///CodeReview/CodeReviewId/{projectId}/{pullRequestId}
func (s *PolicyEvaluationsService) GetPullRequestArtifactID(projectID string, pullRequestID string) string {
	return fmt.Sprintf("vstfs:///CodeReview/CodeReviewId/%s/%s", projectID, pullRequestID)
}

// List retrieves a list of all the policy evaluation statuses for a specific pull request.
// https://docs.microsoft.com/en-us/rest/api/azure/devops/policy/evaluations/list?view=azure-devops-rest-5.1
func (s *PolicyEvaluationsService) List(ctx context.Context, owner, project, artifactID string, opts *PolicyEvaluationsListOptions) ([]*PolicyEvaluationRecord, *http.Response, error) {
	URL := fmt.Sprintf("%s/%s/_apis/policy/evaluations?artifactId=%s&api-version=5.1-preview",
		owner,
		project,
		artifactID,
	)
	URL, err := addOptions(URL, opts)

	req, err := s.client.NewRequest("GET", URL, nil)
	if err != nil {
		return nil, nil, err
	}

	r := new(PolicyEvaluationsListResponse)
	resp, err := s.client.Execute(ctx, req, r)
	if err != nil {
		return nil, nil, err
	}

	return r.PolicyEvaluations, resp, err
}
