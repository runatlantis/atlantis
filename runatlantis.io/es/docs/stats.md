# Métricas/Estadísticas

Atlantis expone un conjunto de métricas para cada una de sus operaciones, incluidos errores, éxitos y latencias.

Las métricas de latencia y duración se informan en segundos (por ejemplo, las métricas `*_time` en la salida de Prometheus de abajo), mientras que las métricas de contador, como éxitos y errores, no tienen unidad.

::: warning NOTE
Actualmente, Statsd y Prometheus son compatibles. Vea la configuración de abajo para más detalles.
:::

## Configuración

Las métricas se configuran mediante [Server Side Config](server-side-repo-config.md#metrics).

## Métricas disponibles

Suponiendo que las métricas se exponen desde el endpoint `/metrics` de la configuración del lado del servidor [metrics](server-side-repo-config.md#metrics), p. ej.

```yaml
metrics:
  prometheus:
    endpoint: "/metrics"
```

Para ver todas las métricas expuestas desde el servicio atlantis, haga una solicitud GET al endpoint `/metrics`.

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
La salida mostrada arriba está recortada, ya que con cada nueva versión publicada este conjunto de métricas deberá actualizarse en consecuencia, porque puede darse el caso de que algunas métricas se agreguen/modifiquen/queden obsoletas, por lo que la salida mostrada arriba solo da una idea breve de cómo se ven estas métricas y el resto se puede explorar.
:::

Las métricas importantes para monitorear son

| Nombre de la métrica                           | Tipo de métrica                                                      | Propósito                                                                                   |
| ---------------------------------------------- | -------------------------------------------------------------------- | ------------------------------------------------------------------------------------------- |
| `atlantis_cmd_autoplan_execution_error`        | [counter](https://prometheus.io/docs/concepts/metric_types/#counter) | número de veces que [autoplan](autoplanning.md#autoplanning) ha producido un error.         |
| `atlantis_cmd_comment_plan_execution_error`    | [counter](https://prometheus.io/docs/concepts/metric_types/#counter) | número de veces que al comentar `atlantis plan` se ha producido un error.                   |
| `atlantis_cmd_autoplan_execution_success`      | [counter](https://prometheus.io/docs/concepts/metric_types/#counter) | número de veces que [autoplan](autoplanning.md#autoplanning) se ha ejecutado correctamente. |
| `atlantis_cmd_comment_apply_execution_error`   | [counter](https://prometheus.io/docs/concepts/metric_types/#counter) | número de veces que al comentar `atlantis apply` se ha producido un error.                  |
| `atlantis_cmd_comment_apply_execution_success` | [counter](https://prometheus.io/docs/concepts/metric_types/#counter) | número de veces que al comentar `atlantis apply` se ha ejecutado correctamente.             |

::: tip NOTE
Hay muchas métricas adicionales expuestas por atlantis que no se describen arriba.
:::
