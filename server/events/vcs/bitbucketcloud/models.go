package bitbucketcloud

const (
	PullCreatedHeader        = "pullrequest:created"
	PullUpdatedHeader        = "pullrequest:updated"
	PullFulfilledHeader      = "pullrequest:fulfilled"
	PullRejectedHeader       = "pullrequest:rejected"
	PullCommentCreatedHeader = "pullrequest:comment_created"
)

type CommentEvent struct {
	CommonEventData
	Comment *Comment `json:"comment,omitempty" validate:"required"`
}

type PullRequestEvent struct {
	CommonEventData
}

type CommonEventData struct {
	Actor       *Actor       `json:"actor,omitempty" validate:"required"`
	Repository  *Repository  `json:"repository,omitempty" validate:"required"`
	PullRequest *PullRequest `json:"pullrequest,omitempty" validate:"required"`
}

type DiffStat struct {
	Values []DiffStatValue `json:"values,omitempty" validate:"required"`
	Next   *string         `json:"next,omitempty"`
}
type DiffStatValue struct {
	Status *string `json:"status,omitempty" validate:"required"`
	// Old is the old file, this can be null.
	Old *DiffStatFile `json:"old,omitempty"`
	// New is the new file, this can be null.
	New *DiffStatFile `json:"new,omitempty"`
}
type DiffStatFile struct {
	Path *string `json:"path,omitempty" validate:"required"`
}

type Actor struct {
	AccountID *string `json:"account_id,omitempty" validate:"required"`
}
type Repository struct {
	FullName *string `json:"full_name,omitempty" validate:"required"`
	Links    Links   `json:"links,omitempty" validate:"required"`
}
type PullRequest struct {
	ID           *int          `json:"id,omitempty" validate:"required"`
	Source       *BranchMeta   `json:"source,omitempty" validate:"required"`
	Destination  *BranchMeta   `json:"destination,omitempty" validate:"required"`
	Participants []Participant `json:"participants,omitempty" validate:"required"`
	Links        *Links        `json:"links,omitempty" validate:"required"`
	State        *string       `json:"state,omitempty" validate:"required"`
	Author       *Author       `jsonN:"author,omitempty" validate:"required"`
}
type Links struct {
	HTML *Link `json:"html,omitempty" validate:"required"`
}
type Link struct {
	HREF *string `json:"href,omitempty" validate:"required"`
}
type Participant struct {
	Approved *bool `json:"approved,omitempty" validate:"required"`
	User     *struct {
		UUID *string `json:"uuid,omitempty" validate:"required"`
	} `json:"user,omitempty" validate:"required"`
}
type BranchMeta struct {
	Repository *Repository `json:"repository,omitempty" validate:"required"`
	Commit     *Commit     `json:"commit,omitempty" validate:"required"`
	Branch     *Branch     `json:"branch,omitempty" validate:"required"`
}
type Branch struct {
	Name *string `json:"name,omitempty" validate:"required"`
}
type Commit struct {
	Hash *string `json:"hash,omitempty" validate:"required"`
}
type Comment struct {
	Content *CommentContent `json:"content,omitempty" validate:"required"`
}
type CommentContent struct {
	Raw *string `json:"raw,omitempty" validate:"required"`
}
type Author struct {
	UUID *string `json:"uuid,omitempty" validate:"required"`
}
