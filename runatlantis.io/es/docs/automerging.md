# Fusión automática

Atlantis se puede configurar para fusionar automáticamente un pull request después de que todos los plans se hayan aplicado correctamente.

![Automerge](../../docs/images/automerge.png)

## Cómo habilitar

La fusión automática se puede habilitar de cualquiera de estas formas:

1. Pasando la bandera `--automerge` a `atlantis server`. Esto establece el parámetro globalmente; sin embargo, la declaración explícita en la configuración del repo será respetada y tendrá prioridad.
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

Si la fusión automática está habilitada, puedes deshabilitarla para un solo comando `atlantis apply`
con la opción `--auto-merge-disabled`.

## Cómo establecer el método de fusión para la fusión automática

Si la fusión automática está habilitada, puedes establecer un método de fusión predeterminado con la
bandera del servidor `--automerge-method` o la variable de entorno `ATLANTIS_AUTOMERGE_METHOD`.

```shell
atlantis server --automerge-method <method>
```

Puedes sobrescribir el valor predeterminado del servidor para un solo comando `atlantis apply` con
la opción `--auto-merge-method`.

```shell
atlantis apply --auto-merge-method <method>
```

El valor `method` debe ser uno de:

- merge
- rebase
- squash

Esto actualmente solo está implementado para el VCS de GitHub.

## Requisitos

### Todos los plans deben tener éxito

Cuando la fusión automática está habilitada, **todos los plans** en un pull request **deben tener éxito** antes de que
**cualquier** plan pueda ser aplicado.

Por ejemplo, imagina este escenario:

1. Abro un pull request que hace cambios en dos proyectos de Terraform, en `dir1/`
   e `dir2/`.
1. El plan para `dir2/` falla porque mi sintaxis de Terraform es incorrecta.

En este escenario, no puedo ejecutar

```shell
atlantis apply -d dir1
```

Aunque ese plan tuvo éxito, porque **todos** los plans deben tener éxito para que **cualquier** plan
pueda ser guardado.

Una vez que corrija el problema en `dir2`, puedo hacer push de un nuevo commit que activará un
autoplan. Entonces podré aplicar ambos plans.

### Todos los plans deben ser aplicados

Si múltiples proyectos/dirs/workspaces están configurados para ser planeados automáticamente,
entonces todos deben ser aplicados antes de que Atlantis fusione automáticamente el PR.

## Permisos

El usuario de VCS de Atlantis debe tener la capacidad de fusionar pull requests.
