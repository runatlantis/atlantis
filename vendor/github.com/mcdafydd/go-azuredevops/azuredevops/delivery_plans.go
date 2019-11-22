package azuredevops

import (
	"context"
	"fmt"
	"net/http"
	"time"
)

// DeliveryPlansService handles communication with the deliverytimeline methods on the API
// utilising https://docs.microsoft.com/en-us/rest/api/vsts/work/deliverytimeline
type DeliveryPlansService struct {
	client *Client
}

// DeliveryPlansListResponse describes the delivery plans list response
type DeliveryPlansListResponse struct {
	Count         int             `json:"count"`
	DeliveryPlans []*DeliveryPlan `json:"value"`
}

// DeliveryPlanTimeLine describes the delivery plan get response
type DeliveryPlanTimeLine struct {
	StartDate *string         `json:"startDate,omitempty"`
	EndDate   *string         `json:"endDate,omitempty"`
	ID        *string         `json:"id,omitempty"`
	Revision  *int            `json:"revision,omitempty"`
	Teams     []*DeliveryTeam `json:"teams,omitempty"`
}

// DeliveryPlan describes an plan
type DeliveryPlan struct {
	ID      *string `json:"id,omitempty"`
	Name    *string `json:"name,omitempty"`
	Type    *string `json:"type,omitempty"`
	Created *string `json:"createdDate,omitempty"`
	URL     *string `json:"url,omitempty"`
}

// DeliveryTeam describes the teams in a specific plan
type DeliveryTeam struct {
	ID         *string      `json:"id,omitempty"`
	Name       *string      `json:"name,omitempty"`
	Iterations []*Iteration `json:"iterations,omitempty"`
}

const (
	// DeliveryPlanWorkItemIDKey is the key for which part of the workItems[] slice has the ID
	DeliveryPlanWorkItemIDKey = 0
	// DeliveryPlanWorkItemIterationKey is the key for which part of the workItems[] slice has the Iteration
	DeliveryPlanWorkItemIterationKey = 1
	// DeliveryPlanWorkItemTypeKey is the key for which part of the workItems[] slice has the Type
	DeliveryPlanWorkItemTypeKey = 2
	// DeliveryPlanWorkItemNameKey is the key for which part of the workItems[] slice has the Name
	DeliveryPlanWorkItemNameKey = 4
	// DeliveryPlanWorkItemStatusKey is the key for which part of the workItems[] slice has the Status
	DeliveryPlanWorkItemStatusKey = 5
	// DeliveryPlanWorkItemTagKey is the key for which part of the workItems[] slice has the Tag
	DeliveryPlanWorkItemTagKey = 6
)

// DeliveryPlansListOptions describes what the request to the API should look like
type DeliveryPlansListOptions struct {
}

// List returns a list of delivery plans
func (s *DeliveryPlansService) List(ctx context.Context, owner string, project string, opts *DeliveryPlansListOptions) ([]*DeliveryPlan, *http.Response, error) {
	URL := fmt.Sprintf("%s/%s/_apis/work/plans?api-version=5.1-preview.1",
		owner,
		project,
	)
	URL, err := addOptions(URL, opts)

	req, err := s.client.NewRequest("GET", URL, nil)
	if err != nil {
		return nil, nil, err
	}
	r := new(DeliveryPlansListResponse)
	resp, err := s.client.Execute(ctx, req, r)

	return r.DeliveryPlans, resp, err
}

// GetTimeLine will fetch the details about a specific delivery plan
func (s *DeliveryPlansService) GetTimeLine(ctx context.Context, owner string, project string, ID string, startDate, endDate string) (*DeliveryPlanTimeLine, *http.Response, error) {
	URL := fmt.Sprintf(
		"%s/%s/_apis/work/plans/%s/deliverytimeline?api-version=5.1-preview.1",
		owner,
		project,
		ID,
	)

	if startDate == "" {
		startDate = time.Now().Format("2006-01-02")
		// The 65 date thing is arbitrary from the API
		endDate = time.Now().AddDate(0, 0, 65).Format("2006-01-02")
	}

	URL = fmt.Sprintf(URL+"&startDate=%s&endDate=%s", startDate, endDate)

	req, err := s.client.NewRequest("GET", URL, nil)
	if err != nil {
		return nil, nil, err
	}
	r := new(DeliveryPlanTimeLine)
	resp, err := s.client.Execute(ctx, req, r)

	return r, resp, err
}
