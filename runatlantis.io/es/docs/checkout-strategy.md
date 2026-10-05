# Estrategia de Checkout

Puede configurar cómo Atlantis hace checkout del código de su pull request mediante
la bandera `--checkout-strategy` o la variable de entorno `ATLANTIS_CHECKOUT_STRATEGY`
que se pasan al comando `atlantis server`.

Atlantis soporta las estrategias `branch` e `merge`.

## Branch

Si se establece en `branch` (la predeterminada), Atlantis hará checkout de la rama fuente
del pull request.

Por ejemplo, dado el siguiente historial de git:

![Git History](../../docs/images/branch-strategy.png)

Si el pull request pidiera fusionar `branch` en `main`,
Atlantis haría checkout de `branch` en el commit `C3`.

## Merge

El problema con la estrategia `branch` es que, si los usuarios envían ramas que están
desactualizadas con respecto a `main`, entonces su `terraform plan` podría eliminar
algunos recursos que fueron configurados en la rama principal.

Por ejemplo, en el diagrama anterior, si los commits `C4` e `C5` han modificado el
estado de terraform y agregado nuevos recursos, entonces cuando Atlantis ejecute `terraform plan`
en el commit `C3`, como el código no tiene los cambios de `C4` y `C5`,
Terraform intentará eliminar esos recursos.

Para corregir esto, los usuarios podrían fusionar `main` en su rama, *o* puede ejecutar
Atlantis con `--checkout-strategy=merge`. Con esta estrategia, Atlantis
intentará realizar una fusión localmente mediante:

* Hacer checkout de la rama de destino del pull request (p. ej. `main`)
* Realizar localmente un `git merge {source branch}`
* Luego ejecutar sus comandos de Terraform

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
Cuando Atlantis se autentica como una GitHub App y usa la estrategia de checkout `merge`
para un pull request de GitHub, obtiene el head del pull request desde la
ref `pull/<PR number>/head` del repositorio base en `origin`. No crea ni
actualiza un remote `source` separado para el repositorio head del pull request. Esto
evita obtener forks a los que la instalación de GitHub App no puede acceder. La
ruta de checkout que no es GitHub App todavía usa el remote `source`.
:::

:::warning
Atlantis solo realiza esta fusión durante la fase `terraform plan`. Si otro
commit se envía a `main` **después** de que Atlantis ejecute `plan`, no sucederá nada.
:::

Para optimizar el tiempo de clonación, Atlantis puede realizar un clone superficial especificando la bandera `--checkout-depth`. La clonación se realiza de la siguiente manera:

* Se realiza un clone superficial de la rama predeterminada con una profundidad igual al valor de `--checkout-depth` de cero (clone completo).
* Se recupera `branch`, incluyendo la misma cantidad de commits.
* Se verifica la existencia de la base de fusión de la rama predeterminada e `branch` en el clone superficial.
* Si la base de fusión no está presente, significa que alguna de las ramas está por delante de la base de fusión por más de `--checkout-depth` commits. En este caso se recupera el historial completo del repositorio.

Si el historial de commits a menudo diverge por más de la profundidad de checkout predeterminada, entonces la bandera `--checkout-depth` debe ajustarse para evitar recuperaciones completas.
