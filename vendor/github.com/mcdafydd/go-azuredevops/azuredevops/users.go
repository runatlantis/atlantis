package azuredevops

import (
	"context"
	"fmt"
	"net/http"
)

// UsersService handles communication with the Graph.Users methods on the API
// utilising https://docs.microsoft.com/en-us/rest/api/azure/devops/graph/users/get
type UsersService struct {
	client *Client
}

// GraphGroup is the parent struct describing a Microsoft Graph group for Azure Devops
type GraphGroup struct {
	GraphMember
	/**
	 * A short phrase to help human readers disambiguate groups with similar names
	 */
	Description         *string `json:"description,omitempty"`
	IsCrossProject      *bool   `json:"isCrossProject,omitempty"`
	IsDeleted           *bool   `json:"isDeleted,omitempty"`
	IsGlobalScope       *bool   `json:"isGlobalScope,omitempty"`
	IsRestrictedVisible *bool   `json:"isRestrictedVisible,omitempty"`
	LocalScopeID        *string `json:"localScopeId,omitempty"`
	ScopeID             *string `json:"scopeId,omitempty"`
	ScopeName           *string `json:"scopeName,omitempty"`
	ScopeType           *string `json:"scopeType,omitempty"`
	SecuringHostID      *string `json:"securingHostId,omitempty"`
	SpecialType         *string `json:"specialType,omitempty"`
}

// GraphMember is a child of the GraphUser struct
type GraphMember struct {
	GraphSubject
	/**
	 * This represents the name of the container of origin for a graph member.
	 * (For MSA this is "Windows Live ID", for AD the name of the domain, for
	 * AAD the tenantID of the directory, for VSTS groups the ScopeId, etc)
	 */
	Domain *string `json:"domain,omitempty"`
	/**
	 * The email address of record for a given graph member. This may be
	 * different than the principal name.
	 */
	MailAddress *string `json:"mailAddress,omitempty"`
	/**
	 * This is the PrincipalName of this graph member from the source
	 * provider. The source provider may change this field over time and it
	 * is not guaranteed to be immutable for the life of the graph member by
	 * VSTS.
	 */
	PrincipalName *string `json:"principalName,omitempty"`
}

// GraphSubjectBase Base struct for other graph entities
type GraphSubjectBase struct {
	/*
	 * This field contains zero or more interesting links about the graph
	 * subject. These links may be invoked to obtain additional relationships
	 * or more detailed information about this graph subject.
	 */
	Links map[string]Link `json:"_links,omitempty"`
	/**
	 * The descriptor is the primary way to reference the graph subject while
	 * the system is running. This field will uniquely identify the same
	 * graph subject across both Accounts and Organizations.
	 */
	Descriptor *string `json:"descriptor,omitempty"`
	/**
	 * This is the non-unique display name of the graph subject. To change
	 * this field, you must alter its value in the source provider.
	 */
	DisplayName *string `json:"displayName,omitempty"`
	/**
	 * This url is the full route to the source resource of this graph subject.
	 */
	URL *string `json:"url,omitempty"`
}

// GraphSubject A graph subject entity
type GraphSubject struct {
	GraphSubjectBase
	/**
	 * [Internal Use Only] The legacy descriptor is here in case you need to
	 * access old version IMS using identity descriptor.
	 */
	LegacyDescriptor *string `json:"legacyDescriptor,omitempty"`
	/**
	 * The type of source provider for the origin identifier (ex:AD, AAD, MSA)
	 */
	Origin *string `json:"origin,omitempty"`
	/**
	 * The unique identifier from the system of origin. Typically a sid, object
	 * id or Guid. Linking and unlinking operations can cause this value to
	 * change for a user because the user is not backed by a different provider
	 * and has a different unique id in the new provider.
	 */
	OriginID *string `json:"originId,omitempty"`
	/**
	 * This field identifies the type of the graph subject (ex: Group, Scope, User).
	 */
	SubjectKind *string `json:"subjectKind,omitempty"`
}

// GraphUser is the parent struct describing a Microsoft Graph user for Azure Devops
type GraphUser struct {
	GraphMember
	IsDeletedInOrigin  *bool `json:"isDeletedOrigin,omitempty"`
	MetadataUpdateDate *Time `json:"metadataUpdateDate,omitempty"`
	/**
	* The meta type of the user in the origin, such as "member", "guest",
	* etc. See UserMetaType for the set of possible values.
	 */
	MetaType *string `json:"metaType,omitempty"`
}

// GraphUsersListResponse describes what a response from the Users.List()
// API should look like
type GraphUsersListResponse struct {
	Count      int          `json:"count"`
	GraphUsers []*GraphUser `json:"value"`
}

// Get returns information about a single user in an org
// https://docs.microsoft.com/en-us/rest/api/azure/devops/graph/users/get
func (s *UsersService) Get(ctx context.Context, owner, descriptor string) (*GraphUser, *http.Response, error) {
	URL := fmt.Sprintf("%s%s/_apis/graph/users/%s?api-version=5.1-preview.1",
		s.client.VsspsBaseURL.String(),
		owner,
		descriptor,
	)

	request, err := s.client.NewRequest("GET", URL, nil)
	if err != nil {
		return nil, nil, err
	}
	var r GraphUser
	resp, err := s.client.Execute(ctx, request, &r)

	return &r, resp, err
}

// List returns a list of users in an org
// utilising https://docs.microsoft.com/en-us/rest/api/azure/devops/graph/users/list
func (s *UsersService) List(ctx context.Context, owner string) ([]*GraphUser, *http.Response, error) {
	URL := fmt.Sprintf("%s%s/_apis/graph/users?api-version=5.1-preview.1",
		s.client.VsspsBaseURL.String(),
		owner,
	)

	request, err := s.client.NewRequest("GET", URL, nil)
	if err != nil {
		return nil, nil, err
	}
	var r GraphUsersListResponse
	resp, err := s.client.Execute(ctx, request, &r)

	return r.GraphUsers, resp, err
}

// GraphDescriptorResult Returns user descriptor and links related to the
// request
type GraphDescriptorResult struct {
	Links map[string]Link `json:"_links,omitempty"`
	Value *string         `json:"value,omitempty"`
}

// GetDescriptors returns descriptors for one or more users based on filter
// criteria
// https://docs.microsoft.com/en-us/rest/api/azure/devops/graph/descriptors/get?view=azure-devops-rest-5.1
func (s *UsersService) GetDescriptors(ctx context.Context, owner, storageKey string) (*GraphDescriptorResult, *http.Response, error) {
	URL := fmt.Sprintf("%s%s/_apis/graph/descriptors/%s?api-version=5.1-preview.1",
		s.client.VsspsBaseURL.String(),
		owner,
		storageKey,
	)

	request, err := s.client.NewRequest("GET", URL, nil)
	if err != nil {
		return nil, nil, err
	}
	r := &GraphDescriptorResult{}
	resp, err := s.client.Execute(ctx, request, r)

	return r, resp, err
}
