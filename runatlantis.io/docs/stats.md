# Metrics/Stats

Atlantis exposes a set of metrics for each of its operations including errors, successes, and latencies.

::: warning NOTE
Only statsd is supported currently, but it should be relatively straightforward to add other providers such as prometheus.
:::

## Configuration

Metrics are configured through the (server-side-repo-config.html#Metrics). 