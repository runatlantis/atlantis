package lyft

import (
	"testing"

	"github.com/google/go-github/v45/github"
	"github.com/stretchr/testify/assert"
)

func pointer(str string) *string {
	temp := str
	return &temp
}

func TestIsPRLocked(t *testing.T) {

	// shouldn't need any fields for this
	subject := SQBasedPullStatusFetcher{}

	cases := []struct {
		description string
		statuses    []*github.RepoStatus
		isLocked    bool
	}{
		{
			"pull is locked",
			[]*github.RepoStatus{
				{
					Context:     pointer(SubmitQueueReadinessStatusContext),
					Description: pointer("{\"pr_number\": 176, \"waiting\": [\"approval\", \"lock\"]}"),
				},
			},
			true,
		},
		{
			"pull is unlocked",
			[]*github.RepoStatus{
				{
					Context:     pointer(SubmitQueueReadinessStatusContext),
					Description: pointer("{\"pr_number\": 176, \"waiting\": [\"approval\"]}"),
				},
			},
			false,
		},
		{
			"sq not found",
			[]*github.RepoStatus{
				{
					Context:     pointer("random"),
					Description: pointer("{\"pr_number\": 176, \"waiting\": [\"approval\"]}"),
				},
			},
			false,
		},
		{
			"waiting key not found",
			[]*github.RepoStatus{
				{
					Context:     pointer(SubmitQueueReadinessStatusContext),
					Description: pointer("{\"pr_number\": 176}"),
				},
			},
			false,
		},
		{
			"empty sq status",
			[]*github.RepoStatus{
				{
					Context:     pointer(SubmitQueueReadinessStatusContext),
					Description: pointer(""),
				},
			},
			false,
		},
	}

	for _, c := range cases {
		t.Run(c.description, func(t *testing.T) {
			isLocked, err := subject.isPRLocked(c.statuses)

			assert.NoError(t, err)
			assert.Equal(t, c.isLocked, isLocked)
		})
	}
}
