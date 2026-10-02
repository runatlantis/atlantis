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

Si la fusión automática está habilitada, el método de fusión se puede establecer en tres
lugares. De menor a mayor prioridad:

1. La bandera del servidor `--automerge-method` (o la variable de entorno
   `ATLANTIS_AUTOMERGE_METHOD`), que establece el valor predeterminado para todos los repos:

    ```shell
    atlantis server --automerge-method <method>
    ```

2. La key `automerge_method` en el `atlantis.yaml` de un repo, que sobrescribe el
   valor predeterminado del servidor para ese repo:

    ```yaml
    version: 3
    automerge: true
    automerge_method: <method>
    projects:
    - dir: .
    ```

3. La bandera `--auto-merge-method` en un solo comando `atlantis apply`, que
   sobrescribe las dos anteriores para ese comando:

    ```shell
    atlantis apply --auto-merge-method <method>
    ```

El valor `method` es uno de un conjunto normalizado de valores que Atlantis traduce a la
estrategia de fusión propia de cada proveedor:

- `merge` — fusionar con un commit de fusión
- `rebase` — hacer rebase sobre la rama base
- `squash` — combinar los commits en uno solo
- `fast-forward` — avanzar la rama base con fast-forward, es decir, fusionar sin commit de fusión

No todos los proveedores pueden realizar todos los métodos. Si un método no está soportado por
el proveedor del pull request, el comando se rechaza con la lista de métodos que
ese proveedor sí soporta:

| Método         | GitHub | Gitea / Forgejo | GitLab | Bitbucket Cloud | Bitbucket Server | Azure DevOps |
| -------------- | :----: | :-------------: | :----: | :-------------: | :--------------: | :----------: |
| `merge`        | ✓      | ✓               | ✓      | ✓               | ✓                | ✓            |
| `rebase`       | ✓      | ✓               |        |                 | ✓                | ✓            |
| `squash`       | ✓      | ✓               | ✓      | ✓               | ✓                | ✓            |
| `fast-forward` |        | ✓               |        | ✓               | ✓                |              |

Un método también debe estar permitido por la configuración propia del repositorio. Si el
método solicitado está deshabilitado para el repositorio (por ejemplo, las fusiones squash
están desactivadas), la API de fusión lo rechaza y Atlantis informa del fallo.

:::tip NOTA
Cada proveedor asigna el método normalizado a su estrategia nativa:
GitLab solo permite activar o desactivar el squash al aceptar un merge request (la
elección entre merge/fast-forward/rebase está fijada por la configuración del método de
fusión del proyecto), por lo que solo soporta `merge` y `squash`. En Bitbucket Server los
métodos corresponden a las estrategias `no-ff`, `rebase-ff`, `squash` y `ff-only`
respectivamente, y en Azure DevOps a las estrategias `noFastForward`, `rebase` y `squash`.
`fast-forward` requiere que la rama base se pueda avanzar con fast-forward.
:::

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
