# Secretos de Webhook

Atlantis usa secretos de Webhook para validar que los webhooks que recibe de tu
host de Git son legítimos.

Una forma de confirmar esto sería permitir en la allowlist solicitudes
que solo vengan desde las IPs de tu host de Git, pero una forma más fácil es usar un secreto de Webhook.

::: tip NOTE
Los secretos de webhook en realidad son opcionales. Sin embargo, son altamente recomendados por
seguridad.
:::

::: tip NOTE
Azure DevOps usa autenticación Basic para webhooks en lugar de secretos de webhook.
:::

::: tip NOTE
Se genera un token para toda la app durante la [configuración de GitHub App](access-credentials.md#github-app). Puedes recuperarlo navegando a la [página de configuración de la app de GitHub](https://github.com/settings/apps) y seleccionando "Edit" junto al nombre de tu app de Atlantis. El token aparece después de hacer clic en "Edit" bajo el encabezado Webhook.
:::

## Generar un secreto de Webhook

Puedes usar cualquier generador de cadenas aleatorias para crear tu secreto de Webhook. Debe tener > 24 caracteres.

Por ejemplo:

* Generar mediante Ruby con `ruby -rsecurerandom -e 'puts SecureRandom.hex(32)'`
* Generar en línea con [browserling: Generate Random Strings and Numbers](https://www.browserling.com/tools/random-string)

::: tip NOTE
Debes usar **el mismo** secreto de webhook para cada repo.
:::

## Siguientes pasos

* Registra tu secreto
* Lo usarás más tarde para [configurar tus webhooks](configuring-webhooks.md), sin embargo, si estás
siguiendo la [Guía de instalación](installation-guide.md), entonces tu siguiente paso es
[Desplegar Atlantis](deployment.md)
