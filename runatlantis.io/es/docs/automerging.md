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

### Los plans fallidos no descartan los plans exitosos

Cuando un plan en un pull request falla, Atlantis conserva los plans que tuvieron éxito, para que
todavía puedan aplicarse individualmente. Atlantis no fusionará automáticamente el pull request
hasta que todos los proyectos hayan sido aplicados (ver más abajo).

Por ejemplo, imagina este escenario:

1. Abro un pull request que hace cambios en dos proyectos de Terraform, en `dir1/`
   e `dir2/`.
1. El plan para `dir2/` falla porque mi sintaxis de Terraform es incorrecta.

En este escenario, todavía puedo ejecutar

```shell
atlantis apply -d dir1
```

porque el plan para `dir1/` tuvo éxito y fue guardado.

Una vez que corrija el problema en `dir2`, puedo hacer push de un nuevo commit que activará un
autoplan. Después de aplicar `dir2`, todos los proyectos han sido aplicados y Atlantis
fusiona el pull request.

### Todos los plans deben ser aplicados

Si múltiples proyectos/dirs/workspaces están configurados para ser planeados automáticamente,
entonces todos deben ser aplicados antes de que Atlantis fusione automáticamente el PR.

## Permisos

El usuario de VCS de Atlantis debe tener la capacidad de fusionar pull requests.
