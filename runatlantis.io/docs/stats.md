# Metrics/Stats

Atlantis exposes a set of metrics for each of its operations including errors, successes, and latencies.

::: warning NOTE
Currently Statsd and Prometheus is supported. See configuration below for details.
:::

## Configuration

Metrics are configured through the [Server Side Config](server-side-repo-config.html#metrics).

## Available Metrics/Stats

There are plenty of metrics which are exposed by atlantis service. Let's say you have exposed metrics from a endpoint `/metrics` from your [Server Side Config](server-side-repo-config.html#metrics)
```
metrics:
  prometheus:
    endpoint: "/metrics"
```
then on making curl request to your atlantis server on path `/metrics` will list down all the metrics exposed from atlantis service.
```
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
Some of the important metrics which you would like to monitor are

| Metric Name                                  | Metric Type                                                          | Purpose                                                                                                                                  |
|----------------------------------------------|----------------------------------------------------------------------|------------------------------------------------------------------------------------------------------------------------------------------|
| atlantis_cmd_autoplan_execution_error        | [counter](https://prometheus.io/docs/concepts/metric_types/#counter) | since starting of atlantis server, the number of times when [autoplan](autoplanning.html#autoplanning) on the MR’s has thrown error.     |
| atlantis_cmd_comment_plan_execution_error    | [counter](https://prometheus.io/docs/concepts/metric_types/#counter) | since starting of atlantis server, the number of times when on commenting `atlantis plan` on the MR’s has thrown error.                  |
| atlantis_cmd_autoplan_execution_success      | [counter](https://prometheus.io/docs/concepts/metric_types/#counter) | since starting of atlantis server, the number of times when [autoplan](autoplanning.html#autoplanning) on the MR’s has run successfully. |
| atlantis_cmd_comment_apply_execution_error   | [counter](https://prometheus.io/docs/concepts/metric_types/#counter) | since starting of atlantis server, the number of times when on commenting `atlantis apply` on the MR’s has thrown error.                 |
| atlantis_cmd_comment_apply_execution_success | [counter](https://prometheus.io/docs/concepts/metric_types/#counter) | since starting of atlantis server, the number of times when on commenting `atlantis apply` on the MR’s has run successfully.             |

Since there are a plenty of metrics exposed by atlantis, so you can go through them all and çan monitor the one's which are well suited for you.
