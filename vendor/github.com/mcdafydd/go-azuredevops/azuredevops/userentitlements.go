package azuredevops

import (
	"context"
	"fmt"
	"net/http"
)

// UserEntitlementsService handles communication with the user entitlements methods on the API
// utilising https://docs.microsoft.com/en-us/rest/api/azure/devops/memberentitlementmanagement/user%20entitlements?view=azure-devops-rest-6.0
type UserEntitlementsService struct {
	client *Client
}

// UserEntitlements is a wrapper class around the main response for the Get of UserEntitlement
type UserEntitlements struct {
	Members           []Item      `json:"members,omitempty"`
	ContinuationToken interface{} `json:"continuationToken"`
	TotalCount        int64       `json:"totalCount,omitempty"`
	Items             []Item      `json:"items,omitempty"`
}

// Item is a wrapper class used by UserEntitlements
type Item struct {
	ID                  string        `json:"id,omitempty"`
	User                User          `json:"user"`
	AccessLevel         AccessLevel   `json:"accessLevel"`
	LastAccessedDate    string        `json:"lastAccessedDate,omitempty"`
	DateCreated         string        `json:"dateCreated,omitempty"`
	ProjectEntitlements []interface{} `json:"projectEntitlements,omitempty"`
	Extensions          []interface{} `json:"extensions,omitempty"`
	GroupAssignments    []interface{} `json:"groupAssignments,omitempty"`
}

// AccessLevel is a wrapper class used by Item
type AccessLevel struct {
	LicensingSource    string `json:"licensingSource,omitempty"`
	AccountLicenseType string `json:"accountLicenseType,omitempty"`
	MSDNLicenseType    string `json:"msdnLicenseType,omitempty"`
	LicenseDisplayName string `json:"licenseDisplayName,omitempty"`
	Status             string `json:"status,omitempty"`
	StatusMessage      string `json:"statusMessage,omitempty"`
	AssignmentSource   string `json:"assignmentSource,omitempty"`
}

// User is a wrapper class used by Item
type User struct {
	SubjectKind   string  `json:"subjectKind,omitempty"`
	MetaType      *string `json:"metaType,omitempty"`
	Domain        string  `json:"domain,omitempty"`
	PrincipalName string  `json:"principalName,omitempty"`
	MailAddress   string  `json:"mailAddress,omitempty"`
	Origin        string  `json:"origin,omitempty"`
	OriginID      string  `json:"originId,omitempty"`
	DisplayName   string  `json:"displayName,omitempty"`
	Links         Links   `json:"_links,omitempty"`
	URL           string  `json:"url,omitempty"`
	Descriptor    string  `json:"descriptor,omitempty"`
}

// Links is a wrapper class used by User
type Links struct {
	Self            Avatar `json:"self"`
	Memberships     Avatar `json:"memberships"`
	MembershipState Avatar `json:"membershipState"`
	StorageKey      Avatar `json:"storageKey"`
	Avatar          Avatar `json:"avatar"`
}

// Avatar is a wrapper class used by Links
type Avatar struct {
	Href string `json:"href"`
}

// Get returns a single user entitlement filtering by the user name in the organization
// https://docs.microsoft.com/en-us/rest/api/azure/devops/memberentitlementmanagement/user%20entitlements/search%20user%20entitlements?view=azure-devops-rest-6.0
func (s *UserEntitlementsService) Get(ctx context.Context, userName string, orgName string) (*UserEntitlements, *http.Response, error) {
	URL := fmt.Sprintf("%s/%s/_apis/userentitlements?$filter=name+eq+'%s'&$api-version=6.0-preview.3",
		s.client.VsaexBaseURL.String(),
		orgName,
		userName,
	)

	req, err := s.client.NewRequest("GET", URL, nil)
	if err != nil {
		return nil, nil, err
	}

	r := new(UserEntitlements)
	resp, err := s.client.Execute(ctx, req, r)
	if err != nil {
		return nil, nil, err
	}

	return r, resp, err
}

// GetUserID returns the user id by the user name and the organizatino name
func (s *UserEntitlementsService) GetUserID(ctx context.Context, userName string, orgName string) (*string, error) {
	userEntitlements, _, err := s.Get(context.Background(), userName, orgName)
	if err != nil {
		return nil, err
	}

	if len(userEntitlements.Items) > 0 {
		return &userEntitlements.Items[0].ID, nil
	}
	var nilValue *string = nil
	return nilValue, nil
}
