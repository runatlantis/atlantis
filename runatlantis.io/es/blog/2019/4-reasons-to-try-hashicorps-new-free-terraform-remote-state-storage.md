---
title: 4 razones para probar el (nuevo) almacenamiento remoto de state gratuito
  de Terraform de HashiCorp
lang: en-US
---

# 4 razones para probar el (nuevo) almacenamiento remoto de state gratuito de Terraform de HashiCorp

::: info
Esta publicación fue escrita originalmente el 2 de abril de 2019

Publicación original: <https://medium.com/runatlantis/4-reasons-to-try-hashicorps-new-free-terraform-remote-state-storage-b03f01bfd251>
:::

Actualización (20 de mayo de 2019) — Free State Storage ahora se llama Terraform Cloud y salió de Beta, lo que significa que cualquiera puede registrarse.

HashiCorp está planeando ofrecer almacenamiento remoto de state gratuito para Terraform y tienen una versión beta disponible ahora. En este artículo, hablo de 4 razones por las que deberías probarlo (Divulgación: trabajo en HashiCorp).

> _[Regístrate en Terraform Cloud](https://goo.gl/X5t5EM)._

## ¿Qué es Terraform State?

Antes de entrar en por qué deberías usar el nuevo almacenamiento remoto de state, hablemos de qué queremos decir exactamente con state en Terraform.

Terraform usa _state_ para mapear tu código de Terraform a los recursos del mundo real que aprovisiona. Por ejemplo, si tengo código de Terraform para crear una instancia EC2 de AWS:

```tf
resource "aws_instance" "web" {
  ami           = "ami-e6d9d68c"
  instance_type = "t2.micro"
}
```

Cuando ejecuto `terraform apply`, Terraform hará una llamada a la API de AWS de “crear instancia EC2” y AWS devolverá el ID único de esa instancia (p. ej. `i-0ad17607e5ee026d0`). Terraform necesita registrar ese ID en algún lugar para que más tarde pueda hacer llamadas a la API para cambiar o eliminar la instancia.

Para almacenar esta información, Terraform usa un archivo de state. Para el código anterior, el archivo de state se verá algo así:

```json{4,7}
{
    ...
    "resources": {
      "aws_instance.web": {
        "type": "aws_instance",
        "primary": {
          "id": "i-0ad17607e5ee026d0",
     ...
}
```

Aquí puedes ver que el recurso `aws_instance.web` de nuestro código de Terraform está mapeado al ID de instancia `i-0ad17607e5ee026d0`.

Entonces, si el state de Terraform es solo un archivo, ¿qué es el state remoto?

## State remoto

Por defecto, Terraform escribe su archivo de state en tu sistema de archivos local. Esto está bien para proyectos personales, pero una vez que comienzas a trabajar con un equipo, las cosas se complican. En un equipo, necesitas asegurarte de que todos tengan una versión actualizada del archivo de state **y** asegurar que dos personas no estén haciendo cambios concurrentes.

¡Aquí entra el state remoto! El state remoto es simplemente almacenar el archivo de state de forma remota, en lugar de en tu sistema de archivos. Con state remoto, solo hay una copia, por lo que Terraform puede asegurar que siempre estés actualizado. Para evitar que los miembros del equipo modifiquen el state al mismo tiempo, Terraform puede bloquear el state remoto.

> El state remoto es simplemente almacenar el archivo de state de forma remota, en lugar de en tu sistema de archivos.

Bien, entonces el state remoto es excelente, pero desafortunadamente configurarlo puede ser un poco complicado. En AWS, puedes almacenarlo en un bucket de S3, pero necesitas crear el bucket, configurarlo correctamente, establecer sus permisos correctamente, crear una tabla de DynamoDB para el locking y luego asegurar que todos tengan credenciales adecuadas para escribir en él. Es prácticamente la misma historia en las otras nubes.

Como resultado, configurar state remoto puede ser un obstáculo molesto mientras los equipos adoptan Terraform.

Esto nos lleva a la primera razón para probar el almacenamiento remoto de state gratuito de HashiCorp...

## Razón n.º 1 — Fácil de configurar

A diferencia de otras soluciones de state remoto que requieren una configuración complicada para hacerlo bien, configurar el almacenamiento remoto de state gratuito es fácil.

> Configurar el almacenamiento remoto de state gratuito de HashiCorp es fácil

Paso 1 — Regístrate para obtener tu cuenta gratuita de [Terraform Cloud](https://app.terraform.io/signup)

Paso 2 — Cuando inicies sesión, llegarás a esta página donde crearás tu organización:

![](../../../blog/2019/4-reasons-to-try-hashicorps-new-free-terraform-remote-state-storage/pic1.webp)

Paso 3 — A continuación, ve a User Settings y genera un token:

![](../../../blog/2019/4-reasons-to-try-hashicorps-new-free-terraform-remote-state-storage/pic2.webp)

Paso 4 — Toma este token y crea un archivo local ~/.terraformrc:

```tf
credentials "app.terraform.io" {
  token = "mhVn15hHLylFvQ.atlasv1.jAH..."
}
```

Paso 5 — ¡Eso es todo! Ahora estás listo para almacenar tu state.

En tu proyecto de Terraform, agrega un bloque `terraform`:

```tf{3,5}
terraform {
  backend "remote" {
    organization = "my-org" # org name from step 2.
    workspaces {
      name = "my-app" # name for your app's state.
    }
  }
}
```

Ejecuta `terraform init` y ¡tachán! Tu state ahora se está almacenando en Terraform Enterprise. Puedes ver el state en la UI:

![](../../../blog/2019/4-reasons-to-try-hashicorps-new-free-terraform-remote-state-storage/pic3.webp)

Hablando de ver el state en una UI...

## Razón n.º 2 — Visor de state con todas las funciones

La segunda razón para probar Terraform Cloud es su visor de state con todas las funciones:

![](../../../blog/2019/4-reasons-to-try-hashicorps-new-free-terraform-remote-state-storage/pic4.webp)

Si alguna vez arruinaste tu state de Terraform y necesitaste descargar una versión antigua o quisiste un registro de auditoría para saber quién cambió qué, entonces te encantará esta funcionalidad.

Puedes ver el archivo de state completo en cada momento del tiempo:

![](../../../blog/2019/4-reasons-to-try-hashicorps-new-free-terraform-remote-state-storage/pic5.webp)

También puedes ver el diff de lo que cambió:

![](../../../blog/2019/4-reasons-to-try-hashicorps-new-free-terraform-remote-state-storage/pic6.webp)

Por supuesto, puedes encontrar una manera de obtener esta información de algunos de los otros backends de state, pero es difícil. Con el almacenamiento remoto de state de HashiCorp, la obtienes gratis.

## Razón n.º 3 — Locking manual

La tercera razón para probar Terraform Cloud es la capacidad de bloquear manualmente tu state.

¿Alguna vez has estado trabajando en una pieza de infraestructura y has querido asegurar que nadie pudiera hacer cambios en ella al mismo tiempo?

Terraform Cloud viene con la capacidad de bloquear y desbloquear states desde la UI:

![](../../../blog/2019/4-reasons-to-try-hashicorps-new-free-terraform-remote-state-storage/pic7.webp)

Mientras el state esté bloqueado, las operaciones `terraform` recibirán un error:

![](../../../blog/2019/4-reasons-to-try-hashicorps-new-free-terraform-remote-state-storage/pic8.webp)

Esto te ahorra muchas de estas:

![](../../../blog/2019/4-reasons-to-try-hashicorps-new-free-terraform-remote-state-storage/pic9.webp)

## Razón n.º 4 — Funciona con Atlantis

La razón final para probar Terraform Cloud es que funciona perfectamente con [Atlantis](https://www.runatlantis.io/).

Establece una variable de entorno `ATLANTIS_TFE_TOKEN` con un token de TFE y estás listo para continuar. Ve a <https://www.runatlantis.io/docs/terraform-cloud.html> para aprender más.

Conclusión
Te animo mucho a probar el nuevo backend de almacenamiento remoto de state gratuito. Es una oferta convincente frente a otros backends de state gracias a su facilidad de configuración, su visor de state con todas las funciones y sus capacidades de locking.

Si no estás en la lista de espera, regístrate aquí: <https://app.terraform.io/signup>.
