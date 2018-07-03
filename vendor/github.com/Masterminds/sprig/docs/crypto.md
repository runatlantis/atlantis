# Cryptographic and Security Functions

Sprig provides a couple of advanced cryptographic functions.

## sha1sum

The `sha1sum` function receives a string, and computes it's SHA1 digest.

```
sha1sum "Hello world!"
```

## sha256sum

The `sha256sum` function receives a string, and computes it's SHA256 digest.

```
sha256sum "Hello world!"
```

The above will compute the SHA 256 sum in an "ASCII armored" format that is
safe to print.

## derivePassword

The `derivePassword` function can be used to derive a specific password based on
some shared "master password" constraints. The algorithm for this is
[well specified](http://masterpasswordapp.com/algorithm.html).

```
derivePassword 1 "long" "password" "user" "example.com"
```

Note that it is considered insecure to store the parts directly in the template.

## genPrivateKey

The `genPrivateKey` function generates a new private key encoded into a PEM
block.

It takes one of the values for its first param:

- `ecdsa`: Generate an elyptical curve DSA key (P256)
- `dsa`: Generate a DSA key (L2048N256)
- `rsa`: Generate an RSA 4096 key

## buildCustomCert

The `buildCustomCert` function allows customizing the certificate.

It takes the following string parameters:

- A base64 encoded PEM format certificate
- A base64 encoded PEM format private key

It returns a certificate object with the following attributes:

- `Cert`: A PEM-encoded certificate
- `Key`: A PEM-encoded private key

Example:

```
$ca := buildCustomCert "base64-encoded-ca-key" "base64-encoded-ca-crt"
```

Note that the returned object can be passed to the `genSignedCert` function
to sign a certificate using this CA.

## genCA

The `genCA` function generates a new, self-signed x509 certificate authority.

It takes the following parameters:

- Subject's common name (cn)
- Cert validity duration in days

It returns an object with the following attributes:

- `Cert`: A PEM-encoded certificate
- `Key`: A PEM-encoded private key

Example:

```
$ca := genCA "foo-ca" 365
```

Note that the returned object can be passed to the `genSignedCert` function
to sign a certificate using this CA.

## genSelfSignedCert

The `genSelfSignedCert` function generates a new, self-signed x509 certificate.

It takes the following parameters:

- Subject's common name (cn)
- Optional list of IPs; may be nil
- Optional list of alternate DNS names; may be nil
- Cert validity duration in days

It returns an object with the following attributes:

- `Cert`: A PEM-encoded certificate
- `Key`: A PEM-encoded private key

Example:

```
$cert := genSelfSignedCert "foo.com" (list "10.0.0.1" "10.0.0.2") (list "bar.com" "bat.com") 365
```

## genSignedCert

The `genSignedCert` function generates a new, x509 certificate signed by the
specified CA.

It takes the following parameters:

- Subject's common name (cn)
- Optional list of IPs; may be nil
- Optional list of alternate DNS names; may be nil
- Cert validity duration in days
- CA (see `genCA`)

Example:

```
$ca := genCA "foo-ca" 365
$cert := genSignedCert "foo.com" (list "10.0.0.1" "10.0.0.2") (list "bar.com" "bat.com") 365 $ca
```
