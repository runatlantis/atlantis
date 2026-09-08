# HTTPS, SSL, TLS

Al usar un certificado autofirmado para Atlantis (con los flags `--ssl-cert-file` y `--ssl-key-file`),
hay algunas consideraciones.

Atlantis usa el servidor web de la biblioteca estándar de Go,
el nombre del método es [ListenAndServeTLS](https://pkg.go.dev/net/http#ListenAndServeTLS).

`ListenAndServeTLS` actúa de manera idéntica a [ListenAndServe](https://pkg.go.dev/net/http#ListenAndServe),
excepto que espera conexiones HTTPS.
Además, se deben proporcionar archivos que contengan un certificado y la clave privada correspondiente para el servidor.
Si el certificado está firmado por una autoridad certificadora,
el archivo pasado a `--ssl-cert-file` debe ser la concatenación del certificado del servidor, cualquier intermedio y el certificado de la CA.

Si tiene este error al especificar un certificado TLS con una clave:

```plain
[ERROR] server.go:413 server: Tls: private key does not match public key
```

Verifique que la autoridad certificadora firmada localmente esté antepuesta al certificado autofirmado.
Se muestra un buen ejemplo en [Seth Vargo terraform implementation of atlantis-on-gke](https://github.com/sethvargo/atlantis-on-gke/blob/master/terraform/tls.tf#L64-L84)

Para recursos TLS específicos de Go, eche un vistazo al repositorio de [denji llamado golang-tls](https://github.com/denji/golang-tls).

Para una explicación completa sobre PKI, lea este [artículo](https://smallstep.com/blog/everything-pki.html).
