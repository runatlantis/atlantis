package azuredevops

import (
	"context"
	"fmt"
	"net/http"
)

// BuildDefinitionsService handles communication with the build definitions methods on the API
// utilising https://docs.microsoft.com/en-gb/rest/api/vsts/build/definitions
type BuildDefinitionsService struct {
	client *Client
}

// BuildDefinitionsListResponse describes the build definitions list response
type BuildDefinitionsListResponse struct {
	Count            int                `json:"count"`
	BuildDefinitions []*BuildDefinition `json:"value"`
}

// BuildRepository represents a repository used by a build definition
type BuildRepository struct {
	ID                 *string                `json:"id,omitempty"`
	Type               *string                `json:"type,omitempty"`
	Name               *string                `json:"name,omitempty"`
	URL                *string                `json:"url,omitempty"`
	RootFolder         *string                `json:"root_folder"`
	Properties         map[string]interface{} `json:"properties"`
	Clean              *string                `json:"clean"`
	DefaultBranch      *string                `json:"default_branch"`
	CheckoutSubmodules *bool                  `json:"checkout_submodules"`
}

// BuildDefinition represents a build definition
type BuildDefinition struct {
	ID         *int             `json:"id,omitempty"`
	Name       *string          `json:"name,omitempty"`
	Repository *BuildRepository `json:"repository,omitempty"`
}

// BuildDefinitionsListOptions describes what the request to the API should look like
type BuildDefinitionsListOptions struct {
	Path                 *string `url:"path,omitempty"`
	IncludeAllProperties *bool   `url:"includeAllProperties,omitempty"`
}

// List returns a list of build definitions
// utilising https://docs.microsoft.com/en-gb/rest/api/vsts/build/definitions/list
func (s *BuildDefinitionsService) List(ctx context.Context, owner string, project string, opts *BuildDefinitionsListOptions) ([]*BuildDefinition, *http.Response, error) {
	URL := fmt.Sprintf("%s/%s/_apis/build/definitions?api-version=5.1-preview.1",
		owner,
		project,
	)
	URL, err := addOptions(URL, opts)

	req, err := s.client.NewRequest("GET", URL, nil)
	if err != nil {
		return nil, nil, err
	}
	r := new(BuildDefinitionsListResponse)
	resp, err := s.client.Execute(ctx, req, r)

	return r.BuildDefinitions, resp, err
}
