# Verificaciones de políticas personalizadas

Si desea ejecutar herramientas o scripts de políticas personalizados en lugar de la integración incorporada de Conftest, puede hacerlo configurando la opción `custom_policy_check` y ejecutándola en un workflow personalizado. Nota: la salida de la herramienta de políticas personalizada simplemente se analiza en busca de subcadenas "fail" para determinar si el conjunto de políticas pasó.

Esta opción se puede configurar ya sea a nivel del servidor en un [archivo de configuración repos.yaml](server-configuration.md) o a nivel del repositorio en un [archivo atlantis.yaml](repo-level-atlantis-yaml.md).

::: tip
Las verificaciones de políticas personalizadas admiten [aprobaciones de políticas sticky](policy-checking.md#sticky-policy-approvals). Use `policy_item_regex` para apuntar a las líneas de fallo significativas en la salida de su herramienta.
:::

## Ejemplo de configuración del lado del servidor

Configure las opciones `policy_check` e `custom_policy_check` en true, y ejecute la herramienta personalizada en los pasos de verificación de políticas como se ve a continuación.

```yaml
repos:
  - id: /.*/
    branch: /^main$/
    apply_requirements: [mergeable, undiverged, approved]
    policy_check: true
    custom_policy_check: true
    workflow: custom
workflows:
  custom:
    policy_check:
      steps:
        - show
        - run: cnspec scan terraform plan $SHOWFILE --policy-bundle example-cnspec-policies.mql.yaml 
policies:
  owners:
    users:
      - example_ghuser
  policy_sets:
    - name: example-set
      path: example-cnspec-policies.mql.yaml 
      source: local
```

## Ejemplo de atlantis.yaml a nivel de repositorio

Primero, deberá asegurarse de que `custom_policy_check` esté dentro del campo `allowed_overrides` de la configuración del lado del servidor. Luego, simplemente configure la opción personalizada en true en el proyecto específico que desea, como se muestra en el ejemplo `atlantis.yaml` a continuación:

```yaml
version: 3
projects:
  - name: example
    dir: ./example
    custom_policy_check: true
    autoplan:
      when_modified: ["*.tf"]
```
