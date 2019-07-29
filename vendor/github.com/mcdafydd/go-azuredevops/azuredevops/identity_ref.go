package azuredevops

// IdentityRef describes an Azure Devops identity
type IdentityRef struct {
	Links             *map[string]Link `json:"_links,omitempty"`
	Descriptor        *string          `json:"descriptor,omitempty"`
	DirectoryAlias    *string          `json:"directoryAlias,omitempty"`
	DisplayName       *string          `json:"displayName,omitempty"`
	ID                *string          `json:"id,omitempty"`
	ImageURL          *string          `json:"imageUrl,omitempty"`
	Inactive          *bool            `json:"inactive,omitempty"`
	IsAadIdentity     *bool            `json:"isAadIdentity,omitempty"`
	IsContainer       *bool            `json:"isContainer,omitempty"`
	IsDeletedInOrigin *bool            `json:"isDeletedInOrigin,omitempty"`
	ProfileURL        *string          `json:"profileUrl,omitempty"`
	URL               *string          `json:"url,omitempty"`
	UniqueName        *string          `json:"uniqueName,omitempty"`
}

// IdentityRefWithVote Identity information including a vote on a pull request.
type IdentityRefWithVote struct {
	IdentityRef
	IsRequired  *bool                  `json:"isRequired,omitempty"`
	ReviewerURL *string                `json:"reviewerUrl,omitempty"`
	Vote        *int                   `json:"vote,omitempty"`
	VotedFor    []*IdentityRefWithVote `json:"votedFor,omitempty"`
}
