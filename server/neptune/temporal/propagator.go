package temporal

import (
	"context"
	internalContext "github.com/runatlantis/atlantis/server/neptune/context"

	"github.com/pkg/errors"
	"go.temporal.io/sdk/converter"
	"go.temporal.io/sdk/workflow"
)

const (
	propagationKey = "propagation-key"
)

type ctxPropagator struct{}

func (p *ctxPropagator) Inject(ctx context.Context, writer workflow.HeaderWriter) error {
	value := internalContext.ExtractFields(ctx)
	payload, err := converter.GetDefaultDataConverter().ToPayload(value)

	if err != nil {
		return errors.Wrap(err, "converting context values to payload")
	}

	writer.Set(propagationKey, payload)
	return nil
}

func (p *ctxPropagator) InjectFromWorkflow(ctx workflow.Context, writer workflow.HeaderWriter) error {
	value := internalContext.ExtractFields(ctx)
	payload, err := converter.GetDefaultDataConverter().ToPayload(value)

	if err != nil {
		return errors.Wrap(err, "converting context values to payload")
	}

	writer.Set(propagationKey, payload)
	return nil

}

func (p *ctxPropagator) Extract(ctx context.Context, reader workflow.HeaderReader) (context.Context, error) {
	if value, ok := reader.Get(propagationKey); ok {
		var values map[string]interface{}

		if err := converter.GetDefaultDataConverter().FromPayload(value, &values); err != nil {
			return ctx, errors.Wrap(err, "extracting context values")
		}

		for k, v := range values {
			ctx = context.WithValue(ctx, internalContext.Key(k), v)
		}
	}
	return ctx, nil
}

func (p *ctxPropagator) ExtractToWorkflow(ctx workflow.Context, reader workflow.HeaderReader) (workflow.Context, error) {
	if value, ok := reader.Get(propagationKey); ok {
		var values map[string]interface{}

		if err := converter.GetDefaultDataConverter().FromPayload(value, &values); err != nil {
			return ctx, errors.Wrap(err, "extracting context values")
		}

		for k, v := range values {
			ctx = workflow.WithValue(ctx, internalContext.Key(k), v)
		}
	}

	return ctx, nil
}
