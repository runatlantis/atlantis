# HTTPS, SSL, TLS

When using a self-signed certificate for Atlantis (with flags `--ssl-cert-file` and `--ssl-key-file`),
there are a few considerations.

Atlantis uses the web server from the standard Go library, 
the method name is [ListenAndServeTLS](https://pkg.go.dev/net/http#ListenAndServeTLS).

`ListenAndServeTLS` acts identically to [ListenAndServe](https://pkg.go.dev/net/http#ListenAndServe),
except that it expects HTTPS connections. 
Additionally, files containing a certificate and matching private key for the server must be provided. 
If the certificate is signed by a certificate authority, 
the file passed to `--ssl-cert-file` should be the concatenation of the server's certificate, any intermediates, and the CA's certificate. 

If you have this error when specifying a TLS cert with a key: 
```
[ERROR] server.go:413 server: Tls: private key does not match public key
```

Check that the locally signed certificate authority is prepended to the self signed certificate.
A good example is shown at [Seth Vargo terraform implementation of atlantis-on-gke](https://github.com/sethvargo/atlantis-on-gke/blob/master/terraform/tls.tf#L64)

For Go specific TLS resources have a look at the repository by [denji called golang-tls](https://github.com/denji/golang-tls).

For a complete explanation on PKI, read this [article](https://smallstep.com/blog/everything-pki.html).


