package temporal

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"github.com/runatlantis/atlantis/server/events/metrics"
	contextInternal "github.com/runatlantis/atlantis/server/neptune/gateway/context"
	"io"
	"time"

	"github.com/runatlantis/atlantis/server/core/config/valid"
	"github.com/uber-go/tally/v4"
	"go.temporal.io/sdk/client"
	temporaltally "go.temporal.io/sdk/contrib/tally"
	"go.temporal.io/sdk/converter"
	"go.temporal.io/sdk/interceptor"
	"go.temporal.io/sdk/workflow"
	"logur.dev/logur"
)

const StatsNamespace = "temporalclient"

type Options struct {
	Interceptors  []interceptor.ClientInterceptor
	StatsReporter tally.StatsReporter
}

func (o *Options) WithClientInterceptors(i ...interceptor.ClientInterceptor) *Options {
	o.Interceptors = i
	return o
}

func (o *Options) WithStatsReporter(reporter tally.StatsReporter) *Options {
	o.StatsReporter = reporter
	return o
}

func NewClient(logger logur.Logger, cfg valid.Temporal, options *Options) (*ClientWrapper, error) {
	opts := client.Options{
		Namespace:          cfg.Namespace,
		Logger:             logur.LoggerToKV(logger),
		ContextPropagators: []workflow.ContextPropagator{&ctxPropagator{}},
		Interceptors:       options.Interceptors,
	}

	var clientScope tally.Scope
	var clientScopeCloser io.Closer
	if options.StatsReporter != nil {
		clientScope, clientScopeCloser = tally.NewRootScope(tally.ScopeOptions{
			Prefix:   StatsNamespace,
			Reporter: options.StatsReporter,
		}, time.Second)
		opts.MetricsHandler = temporaltally.NewMetricsHandler(clientScope)
	}

	if cfg.UseSystemCACert {
		certs, err := x509.SystemCertPool()
		if err != nil {
			return nil, err
		}
		opts.ConnectionOptions = client.ConnectionOptions{
			TLS: &tls.Config{
				RootCAs:    certs,
				MinVersion: tls.VersionTLS12,
			},
		}
	}

	if cfg.Host != "" || cfg.Port != "" {
		opts.HostPort = fmt.Sprintf("%s:%s", cfg.Host, cfg.Port)

	}

	client, err := client.Dial(opts)
	if err != nil {
		return &ClientWrapper{}, err
	}

	return &ClientWrapper{
		Client:      client,
		statsCloser: clientScopeCloser,
	}, nil
}

// ClientWrapper for now just exists to intercept Close()
// so we can clean up our stats object.
// Users still should feel free to refer to the underlying
// client as necessary
type ClientWrapper struct {
	client.Client
	statsCloser io.Closer
}

func (c *ClientWrapper) Close() {
	c.Client.Close()

	if c.statsCloser != nil {
		c.statsCloser.Close()
	}
}

func NewMetricsInterceptor(scope tally.Scope) interceptor.ClientInterceptor {
	return &clientMetricsInterceptor{
		scope: scope,
	}
}

type clientMetricsInterceptor struct {
	scope tally.Scope
	// according to docs we should be embedding this for future changes
	interceptor.ClientInterceptorBase
}

func (i *clientMetricsInterceptor) InterceptClient(
	next interceptor.ClientOutboundInterceptor,
) interceptor.ClientOutboundInterceptor {

	return &clientMetricsOutboundInterceptor{
		scope: i.scope,
		ClientOutboundInterceptorBase: interceptor.ClientOutboundInterceptorBase{
			Next: next,
		},
	}

}

type clientMetricsOutboundInterceptor struct {
	scope tally.Scope
	// according to docs we should be embedding this for future changes
	interceptor.ClientOutboundInterceptorBase
}

func (i *clientMetricsOutboundInterceptor) ExecuteWorkflow(ctx context.Context, in *interceptor.ClientExecuteWorkflowInput) (client.WorkflowRun, error) {
	s := i.scope.SubScope("execute_workflow")
	s = addTags(ctx, s)

	timer := s.Timer("latency").Start()
	defer timer.Stop()

	run, err := i.ClientOutboundInterceptorBase.Next.ExecuteWorkflow(ctx, in)

	if err != nil {
		s.Counter("error").Inc(1)
		return run, err
	}

	s.Counter("success").Inc(1)
	return run, err

}

func (i *clientMetricsOutboundInterceptor) SignalWorkflow(ctx context.Context, in *interceptor.ClientSignalWorkflowInput) error {
	s := i.scope.SubScope("signal_workflow")
	s = addTags(ctx, s)

	timer := s.Timer("latency").Start()
	defer timer.Stop()

	if err := i.ClientOutboundInterceptorBase.Next.SignalWorkflow(ctx, in); err != nil {
		s.Counter("error").Inc(1)
		return err
	}

	s.Counter("success").Inc(1)
	return nil
}

func (i *clientMetricsOutboundInterceptor) SignalWithStartWorkflow(ctx context.Context, in *interceptor.ClientSignalWithStartWorkflowInput) (client.WorkflowRun, error) {
	s := i.scope.SubScope("signal_with_start_workflow")
	s = addTags(ctx, s)

	timer := s.Timer("latency").Start()
	defer timer.Stop()

	run, err := i.ClientOutboundInterceptorBase.Next.SignalWithStartWorkflow(ctx, in)

	if err != nil {
		s.Counter("error").Inc(1)
		return run, err
	}

	s.Counter("success").Inc(1)
	return run, err
}

func (i *clientMetricsOutboundInterceptor) CancelWorkflow(ctx context.Context, in *interceptor.ClientCancelWorkflowInput) error {
	s := i.scope.SubScope("cancel_workflow")
	s = addTags(ctx, s)

	timer := s.Timer("latency").Start()
	defer timer.Stop()

	if err := i.ClientOutboundInterceptorBase.Next.CancelWorkflow(ctx, in); err != nil {
		s.Counter("error").Inc(1)
		return err
	}

	s.Counter("success").Inc(1)
	return nil
}

func (i *clientMetricsOutboundInterceptor) TerminateWorkflow(ctx context.Context, in *interceptor.ClientTerminateWorkflowInput) error {
	s := i.scope.SubScope("terminate_workflow")
	s = addTags(ctx, s)

	timer := s.Timer("latency").Start()
	defer timer.Stop()

	if err := i.ClientOutboundInterceptorBase.Next.TerminateWorkflow(ctx, in); err != nil {
		s.Counter("error").Inc(1)
		return err
	}

	s.Counter("success").Inc(1)
	return nil
}

func (i *clientMetricsOutboundInterceptor) QueryWorkflow(ctx context.Context, in *interceptor.ClientQueryWorkflowInput) (converter.EncodedValue, error) {
	s := i.scope.SubScope("query_workflow")
	s = addTags(ctx, s)

	timer := s.Timer("latency").Start()
	defer timer.Stop()

	val, err := i.ClientOutboundInterceptorBase.Next.QueryWorkflow(ctx, in)

	if err != nil {
		s.Counter("error").Inc(1)
		return val, err
	}

	s.Counter("success").Inc(1)
	return val, err
}

func addTags(ctx context.Context, scope tally.Scope) tally.Scope {
	tags := make(map[string]string)
	if ctx.Value(contextInternal.ProjectKey) != nil {
		tags[metrics.RootTag] = ctx.Value(contextInternal.ProjectKey).(string)
	}
	if ctx.Value(contextInternal.RepositoryKey) != nil {
		tags[metrics.RepoTag] = ctx.Value(contextInternal.RepositoryKey).(string)
	}
	return scope.Tagged(tags)
}
