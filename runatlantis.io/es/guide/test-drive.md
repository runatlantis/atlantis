# Prueba Atlantis

Para probar Atlantis en un repositorio de ejemplo, descarga la última release desde
[GitHub](https://github.com/runatlantis/atlantis/releases)

Una vez que hayas extraído el archivo, ejecuta:

```bash
./atlantis testdrive
```

Este modo configura Atlantis en un repositorio de prueba para que puedas probarlo. Hará lo siguiente:

- Hará un fork de un proyecto de ejemplo de Terraform en tu cuenta de GitHub
- Instalará Terraform (si aún no está en tu PATH)
- Instalará [ngrok](https://ngrok.com/) para que podamos exponer Atlantis a GitHub
- Iniciará Atlantis para que puedas ejecutar comandos en el pull request

## Siguientes pasos

- Si estás listo para probar la ejecución de Atlantis en **tus repositorios**, entonces lee [Testing Locally](testing-locally.md).
- Si estás listo para instalar Atlantis correctamente en infraestructura real, entonces ve a la [Installation Guide](../docs/installation-guide.md).
