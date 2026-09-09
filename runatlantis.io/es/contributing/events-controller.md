# Controlador de eventos

Los Webhooks son la interacción principal entre el Version Control System (VCS)
y Atlantis. Cada VCS envía las solicitudes al endpoint `/events`. La
implementación de este endpoint se puede encontrar en el archivo
[events_controller.go](https://github.com/runatlantis/atlantis/blob/main/server/controllers/events/events_controller.go).
Este archivo contiene la función Post `func (e *VCSEventsController)
Post(w http.ResponseWriter, r *http.Request`)` que analiza la solicitud
de acuerdo con el VCS configurado.

Atlantis actualmente maneja uno de los siguientes eventos:

- Evento de comentario
- Evento de Pull Request

Todos los demás eventos son ignorados.

```mermaid
---
title: events controller flowchart
---
flowchart LR
    events(/events - Endpoint) --> Comment_Event(Comment - Event)
    events --> Pull_Request_Event(Pull Request - Event)

    Comment_Event --> pre_workflow(pre-workflow - Hook)
    pre_workflow --> plan(plan - command)
    pre_workflow --> apply(apply - command)
    pre_workflow --> approve_policies(approve policies - command)
    pre_workflow --> unlock(unlock - command)
    pre_workflow --> version(version - command)
    pre_workflow --> import(import - command)
    pre_workflow --> state(state - command)

    plan --> post_workflow(post-workflow - Hook)
    apply --> post_workflow
    approve_policies --> post_workflow
    unlock --> post_workflow
    version --> post_workflow
    import --> post_workflow
    state --> post_workflow

    Pull_Request_Event --> Open_Update_PR(Open / Update Pull Request)
    Pull_Request_Event --> Close_PR(Close Pull Request)

    Open_Update_PR --> pre_workflow(pre-workflow - Hook)
    Close_PR --> plan(plan - command)

    pre_workflow --> plan
    plan --> post_workflow(post-workflow - Hook)

    Close_PR --> CleanUpPull(CleanUpPull)
    CleanUpPull --> post_workflow(post-workflow - Hook)
```

## Evento de comentario

Este evento se activa cada vez que un usuario ingresa un comentario en el Pull Request,
Merge Request, o como sea que se llame para el VCS respectivo. Después de analizar la
solicitud específica del VCS, el código llama a la función `handleCommentEvent`, que
luego pasa el procesamiento a la función `handleCommentEvent` en el archivo
[command_runner.go](https://github.com/runatlantis/atlantis/blob/main/server/events/command_runner.go).
Esta función primero llama a los hooks pre-workflow, luego ejecuta uno de los
comandos listados abajo y, al final, los hooks post-workflow.

- [plan_command_runner.go](https://github.com/runatlantis/atlantis/blob/main/server/events/plan_command_runner.go)
- [apply_command_runner.go](https://github.com/runatlantis/atlantis/blob/main/server/events/apply_command_runner.go)
- [approve_policies_command_runner.go](https://github.com/runatlantis/atlantis/blob/main/server/events/approve_policies_command_runner.go)
- [unlock_command_runner.go](https://github.com/runatlantis/atlantis/blob/main/server/events/unlock_command_runner.go)
- [version_command_runner.go](https://github.com/runatlantis/atlantis/blob/main/server/events/version_command_runner.go)
- [import_command_runner.go](https://github.com/runatlantis/atlantis/blob/main/server/events/import_command_runner.go)
- [state_command_runner.go](https://github.com/runatlantis/atlantis/blob/main/server/events/state_command_runner.go)

## Evento de Pull Request

Para manejar eventos de comentario en Pull Requests, primero deben ser creados. Atlantis
también permite la ejecución de comandos para ciertos eventos de Pull Requests.

<details>
  <summary>Webhooks de Pull Request</summary>

La lista de abajo enlaza a los VCS soportados y a su documentación de Webhook
de Pull Request.

- [Azure DevOps Pull Request Created](https://learn.microsoft.com/en-us/azure/devops/service-hooks/events?view=azure-devops#pull-request-created)
- [BitBucket Pull Request](https://support.atlassian.com/bitbucket-cloud/docs/event-payloads/#Pull-request-events)
- [GitHub Pull Request](https://docs.github.com/en/webhooks/webhook-events-and-payloads#pull_request)
- [GitLab Merge Request](https://docs.gitlab.com/user/project/integrations/webhook_events/#merge-request-events)
- [Gitea Webhooks](https://docs.gitea.com/usage/webhooks)

</details>

La siguiente lista muestra los eventos soportados:

- Pull Request abierto
- Pull Request actualizado
- Pull Request cerrado
- Otro evento de Pull Request

La función `RunAutoPlanCommand` en el archivo
[command_runner.go](https://github.com/runatlantis/atlantis/blob/main/server/events/command_runner.go)
es llamada para los eventos de Pull Request _Open_ y _Update_. Cuando está habilitada en
el proyecto, esto ejecuta automáticamente el `plan` para el repositorio específico.

Siempre que un Pull Request se cierra, la función `CleanUpPull` en el archivo
[instrumented_pull_closed_executor.go](https://github.com/runatlantis/atlantis/blob/main/server/events/instrumented_pull_closed_executor.go)
es llamada. Esta función limpia todos los archivos, locks y otra información relacionada
del Pull Request cerrado.
