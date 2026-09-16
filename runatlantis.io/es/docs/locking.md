# Bloqueo

Cuando se ejecuta `plan`, el directorio y el workspace de Terraform quedan **bloqueados** hasta que el pull request se fusione o se cierre, o el plan se elimine manualmente.

Si otro usuario intenta `plan` para el mismo directorio y workspace en un pull request diferente,
verá este error:

![Lock Comment](../../docs/images/lock-comment.png)

Lo cual lo enlaza al pull request que mantiene el bloqueo.

::: warning NOTE
Solo se bloquean el directorio en el repo y el workspace de Terraform, no todo el repo.
:::

Atlantis también verifica el bloqueo global de apply antes de ejecutar `atlantis apply`. Si Atlantis no puede alcanzar el backend de bloqueo mientras verifica ese bloqueo global, falla en cerrado y rechaza el apply hasta que el backend vuelva a estar accesible.

## Por qué

1. Debido a que `atlantis apply` se realiza antes de que el pull request se fusione, después de
un apply tu rama `main` ya no representa la versión más actualizada de tu infraestructura.
Con el bloqueo, puedes asegurar que no se harán otros cambios hasta que el
pull request se fusione.

::: tip Por qué no hacer apply al fusionar?
A veces `terraform apply` falla. Si el apply fallara después de que el pull
request se hubiera fusionado, necesitarías crear un nuevo pull request para corregirlo.
Con bloqueo + aplicar en la rama, efectivamente imitas la fusión a main
pero con la capacidad añadida de volver a hacer plan/apply múltiples veces si las cosas no funcionan.
:::
2. Si ya hay un `plan` en progreso, otros usuarios no verán un plan que
quedará invalidado después de que se aplique el plan en progreso.

## Ver bloqueos

Para ver los bloqueos, ve a la URL en la que está alojado Atlantis:

![Locks View](../../docs/images/locks-ui.png)

Puedes hacer clic en un bloqueo para ver sus detalles:

<p align="center">
    <img src="../../docs/images/lock-detail-ui.png" alt="Lock Detail View" height="400px">
</p>

## Desbloqueo

El proyecto y el workspace se desbloquearán automáticamente cuando el PR se fusione o se cierre.

Para desbloquear el proyecto y el workspace sin completar un `apply` y fusionar, comenta `atlantis unlock` en el PR,
o haz clic en el enlace en la parte inferior del comentario del plan para descartar el plan y eliminar el bloqueo donde
dice **"To discard this plan click here"**:

![Locks View](../../docs/images/lock-delete-comment.png)

El enlace te llevará a la vista de detalles del bloqueo donde puedes hacer clic en **Discard Plan and Unlock**
para eliminar el bloqueo.

<p align="center">
    <img src="../../docs/images/lock-detail-ui.png" alt="Lock Detail View" height="400px">
</p>

Una vez que se descarta un plan, necesitarás ejecutar `plan` de nuevo antes de ejecutar `apply` cuando regreses a ese pull request.

## Relación con Terraform State Locking

Atlantis no entra en conflicto con [Terraform State Locking](https://developer.hashicorp.com/terraform/language/state/locking). Internamente, todo lo que
hace Atlantis es ejecutar `terraform plan` y `apply`, por lo que todo el
bloqueo incorporado en esos comandos por Terraform no se ve afectado.

En más detalle, Terraform state locking bloquea el estado mientras ejecutas `terraform apply`
para que múltiples applies no puedan ejecutarse de forma concurrente. El bloqueo de Atlantis está en un nivel más alto
porque evita que múltiples pull requests trabajen sobre el mismo estado.

## Bloqueo y detección de drift

Cuando la detección o remediación de drift se ejecuta mediante la [API](api-endpoints.md) con `PR: 0` (flujo de trabajo no PR), Atlantis igualmente adquiere y libera bloqueos del directorio de trabajo para evitar operaciones concurrentes en el mismo proyecto. Sin embargo, dado que estas operaciones no están asociadas con un pull request, no crean bloqueos a nivel de PR visibles en la UI de Locks. El bloqueo del directorio de trabajo se libera automáticamente después de que cada operación de detección o remediación de drift se completa.
