# Generating and Upload Certs

*This process needs to be automated!*

* Generate Let's Encrypt certs for [atlantis.run](https://atlantis.run)

```bash
sudo certbot certonly --manual --server https://acme-v01.api.letsencrypt.org/directory -d atlantis.run -d www.atlantis.run
```

Follow the instructions after running the command

* Upload the certs to AWS to be used by Cloudfront

```bash
sudo aws iam upload-server-certificate --server-certificate-name atlantis_run_lets_encrypt_cert --certificate-body file:///etc/letsencrypt/live/atlantis.run/cert.pem --private-key file:///etc/letsencrypt/live/atlantis.run/privkey.pem --certificate-chain file:///etc/letsencrypt/live/atlantis.run/chain.pem --path /cloudfront/certs/
```

## Automated Process using certbot-s3front

* First get your certbot installation path

```bash
cat $(which certbot) | head -1
#!/usr/local/Cellar/certbot/0.17.0_1/libexec/bin/python2.7
```

* Use the `pip` found in the `bin` folder from the above path to install certbot-s3front

```bash
/usr/local/Cellar/certbot/0.17.0_1/libexec/bin/pip install certbot-s3front
```

* Run the following command with aws credentials from lastpass

```bash
AWS_ACCESS_KEY_ID="REPLACE_WITH_YOUR_KEY" \
AWS_SECRET_ACCESS_KEY="REPLACE_WITH_YOUR_SECRET" \
certbot --agree-tos -a certbot-s3front:auth \
--certbot-s3front:auth-s3-bucket atlantis.run \
-i certbot-s3front:installer \
--certbot-s3front:installer-cf-distribution-id DISTRIBUTION_ID \
-d atlantis.run
```