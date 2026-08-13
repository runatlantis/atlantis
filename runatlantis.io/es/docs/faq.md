# Preguntas frecuentes

**P: ¿Atlantis afecta al [remote state](https://developer.hashicorp.com/terraform/language/state/remote) de Terraform?**

R: No. Atlantis no interfiere con el remote state de Terraform de ninguna manera. Internamente, Atlantis simplemente ejecuta `terraform plan` e `terraform apply`.

**P: ¿Cómo interactúa el locking de Atlantis con el [locking](https://developer.hashicorp.com/terraform/language/state/locking) de Terraform?**

R: Atlantis proporciona locking de pull requests que evita la modificación concurrente de la misma infraestructura (proyecto de Terraform), mientras que el locking de Terraform solo evita que ocurran dos `terraform apply` concurrentes.

El locking de Terraform puede usarse junto con el locking de Atlantis, ya que Atlantis simplemente ejecuta comandos de Terraform.

**P: ¿Cómo ejecutar Atlantis en modo de alta disponibilidad? ¿Es necesario?**

R: El servidor Atlantis puede ejecutarse fácilmente bajo la supervisión de un sistema init como `upstart` o `systemd` para asegurar que `atlantis server` esté siempre en ejecución.

Atlantis, de forma predeterminada, almacena todo el locking y los planes de Terraform localmente en disco en el directorio `--data-dir` (el valor predeterminado es `~/.atlantis`). Si se ejecutan múltiples hosts de Atlantis utilizando un backend redis compartido, entonces es importante que `data-dir` esté usando un sistema de archivos compartido entre hosts.

Sin embargo, si perdieras los datos, todo lo que necesitarías hacer es ejecutar `atlantis plan` de nuevo en los pull requests que están abiertos. Si alguien intenta ejecutar `atlantis apply` después de que los datos se hayan perdido, entonces recibirá un error, por lo que tendrá que volver a hacer plan de todos modos.

Para HA completamente sin estado (sin sistema de archivos compartido ni volumen persistente), Atlantis admite almacenamiento externo de planes mediante `--enable-external-stores` con un backend compatible con S3. Los planes se suben a S3 después de `terraform plan` y se restauran automáticamente cuando una réplica diferente maneja el `apply`. Combinado con Redis para locking (`--locking-db-type=redis`), esto permite ejecutar múltiples réplicas de Atlantis con volúmenes `emptyDir`. Consulta [server configuration](server-configuration.md) para más detalles.

**P: ¿Cómo agregar SSL al servidor Atlantis?**

R: Primero, necesitarás obtener un par de claves pública/privada para servir sobre SSL.
Estas deben estar en un directorio accesible por Atlantis. Luego inicia `atlantis server` con los flags `--ssl-cert-file` e `--ssl-key-file`.
Consulta `atlantis server --help` para más información.

**P: ¿Atlantis puede detectar drift de infraestructura?**

R: Sí. Cuando el flag `--enable-drift-detection` está configurado, Atlantis expone endpoints de API para detección de drift, estado y remediación. La detección de drift funciona ejecutando `terraform plan` contra la rama/ref especificada (fuera del contexto de un pull request) y analizando la salida del plan en busca de cambios. Puedes activar la detección de drift mediante `POST /api/drift/detect` y recuperar resultados en caché mediante `GET /api/drift/status`. Consulta [API Endpoints](api-endpoints.md) para más detalles.

**P: ¿Cómo configuro la detección de drift programada?**

R: Atlantis proporciona la API de detección de drift, pero no incluye un scheduler incorporado. Puedes usar un scheduler externo (p. ej., cron, pipelines de CI/CD o una herramienta de monitoreo) para llamar periódicamente a `POST /api/drift/detect`. Configura [drift webhooks](sending-notifications-via-webhooks.md) para recibir notificaciones de Slack o HTTP cuando se detecte drift.

**P: ¿Cómo puedo poner Atlantis en funcionamiento en AWS?**

R: Existe el proyecto [terraform-aws-atlantis](https://github.com/terraform-aws-modules/terraform-aws-atlantis), donde se alojan configuraciones completas de Terraform para ejecutar Atlantis en AWS Fargate. Probado y mantenido.
