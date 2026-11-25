---
aside: false
---
# Documentación de contribución de Atlantis

Esta documentación es para usuarios que quieren contribuir al proyecto Atlantis. Esto
puede variar desde escribir documentación, ayudar a la comunidad en Slack, discutir
issues, o escribir código.

:::tip ¿Buscas comenzar o usar Atlantis?
Si eres nuevo, revisa la [Guide](./guide.md) o la
[Documentation](./docs.md).
:::

## Siguientes pasos

### Previsualizar la documentación de configuración por versión

Instala las dependencias del sitio con `npm install` y compila las etiquetas de las versiones seleccionadas:

```sh
npm run website:build:versioned -- --versions v0.28.0,v0.40.0,v0.48.0 --outDir /tmp/atlantis-versioned-docs
python3 -m http.server 8081 --bind 127.0.0.1 --directory /tmp/atlantis-versioned-docs
```

Abre `http://127.0.0.1:8081/docs/server-configuration.html`. La documentación raíz se compila desde la versión más reciente incluida, indicada con `(latest)` en el selector del encabezado. `Main (next)` abre la documentación de desarrollo. Sin `--versions`, se selecciona el parche más reciente de cada una de las tres versiones menores más recientes disponibles en las etiquetas locales de Git. Obtén las etiquetas que falten antes de compilar.

La documentación por versión comienza en v0.28.0, la primera versión con VitePress. La documentación anterior sigue disponible en las [etiquetas de GitHub](https://github.com/runatlantis/atlantis/tags). Usa `--versions all` para compilar todas las etiquetas estables admitidas, o `--list-versions` para enumerarlas. Se rechazan las versiones explícitas anteriores a v0.28.0. El selector lee el manifiesto generado, por lo que las nuevas versiones no requieren cambios en el código fuente.

Para obtener fechas correctas, ejecuta `node scripts/website/fetch-history.mjs` antes de compilar desde una copia con historial limitado.

Configura `ATLANTIS_DOCS_BUILD_CONCURRENCY=4` para compilar cuatro versiones a la vez. El valor predeterminado es uno; se admiten valores de 1 a 8.

La compilación usa el tema actual para las versiones publicadas y Main. La raíz combina la documentación de la última versión con las páginas actuales de Contributing y Blog. La documentación en español sigue la misma selección de versión cuando existe una traducción. El Markdown histórico, los recursos y los espacios de compilación se leen de Git en directorios temporales; solo el sitio final se escribe en `--outDir`. Las páginas y los recursos generados no se guardan en el repositorio. La navegación y la búsqueda se mantienen dentro de su versión; Contributing y Blog enlazan al sitio actual. Cambiar a una página o sección no disponible muestra un aviso.

La última versión se compila tanto en la raíz como en su URL permanente porque sus rutas base son diferentes. Los enlaces al código fuente abren el archivo original de la etiqueta; las páginas sin traducción al español enlazan a su fuente en inglés.

Ejecuta las pruebas de enlaces y del navegador contra esa previsualización:

```sh
npm run website:test:versioned
ATLANTIS_VERSIONED_DOCS_URL=http://127.0.0.1:8081 npx playwright test versioned-docs.spec.js
```

Las pruebas del navegador usan las tres etiquetas de ejemplo anteriores. Configura `ATLANTIS_DOCS_BROWSER=chrome` para usar Chrome instalado en lugar de Chromium de Playwright.

Netlify obtiene las etiquetas y el historial completo, y publica la compilación con `--versions all`. En las previsualizaciones de despliegue (`CONTEXT=deploy-preview`), la documentación raíz muestra los cambios de la PR con Main (next) seleccionado; producción usa la última versión publicada. Las versiones históricas siguen disponibles en ambos casos. Configura `CONTEXT=deploy-preview` al compilar localmente para comprobar este comportamiento.

Las nuevas etiquetas aparecen en el siguiente despliegue del sitio. CI compila versiones representativas y ejecuta las pruebas de versiones y del navegador. `website:build` sigue disponible para previsualizar el código actual sin selección de versión.

### Contribuir a Atlantis

- [Decisiones de arquitectura](./contributing/adr/index.md)&nbsp;&nbsp;–&nbsp;&nbsp;Propuestas y decisiones actuales, con su estado.
- [Events Controller](../contributing/events-controller.md)&nbsp;&nbsp;–&nbsp;&nbsp;¿Cómo funcionan los eventos?
