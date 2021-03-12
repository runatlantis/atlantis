package events

import (
	"errors"
	"reflect"
	"testing"

	"github.com/hashicorp/go-version"
	"github.com/runatlantis/atlantis/server/events/models"
	"github.com/segmentio/stats/v4"
	"github.com/segmentio/stats/v4/statstest"
)

type stepRunner struct{}

func (c *stepRunner) Run(ctx models.ProjectCommandContext, extraArgs []string, path string, envs map[string]string) (string, error) {
	return "success", nil
}

type errStepRunner struct{}

func (c *errStepRunner) Run(ctx models.ProjectCommandContext, extraArgs []string, path string, envs map[string]string) (string, error) {
	return "TLS handshake timeout", errors.New("error")
}

func TestInstrumentation(t *testing.T) {
	tests := []struct {
		scenario string
		function func(*testing.T, *stats.Engine)
	}{
		{
			scenario: "failing calls to ListRecentBuilds reports error metric",
			function: testStepRunError,
		},
		{
			scenario: "successful calls to Run reports timing and success metrics",
			function: testStepRunSuccess,
		},
	}

	for _, test := range tests {
		testFunc := test.function
		t.Run(test.scenario, func(t *testing.T) {
			t.Parallel()
			h := &statstest.Handler{}
			testFunc(t, stats.NewEngine("test", h))
		})
	}
}

func testStepRunSuccess(t *testing.T, eng *stats.Engine) {
	ic := InstrumentStepRunner(&stepRunner{}, eng, "plan")
	tfVersion, _ := version.NewVersion("0.12.18")
	projectContext := models.ProjectCommandContext{
		CommandName:      models.PlanCommand,
		Workspace:        "test-workspace",
		TerraformVersion: tfVersion,
	}
	ic.Run(projectContext, nil, "/path", nil)

	found := measures(t, eng)
	if len(found) != 2 {
		t.Fatalf("expected 2 measures got %d", len(found))
	}

	checkHistogramEqual(t, found[0], stats.Measure{
		Name:   "test.steps",
		Fields: []stats.Field{stats.MakeField("duration", 1, stats.Histogram)},
		Tags: []stats.Tag{
			{Name: "command", Value: "plan"},
			{Name: "stamp", Value: "total"},
			{Name: "step", Value: "plan"},
			{Name: "terraform_version", Value: "0.12.18"},
			{Name: "workspace", Value: "test-workspace"},
		},
	})

	checkCounterEqual(t, found[1], stats.Measure{
		Name:   "test.steps",
		Fields: []stats.Field{stats.MakeField("success", 1, stats.Counter)},
		Tags: []stats.Tag{
			{Name: "command", Value: "plan"},
			{Name: "step", Value: "plan"},
			{Name: "terraform_version", Value: "0.12.18"},
			{Name: "workspace", Value: "test-workspace"},
		},
	})
}

func testStepRunError(t *testing.T, eng *stats.Engine) {
	ic := InstrumentStepRunner(&errStepRunner{}, eng, "plan")
	tfVersion, _ := version.NewVersion("0.12.18")
	projectContext := models.ProjectCommandContext{
		CommandName:      models.PlanCommand,
		Workspace:        "test-workspace",
		TerraformVersion: tfVersion,
	}
	ic.Run(projectContext, nil, "/path", nil)

	found := measures(t, eng)
	if len(found) != 1 {
		t.Fatalf("expected 1 measures got %d", len(found))
	}

	checkCounterEqual(t, found[0], stats.Measure{
		Name:   "test.steps",
		Fields: []stats.Field{stats.MakeField("error", 1, stats.Counter)},
		Tags: []stats.Tag{
			{Name: "command", Value: "plan"},
			{Name: "error_type", Value: "tls"},
			{Name: "step", Value: "plan"},
			{Name: "terraform_version", Value: "0.12.18"},
			{Name: "workspace", Value: "test-workspace"},
		},
	})
}

func checkCounterEqual(t *testing.T, found, expected stats.Measure) {
	if !reflect.DeepEqual(found, expected) {
		t.Error("bad measures:")
		t.Logf("expected: %#v", expected)
		t.Logf("found:    %#v", found)
	}
}

func checkHistogramEqual(t *testing.T, found, expected stats.Measure) {
	if found.Name != expected.Name {
		t.Error("bad histogram name:")
		t.Logf("expected: %s", expected.Name)
		t.Logf("found:    %s", found.Name)
	}
	for i, f := range found.Fields {
		if f.Name != expected.Fields[i].Name {
			t.Error("bad histogram field name:")
			t.Logf("expected: %#v", expected.Fields[i].Name)
			t.Logf("found:    %#v", f.Name)
		}
	}
	if !reflect.DeepEqual(found.Tags, expected.Tags) {
		t.Error("bad histogram tags:")
		t.Logf("expected: %#v", expected.Tags)
		t.Logf("found:    %#v", found.Tags)
	}
}

func measures(t *testing.T, eng *stats.Engine) []stats.Measure {
	return eng.Handler.(*statstest.Handler).Measures()
}
