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

## Bloqueo de todos los proyectos antes de planificar {#locking-all-projects-before-planning}

Por defecto, Atlantis bloquea cada proyecto justo antes de planificarlo, por lo que una
ejecución sobre muchos proyectos intercala bloqueo y planificación:

```plain
lock project A -> plan project A -> lock project B -> plan project B
```

En repositorios con mucha actividad y pull requests grandes, como un cambio de versión
de provider que afecta a todos los proyectos, esto deja una ventana abierta. Otro pull
request puede tomar el bloqueo de un proyecto al que aún no se ha llegado, a mitad de
una ejecución larga. Atlantis entonces falla en ese bloqueo y descarta los planes que ya
había producido.

Configurar [`repo_locks: {mode: on_apply}`](repo-level-atlantis-yaml.md#repolocks)
no resuelve esto: elimina el bloqueo de la planificación por completo, por lo que dos
pull requests pueden planificar el mismo proyecto al mismo tiempo, y solo descubren que
están en desacuerdo cuando uno de ellos aplica. `--lock-all-projects-before-plan`
mantiene la garantía por defecto de que solo un pull request puede estar planificando un
proyecto dado a la vez. Solo cambia _cuándo_ se toma el bloqueo de cada proyecto, no si
se toma.

Iniciar Atlantis con
[`--lock-all-projects-before-plan`](server-configuration.md#-lock-all-projects-before-plan)
adquiere todos los bloqueos por adelantado en su lugar:

```plain
lock project A -> lock project B -> plan project A -> plan project B
```

Si no se puede adquirir algún bloqueo, no se ejecuta ningún plan, y los bloqueos que esta
ejecución ya había tomado se liberan de nuevo, de modo que un pull request competidor
nunca queda bloqueado por una ejecución que se rindió. El comentario del pull request
nombra el proyecto que quedó bloqueado y quién tiene su bloqueo, igual que hoy.

Notas:

* Los proyectos configurados con `repo_locks: {mode: on_apply}` o `mode: disabled` no
  se bloquean por adelantado.
* Los proyectos que terminan sin producir un plan tienen su bloqueo liberado cuando la
  ejecución termina, de modo que el bloqueo por adelantado nunca deja un proyecto
  bloqueado sin nada que aplicar. Esto cubre una ejecución detenida con `atlantis cancel`,
  y un grupo de orden de ejecución anterior que falló con `abort_on_execution_order_fail`.
* Si el propio servidor de Atlantis muere a mitad de la ejecución, los bloqueos ya
  adquiridos permanecen hasta que se cierra el pull request o alguien ejecuta
  `atlantis unlock`. Esa es la misma vía de recuperación que cualquier otra ejecución
  interrumpida, pero el bloqueo por adelantado hace que afecte a más proyectos a la vez.

## Relación con Terraform State Locking

Atlantis no entra en conflicto con [Terraform State Locking](https://developer.hashicorp.com/terraform/language/state/locking). Internamente, todo lo que
hace Atlantis es ejecutar `terraform plan` y `apply`, por lo que todo el
bloqueo incorporado en esos comandos por Terraform no se ve afectado.

En más detalle, Terraform state locking bloquea el estado mientras ejecutas `terraform apply`
para que múltiples applies no puedan ejecutarse de forma concurrente. El bloqueo de Atlantis está en un nivel más alto
porque evita que múltiples pull requests trabajen sobre el mismo estado.

## Bloqueo y detección de drift

Cuando la detección o remediación de drift se ejecuta mediante la [API](api-endpoints.md) con `PR: 0` (flujo de trabajo no PR), Atlantis igualmente adquiere y libera bloqueos del directorio de trabajo para evitar operaciones concurrentes en el mismo proyecto. Sin embargo, dado que estas operaciones no están asociadas con un pull request, no crean bloqueos a nivel de PR visibles en la UI de Locks. El bloqueo del directorio de trabajo se libera automáticamente después de que cada operación de detección o remediación de drift se completa.
