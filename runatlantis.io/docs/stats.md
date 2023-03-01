# Metrics/Stats

Atlantis exposes a set of metrics for each of its operations including errors, successes, and latencies.

::: warning NOTE
Currently Statsd and Prometheus is supported. See configuration below for details.
:::

## Configuration

Metrics are configured through the [Server Side Config](server-side-repo-config.html#metrics).

## Available Metrics

Assuming metrics are exposed from the endpoint `/metrics` from the [metrics](server-side-repo-config.html#metrics) server side config e.g.


```yaml
metrics:
  prometheus:
    endpoint: "/metrics"
```


To see all the metrics exposed from atlantis service, make a GET request to the `/metrics` endpoint.


```bash
curl localhost:4141/metrics
# HELP atlantis_cmd_autoplan_builder_execution_error atlantis_cmd_autoplan_builder_execution_error counter
# TYPE atlantis_cmd_autoplan_builder_execution_error counter
atlantis_cmd_autoplan_builder_execution_error 0
# HELP atlantis_cmd_autoplan_builder_execution_success atlantis_cmd_autoplan_builder_execution_success counter
# TYPE atlantis_cmd_autoplan_builder_execution_success counter
atlantis_cmd_autoplan_builder_execution_success 10
# HELP atlantis_cmd_autoplan_builder_execution_time atlantis_cmd_autoplan_builder_execution_time summary
# TYPE atlantis_cmd_autoplan_builder_execution_time summary
atlantis_cmd_autoplan_builder_execution_time{quantile="0.5"} NaN
atlantis_cmd_autoplan_builder_execution_time{quantile="0.75"} NaN
atlantis_cmd_autoplan_builder_execution_time{quantile="0.95"} NaN
atlantis_cmd_autoplan_builder_execution_time{quantile="0.99"} NaN
atlantis_cmd_autoplan_builder_execution_time{quantile="0.999"} NaN
atlantis_cmd_autoplan_builder_execution_time_sum 11.42403017
atlantis_cmd_autoplan_builder_execution_time_count 10
.....
.....
.....
```


::: tip NOTE
The output shown above is trimmed, since with every new version release this metric set will need to be updated accordingly as there may be a case if some metrics are added/modified/deprecated, so the output shown above just gives a brief idea of how these metrics look like and rest can be explored.
:::

Important metrics to monitor are

| Metric Name                                    | Metric Type                                                          | Purpose                                                                                                            |
|------------------------------------------------|----------------------------------------------------------------------|--------------------------------------------------------------------------------------------------------------------|
| `atlantis_cmd_autoplan_execution_error`        | [counter](https://prometheus.io/docs/concepts/metric_types/#counter) | number of times when [autoplan](autoplanning.html#autoplanning) has thrown error. |
| `atlantis_cmd_comment_plan_execution_error`    | [counter](https://prometheus.io/docs/concepts/metric_types/#counter) | number of times when on commenting `atlantis plan` has thrown error.      |
| `atlantis_cmd_autoplan_execution_success`      | [counter](https://prometheus.io/docs/concepts/metric_types/#counter) | number of times when [autoplan](autoplanning.html#autoplanning) has run successfully. |
| `atlantis_cmd_comment_apply_execution_error`   | [counter](https://prometheus.io/docs/concepts/metric_types/#counter) | number of times when on commenting `atlantis apply` has thrown error.     |
| `atlantis_cmd_comment_apply_execution_success` | [counter](https://prometheus.io/docs/concepts/metric_types/#counter) | number of times when on commenting `atlantis apply` has run successfully. |

::: tip NOTE
There are plenty of additional metrics exposed by atlantis that are not described above.
:::
