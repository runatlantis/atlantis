# Introducción

## Primeros pasos

* Si te gustaría solo probar ejecutar Atlantis en un **repo de ejemplo**, consulta [Test Drive](./guide/test-drive.md).
* Si te gustaría probar ejecutar Atlantis en **tus repos**, entonces lee [Testing Locally](./guide/testing-locally.md).
* Si estás listo para instalar Atlantis correctamente en infraestructura real, entonces dirígete a la [Installation Guide](./docs/installation-guide.md).

::: tip ¿Buscas la documentación completa?
Ve aquí: [www.runatlantis.io/docs](./docs.md)
:::

## Descripción general – ¿Qué es Atlantis?

Atlantis es una aplicación para automatizar Terraform mediante pull requests. Se despliega
como una aplicación independiente en tu infraestructura. Ningún tercero tiene acceso a
tus credenciales.

Atlantis escucha webhooks de GitHub, GitLab o Bitbucket sobre pull requests de Terraform. Luego
ejecuta `terraform plan` y comenta con la salida de vuelta en el pull request.

Cuando quieres aplicar, comenta `atlantis apply` en el pull request y Atlantis
ejecutará `terraform apply` y responderá con la salida.

## Ver

Mira el video a continuación para verlo en acción:

[![Atlantis Walkthrough](../guide/images/atlantis-walkthrough-icon.png)](https://www.youtube.com/watch?v=TmIPWda0IKg)

## ¿Por qué ejecutarías Atlantis?

### Mayor visibilidad

Cuando todos ejecutan Terraform en sus propias computadoras, es difícil conocer el
estado actual de tu infraestructura:

* ¿Está desplegado lo que está en la rama `main`?
* ¿Alguien olvidó crear un pull request para ese último cambio?
* ¿Cuál fue la salida de ese último `terraform apply`?

Con Atlantis, todo es visible en el pull request. Puedes ver el historial
de todo lo que se hizo en tu infraestructura.

### Habilitar colaboración con todos

Probablemente no quieras distribuir credenciales de Terraform a todos en tu
organización de ingeniería, pero ahora cualquiera puede abrir un pull request de Terraform.

Puedes requerir aprobación antes de que el pull request sea aplicado para que no ocurra nada
accidentalmente.

### Revisar mejor los pull requests de Terraform

No puedes revisar completamente un cambio de Terraform sin ver la salida de `terraform plan`.
Ahora esa salida se agrega al pull request automáticamente.

### Estandarizar tus workflows

Atlantis bloquea un directorio/workspace hasta que el pull request se fusiona o el bloqueo
se elimina manualmente. Esto asegura que los cambios se apliquen en el orden esperado.

Los comandos exactos que ejecuta Atlantis son configurables. Puedes ejecutar scripts personalizados
para construir tu workflow ideal.

## Próximos pasos

* Si te gustaría solo probar ejecutar Atlantis en un **repo de ejemplo**, consulta [Test Drive](./guide/test-drive.md).
* Si te gustaría probar ejecutar Atlantis en **tus repos**, entonces lee [Testing Locally](./guide/testing-locally.md).
* Si estás listo para instalar Atlantis correctamente en infraestructura real, entonces dirígete a la [Installation Guide](./docs/installation-guide.md).
