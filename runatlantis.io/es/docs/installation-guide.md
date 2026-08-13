# Guía de instalación

Esta guía es para instalar una instancia **lista para producción** de Atlantis en tu
infraestructura:

1. Primero, asegúrate de que tu configuración de Terraform cumpla los **requisitos** de Atlantis
    * Consulta [Requirements](requirements.md)
1. Crea **credenciales de acceso** para tu host de Git (GitHub, GitLab, Gitea, Bitbucket, Azure DevOps)
    * Consulta [Generating Git Host Access Credentials](access-credentials.md)
1. Crea un **webhook secret** para que Atlantis pueda validar webhooks
    * Consulta [Creating a Webhook Secret](webhook-secrets.md)
1. **Despliega** Atlantis en tu infraestructura
    * Consulta [Deployment](deployment.md)
1. Configura **Webhooks** en tu host de Git para que Atlantis pueda responder a tus pull requests
    * Consulta [Configuring Webhooks](configuring-webhooks.md)
1. Configura **credenciales del provider** para que Atlantis pueda ejecutar realmente comandos de Terraform
    * Consulta [Provider Credentials](provider-credentials.md)

:::tip
Si primero quieres probar Atlantis, revisa [Test Drive](../guide/test-drive.md)
y [Testing Locally](../guide/testing-locally.md).
:::
