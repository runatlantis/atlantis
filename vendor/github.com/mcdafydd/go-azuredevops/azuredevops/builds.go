package azuredevops

import (
	"context"
	"fmt"
	"net/http"
)

// BuildsService handles communication with the builds methods on the API
// utilising https://docs.microsoft.com/en-gb/rest/api/vsts/build/builds
type BuildsService struct {
	client *Client
}

// BuildsListResponse is the wrapper around the main response for the List of Builds
type BuildsListResponse struct {
	Count  int      `json:"count"`
	Builds []*Build `json:"value"`
}

// TaskOrchestrationPlanReference The orchestration plan for the build.
type TaskOrchestrationPlanReference struct {
	Type   *int    `json:"orchestrationType,omitempty"`
	PlanID *string `json:"planId,omitempty"`
}

// Build Represents a build.
type Build struct {
	Definition          *BuildDefinition                  `json:"definition,omitempty"`
	Controller          *BuildController                  `json:"controller,omitempty"`
	LastChangedBy       *IdentityRef                      `json:"lastChangedBy,omitempty"`
	DeletedBy           *IdentityRef                      `json:"deletedBy,omitempty"`
	BuildNumber         *string                           `json:"buildNumber,omitempty"`
	FinishTime          *string                           `json:"finishTime,omitempty"`
	SourceBranch        *string                           `json:"sourceBranch,omitempty"`
	Repository          *BuildRepository                  `json:"repository,omitempty"`
	Demands             []*BuildDemand                    `json:"demands,omitempty"`
	Logs                *BuildLogReference                `json:"logs,omitempty"`
	Project             *TeamProjectReference             `json:"project,omitempty"`
	Properties          map[string]string                 `json:"properties,omitempty"`
	Priority            *string                           `json:"priority,omitempty"`
	OrchestrationPlan   *TaskOrchestrationPlanReference   `json:"orchestrationPlan,omitempty"`
	Plans               []*TaskOrchestrationPlanReference `json:"plans,omitempty"`
	BuildNumberRevision *int                              `json:"buildNumberRevision,omitempty"`
	Deleted             *bool                             `json:"deleted,omitempty"`
	DeletedDate         *string                           `json:"deletedDate,omitempty"`
	DeletedReason       *string                           `json:"deletedReason,omitempty"`
	ID                  *int                              `json:"id,omitempty"`
	KeepForever         *bool                             `json:"keepForever,omitempty"`
	ChangedDate         *string                           `json:"lastChangedDate,omitempty"`
	Params              *string                           `json:"parameters,omitempty"`
	Quality             *string                           `json:"quality,omitempty"`
	Queue               *AgentPoolQueue                   `json:"queue,omitempty"`
	QueueOptions        map[string]string                 `json:"queue_options,omitempty"`
	QueuePosition       *int                              `json:"queuePosition,omitempty"`
	QueueTime           *string                           `json:"queueTime,omitempty"`
	RetainedByRelease   *bool                             `json:"retainedByRelease,omitempty"`
	Version             *string                           `json:"sourceVersion,omitempty"`
	StartTime           *string                           `json:"startTime,omitempty"`
	Status              *string                           `json:"status,omitempty"`
	Result              *string                           `json:"result,omitempty"`
	ValidationResults   []*ValidationResult               `json:"validationResult,omitempty"`
	Tags                []*string                         `json:"tags,omitempty"`
	TriggerBuild        *Build                            `json:"triggeredByBuild,omitempty"`
	URI                 *string                           `json:"uri,omitempty"`
	URL                 *string                           `json:"url,omitempty"`
	TriggerInfo         *TriggerInfo                      `json:"triggerInfo,omitempty"`
}

// AgentPoolQueue The queue. This is only set if the definition type is Build.
type AgentPoolQueue struct {
	Links *map[string]Link        `json:"_links,omitempty"`
	ID    *int                    `json:"id,omitempty"`
	Name  *string                 `json:"name,omitempty"`
	URL   *string                 `json:"url,omitempty"`
	Pool  *TaskAgentPoolReference `json:"pool,omitempty"`
}

// TaskAgentPoolReference Represents a reference to an agent pool.
type TaskAgentPoolReference struct {
	ID       *int    `json:"id,omitempty"`
	IsHosted *bool   `json:"is_hosted,omitempty"`
	Name     *string `json:"name,omitempty"`
}

// TriggerInfo Source provider-specific information about what triggered the build.
type TriggerInfo struct {
	CiSourceBranch *string `json:"ci.sourceBranch,omitempty"`
	CiSourceSha    *string `json:"ci.sourceSha,omitempty"`
	CiMessage      *string `json:"ci.message,omitempty"`
}

