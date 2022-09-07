package state_test

import (
	"fmt"
	"net/url"
	"testing"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/runatlantis/atlantis/server/neptune/workflows/internal/terraform/state"
	"github.com/stretchr/testify/assert"
)

func TestGenerate(t *testing.T) {
	route := &mux.Route{}
	route.Path("/jobs/{job-id}")
	baseURL, err := url.Parse("https://baseurl.com")

	assert.NoError(t, err)

	subject := &state.OutputURLGenerator{
		URLBuilder: route,
	}

	jobID := uuid.New()

	url, err := subject.Generate(jobID, baseURL)
	assert.NoError(t, err)

	expectedURL := fmt.Sprintf("https://baseurl.com/jobs/%s", jobID.String())
	assert.Equal(t, expectedURL, url.String())
}
