package metrics

const (
	ExecutionTimeMetric    = "execution_time"
	ExecutionSuccessMetric = "execution_success"
	ExecutionErrorMetric   = "execution_error"
	ExecutionFailureMetric = "execution_failure"

	FilterPresentMetric = "present"
	FilterAbsentMetric  = "absent"
	FilterErrorMetric   = "error"

	RootTag = "root"
	RepoTag = "repo"

	ActivityExecutionSuccess = "activity_execution_success"
	ActivityExecutionFailure = "activity_execution_failure"

	// Note: This is specifically calculated when the activity starts (not scheduled)
	ActivityExecutionLatency = "activity_execution_latency"

	SignalNameTag = "signal_name"

	// Signal handling metrics before it is added to a buffered channel
	SignalHandleSuccess = "signal_handle_success"
	SignalHandleFailure = "signal_handle_failure"
	SignalHandleLatency = "signal_handle_latency"

	// Signal receive is when we receive it off the channel
	SignalReceive = "signal_receive"

	// Metrics are scoped to workflow namespaces anyways so let's
	// keep these metrics simple.
	WorkflowSuccess = "success"
	WorkflowFailure = "failure"
	WorkflowLatency = "latency"

	ManualOverride          = "manual_override"
	ManualOverrideReasonTag = "reason"
)
