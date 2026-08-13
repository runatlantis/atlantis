# Automerging

Atlantis puede configurarse para fusionar automáticamente un pull request después de que todos los plans se hayan
aplicado correctamente.

![Automerge](../../docs/images/automerge.png)

## Cómo habilitar

Automerging puede habilitarse de cualquiera de estas maneras:

1. Pasando la flag `--automerge` a `atlantis server`. Esto establece el parámetro globalmente; sin embargo, la declaración explícita en la configuración del repo será respetada y tendrá prioridad.
1. Estableciendo `automerge: true` en el archivo `atlantis.yaml` del repo:

    ```yaml
    version: 3
    automerge: true
    projects:
    - dir: .
    ```

    :::tip NOTA
    Si un repo tiene un archivo `atlantis.yaml`, entonces cada proyecto en el repo necesita
    ser configurado bajo la clave `projects`.
    :::

## Cómo deshabilitar

Si automerge está habilitado, puedes deshabilitarlo para un solo comando `atlantis apply`
con la opción `--auto-merge-disabled`.

## Cómo establecer el método de fusión para automerge

Si automerge está habilitado, puedes establecer un método de fusión predeterminado con la
flag del servidor `--automerge-method` o la variable de entorno `ATLANTIS_AUTOMERGE_METHOD`.

```shell
atlantis server --automerge-method <method>
```

Puedes sobrescribir el valor predeterminado del servidor para un solo comando `atlantis apply` con
la opción `--auto-merge-method`.

```shell
atlantis apply --auto-merge-method <method>
```

El `method` debe ser uno de:

- merge
- rebase
- squash

Esto actualmente solo está implementado para el VCS GitHub.

## Requisitos

### Todos los plans deben tener éxito

Cuando automerge está habilitado, **todos los plans** en un pull request **deben tener éxito** antes de que
**cualquier** plan pueda ser aplicado.

Por ejemplo, imagina este escenario:

1. Abro un pull request que realiza cambios en dos proyectos de Terraform, en `dir1/`
   e `dir2/`.
1. El plan para `dir2/` falla porque mi sintaxis de Terraform es incorrecta.

En este escenario, no puedo ejecutar

```shell
atlantis apply -d dir1
```

Aunque ese plan tuvo éxito, porque **todos** los plans deben tener éxito para que **cualquier** plan
pueda ser guardado.

Una vez que corrija el problema en `dir2`, puedo hacer push de un nuevo commit que activará un
autoplan. Luego podré aplicar ambos plans.

### Todos los plans deben ser aplicados

Si múltiples proyectos/dirs/workspaces están configurados para ser planificados automáticamente,
entonces todos ellos deben ser aplicados antes de que Atlantis fusione automáticamente el PR.

## Permisos

El usuario VCS de Atlantis debe tener la capacidad de fusionar pull requests.
