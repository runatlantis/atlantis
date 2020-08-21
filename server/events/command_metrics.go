package events

import (
	"strings"
	"time"

	"github.com/runatlantis/atlantis/server/events/models"
	"github.com/segmentio/stats"
)

type instrumentedStepRunner struct {
	stats  *stats.Engine
	runner StepRunner
	step   string
}

type instrumentedCustomRunner struct {
	stats  *stats.Engine
	runner CustomStepRunner
}

// InstrumentStepRunner wraps step runners to provide metrics for: init, plan, and apply steps
func InstrumentStepRunner(runner StepRunner, eng *stats.Engine, step string) StepRunner {
	return &instrumentedStepRunner{
		stats:  eng,
		runner: runner,
		step:   step,
	}
}

// InstrumentCustomRunner wraps step runners to provide metrics for: run steps
func InstrumentCustomRunner(runner CustomStepRunner, eng *stats.Engine) CustomStepRunner {
	return &instrumentedCustomRunner{
		stats:  eng,
		runner: runner,
	}
}

func (ic *instrumentedStepRunner) Run(ctx models.ProjectCommandContext, extraArgs []string, path string) (string, error) {
	start := time.Now()
	tags := []stats.Tag{
		{Name: "step", Value: ic.step},
		{Name: "command", Value: ctx.Command.String()},
		{Name: "workspace", Value: ctx.Workspace},
		{Name: "terraform_version", Value: ctx.TerraformVersion.String()},
	}

	out, err := ic.runner.Run(ctx, extraArgs, path)
	if err != nil {
		tags = append(tags, stats.Tag{Name: "error_type", Value: errorTag(out)})

		ic.stats.Incr("steps.error", tags...)
		return out, err
	}
	ic.stats.ClockAt("steps.duration", start, tags...).Stop()
	ic.stats.Incr("steps.success", tags...)
	return out, err
}

func (ic *instrumentedCustomRunner) Run(ctx models.ProjectCommandContext, cmd string, path string) (string, error) {
	start := time.Now()
	tags := []stats.Tag{
		{Name: "step", Value: "run"},
		{Name: "command", Value: ctx.Command.String()},
		{Name: "workspace", Value: ctx.Workspace},
		{Name: "terraform_version", Value: ctx.TerraformVersion.String()},
	}

	out, err := ic.runner.Run(ctx, cmd, path)
	if err != nil {
		ic.stats.Incr("steps.error", tags...)
		return out, err
	}
	ic.stats.ClockAt("steps.duration", start, tags...).Stop()
	ic.stats.Incr("steps.success", tags...)
	return out, err
}

// TODO: Make this list a configurable map
func errorTag(output string) string {
	if strings.Contains(output, "TLS handshake timeout") {
		return "tls"
	} else if strings.Contains(output, "error initializing backend") {
		return "tls"
	} else if strings.Contains(output, "failed to execute \"bash\"") {
		return "bash"
	} else if strings.Contains(output, "Could not satisfy plugin requirements") {
		return "plugins"
	} else {
		return "other"
	}
}