// ValidationResult Represents the result of validating a build request.
type ValidationResult struct {
	Message *string `json:"message,omitempty"`
	Result  *string `json:"result,omitempty"`
}

// BuildDemand Represents a demand used by a definition or build.
type BuildDemand struct {
	Name  *string `json:"name,omitempty"`
	Value *string `json:"value,omitempty"`
}

// BuildListOrder is enum type for build list order
type BuildListOrder string

const (
	// FinishTimeAscending orders by finish build time asc
	FinishTimeAscending BuildListOrder = "finishTimeAscending"
	// FinishTimeDescending orders by finish build time desc
	FinishTimeDescending BuildListOrder = "finishTimeDescending"
	// QueueTimeAscending orders by build queue time asc
	QueueTimeAscending BuildListOrder = "queueTimeAscending"
	// QueueTimeDescending orders by build queue time desc
	QueueTimeDescending BuildListOrder = "queueTimeDescending"
	// StartTimeAscending orders by build start time asc
	StartTimeAscending BuildListOrder = "startTimeAscending"
	// StartTimeDescending orders by build start time desc
	StartTimeDescending BuildListOrder = "startTimeDescending"
)

// BuildsListOptions describes what the request to the API should look like
type BuildsListOptions struct {
	Definitions      *string `url:"definitions,omitempty"`
	Branch           *string `url:"branchName,omitempty"`
	Count            *int    `url:"$top,omitempty"`
	Repository       *string `url:"repositoryId,omitempty"`
	BuildIDs         *string `url:"buildIds,omitempty"`
	Order            *string `url:"queryOrder,omitempty"`
	Deleted          *string `url:"deletedFilter,omitempty"`
	MaxPerDefinition *string `url:"maxBuildsPerDefinition,omitempty"`
	Token            *string `url:"continuationToken,omitempty"`
	Props            *string `url:"properties,omitempty"`
	Tags             *string `url:"tagFilters,omitempty"`
	Result           *string `url:"resultFilter,omitempty"`
	Status           *string `url:"statusFilter,omitempty"`
	Reason           *string `url:"reasonFilter,omitempty"`
	UserID           *string `url:"requestedFor,omitempty"`
	MaxTime          *string `url:"maxTime,omitempty"`
	MinTime          *string `url:"minTime,omitempty"`
	BuildNumber      *string `url:"buildNumber,omitempty"`
	Queues           *string `url:"queues,omitempty"`
	RepoType         *string `url:"repositoryType,omitempty"`
}

// BuildLogReference Information about the build logs.
type BuildLogReference struct {
	ID   *int    `json:"id"`
	Type *string `json:"type"`
	URL  *string `json:"url"`
}

// List returns list of the builds
// utilising https://docs.microsoft.com/en-gb/rest/api/vsts/build/builds/list
func (s *BuildsService) List(ctx context.Context, owner string, project string, opts *BuildsListOptions) ([]*Build, *http.Response, error) {
	URL := fmt.Sprintf("%s/%s/_apis/build/builds?api-version=5.1-preview.1",
		owner,
		project,
	)
	URL, err := addOptions(URL, opts)

	req, err := s.client.NewRequest("GET", URL, nil)
	if err != nil {
		return nil, nil, err
	}
	r := new(BuildsListResponse)
	resp, err := s.client.Execute(ctx, req, r)

	return r.Builds, resp, err
}

// QueueBuildOptions describes what the request to the API should look like
type QueueBuildOptions struct {
	IgnoreWarnings bool   `url:"ignoreWarnings,omitempty"`
	CheckInTicket  string `url:"checkInTicket,omitempty"`
}

// Queue inserts new build creation to queue
// Requires build ID number in build.definition
// Example body:
// {"definition": {"id": 1}, "sourceBranch": "refs/heads/master"}
// utilising https://docs.microsoft.com/en-us/rest/api/azure/devops/build/Builds/Queue
func (s *BuildsService) Queue(ctx context.Context, owner string, project string, build *Build, opts *QueueBuildOptions) (*Build, *http.Response, error) {
	URL := fmt.Sprintf("%s/%s/_apis/build/builds?api-version=5.1-preview.5",
		owner,
		project,
	)
	URL, err := addOptions(URL, opts)

	if err != nil {
		return nil, nil, err
	}

	req, err := s.client.NewRequest("POST", URL, build)
	if err != nil {
		return nil, nil, err
	}
	r := new(Build)
	resp, err := s.client.Execute(ctx, req, r)

	return r, resp, err
}
