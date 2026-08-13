# Configuración de Atlantis

Hay tres métodos para configurar Atlantis:

1. Pasar flags al comando `atlantis server`
1. Crear un archivo de configuración de repositorio del lado del servidor y usar el flag `--repo-config`
1. Colocar un archivo `atlantis.yaml` en la raíz de sus repositorios de Terraform

## Flags

Los flags para `atlantis server` se usan para configurar la operación global de
Atlantis, por ejemplo establecer credenciales para su Git Host
o configurar certificados SSL.

Vea [Server Configuration](server-configuration.md) para más detalles.

## Configuración de repositorio del lado del servidor

Un archivo de configuración de repositorio del lado del servidor se usa para controlar el comportamiento por repositorio
y lo que los usuarios pueden hacer en archivos `atlantis.yaml` a nivel de repositorio.

Vea [Server-Side Repo Config](server-side-repo-config.md) para más detalles.

## Archivos `atlantis.yaml` a nivel de repositorio

Los archivos `atlantis.yaml` colocados en la raíz de sus repositorios de Terraform se pueden usar para
cambiar el comportamiento predeterminado de Atlantis para cada repositorio.

Vea [Repo-Level atlantis.yaml Files](repo-level-atlantis-yaml.md) para más detalles.
