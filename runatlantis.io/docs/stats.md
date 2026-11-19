# Metrics/Stats

Atlantis exposes a set of metrics for each of its operations including errors, successes, and latencies.

Latency and duration metrics are reported in seconds (for example, the `*_time` metrics in the Prometheus output below), while counter metrics such as successes and errors are unitless.

::: warning NOTE
Currently Statsd and Prometheus is supported. See configuration below for details.
:::

## Configuration

Metrics are configured through the [Server Side Config](server-side-repo-config.md#metrics).

## Available Metrics

Assuming metrics are exposed from the endpoint `/metrics` from the [metrics](server-side-repo-config.md#metrics) server side config e.g.

```yaml
metrics:
  prometheus:
    endpoint: "/metrics"
```

To see all the metrics exposed from atlantis service, make a GET request to the `/metrics` endpoint.

```bash
curl localhost:4141/metrics
# HELP atlantis_builder_execution_error atlantis_builder_execution_error counter
# TYPE atlantis_builder_execution_error counter
atlantis_builder_execution_error 0
# HELP atlantis_builder_execution_success atlantis_builder_execution_success counter
# TYPE atlantis_builder_execution_success counter
atlantis_builder_execution_success 10
# HELP atlantis_builder_execution_time atlantis_builder_execution_time summary
# TYPE atlantis_builder_execution_time summary
atlantis_builder_execution_time{quantile="0.5"} NaN
atlantis_builder_execution_time{quantile="0.75"} NaN
atlantis_builder_execution_time{quantile="0.95"} NaN
atlantis_builder_execution_time{quantile="0.99"} NaN
atlantis_builder_execution_time{quantile="0.999"} NaN
atlantis_builder_execution_time_sum 11.42403017
atlantis_builder_execution_time_count 10
.....
.....
.....
```

::: tip NOTE
The output shown above is trimmed, since with every new version release this metric set will need to be updated accordingly as there may be a case if some metrics are added/modified/deprecated, so the output shown above just gives a brief idea of how these metrics look like and rest can be explored.
:::

Important metrics to monitor are

| Metric Name                                | Metric Type                                                          | Purpose                                               |
|--------------------------------------------|----------------------------------------------------------------------|-------------------------------------------------------|
| `atlantis_project_plan_execution_success`  | [counter](https://prometheus.io/docs/concepts/metric_types/#counter) | number of times a project plan has run successfully.  |
| `atlantis_project_plan_execution_error`    | [counter](https://prometheus.io/docs/concepts/metric_types/#counter) | number of times a project plan has thrown an error.   |
| `atlantis_project_plan_execution_failure`  | [counter](https://prometheus.io/docs/concepts/metric_types/#counter) | number of times a project plan was not run.           |
| `atlantis_project_apply_execution_success` | [counter](https://prometheus.io/docs/concepts/metric_types/#counter) | number of times a project apply has run successfully. |
| `atlantis_project_apply_execution_error`   | [counter](https://prometheus.io/docs/concepts/metric_types/#counter) | number of times a project apply has thrown an error.  |
| `atlantis_project_apply_execution_failure` | [counter](https://prometheus.io/docs/concepts/metric_types/#counter) | number of times a project apply was not run.          |

An error means something went wrong while running the command, for example a workflow step such as `terraform plan` failed. A failure means Atlantis did not run the command, for example because the project is [locked](locking.md) by another pull request or a [plan or apply requirement](command-requirements.md) is not met.

These counters are reported per project and labelled with the pull request and project they ran for, for example `base_repo`, `pr_number` and `project_path`. The plan counters include both [autoplan](autoplanning.md#autoplanning) and `atlantis plan` comments. To see how long a whole command took, use the `atlantis_cmd_autoplan_execution_time`, `atlantis_cmd_comment_plan_execution_time` and `atlantis_cmd_comment_apply_execution_time` summaries.

::: tip NOTE
There are plenty of additional metrics exposed by atlantis that are not described above.
:::
