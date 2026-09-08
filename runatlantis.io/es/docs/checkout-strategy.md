# Estrategia de checkout

Puede configurar cómo Atlantis hace checkout del código de su pull request mediante
the `--checkout-strategy` flag or the `ATLANTIS_CHECKOUT_STRATEGY` environment
variable that get passed to the `atlantis server` command.

Atlantis soporta las estrategias `branch` y `merge`.

## Branch

Si se establece en `branch` (el valor predeterminado), Atlantis hará checkout de la rama de origen
del pull request.

Por ejemplo, dado el siguiente historial de git:

![Git History](../../docs/images/branch-strategy.png)

Si el pull request estuviera solicitando fusionar `branch` en `main`,
Atlantis haría checkout de `branch` en el commit `C3`.

## Merge

El problema con la estrategia `branch` es que, si los usuarios hacen push de ramas que están
desactualizadas con `main`, entonces su `terraform plan` podría estar eliminando
algunos recursos que fueron configurados en la rama principal.

Por ejemplo, en el diagrama anterior, si los commits `C4` y `C5` han modificado el
estado de Terraform y agregado nuevos recursos, entonces cuando Atlantis ejecuta `terraform plan`
en el commit `C3`, debido a que el código no tiene los cambios de `C4` y `C5`,
Terraform intentará eliminar esos recursos.

Para solucionar esto, los usuarios podrían fusionar `main` en su rama, *o* puede ejecutar
Atlantis con `--checkout-strategy=merge`. Con esta estrategia, Atlantis
intentará realizar una fusión localmente de la siguiente manera:

* Haciendo checkout de la rama de destino del pull request (por ej. `main`)
* Realizando localmente un `git merge {source branch}`
* Luego ejecutando sus comandos de Terraform

En este ejemplo, el código sobre el que Atlantis estaría operando se vería así:

![Git History](../../docs/images/merge-strategy.png)

Donde Atlantis está usando su commit local `C6`.

:::tip NOTE
Atlantis en realidad no hace commit de esta fusión en ningún lugar. Solo la usa localmente.
:::

:::tip NOTE
En el caso de errores transitorios al actualizar la rama fusionada, Atlantis
producirá un error por seguridad para evitar usar una rama obsoleta.
:::

:::tip NOTE
Cuando Atlantis está autenticado como una GitHub App y usa la estrategia de checkout `merge`
para un pull request de GitHub, obtiene la cabecera del pull request desde la
referencia `pull/<PR number>/head` del repositorio base en `origin`. No crea ni
actualiza un remote separado `source` para el repositorio de cabecera del pull request. Esto
evita obtener forks a los que la instalación de GitHub App no puede acceder. La
ruta de checkout que no es de GitHub App todavía usa el remote `source`.
:::

:::warning
Atlantis solo realiza esta fusión durante la fase de `terraform plan`. Si otro
commit es enviado a `main` **después** de que Atlantis ejecute `plan`, no ocurrirá nada.
:::

Para optimizar el tiempo de clonación, Atlantis puede realizar un clon superficial especificando la bandera `--checkout-depth`. La clonación se realiza de la siguiente manera:

* Se realiza un clon superficial de la rama predeterminada con una profundidad del valor de `--checkout-depth` de cero (clon completo).
* Se recupera `branch`, incluyendo la misma cantidad de commits.
* Se comprueba la existencia de la base de fusión de la rama predeterminada e `branch` en el clon superficial.
* Si la base de fusión no está presente, significa que cualquiera de las ramas está por delante de la base de fusión por más de `--checkout-depth` commits. En este caso se obtiene el historial completo del repositorio.

Si el historial de commits a menudo diverge por más que la profundidad de checkout predeterminada, entonces la bandera `--checkout-depth` debe ajustarse para evitar recuperaciones completas.
