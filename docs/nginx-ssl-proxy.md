# NGINX SSL Proxy
This document shows how to configure Nginx with SSL as a reverse proxy for Atlantis server.

* Install NGINX

```bash
sudo apt-get update
sudo apt-get install nginx
```

* Install a SSL Certificate
This certificate can be purchased or generated. Here is a example of generating a self signed SSL certificate.

```bash
cd /etc/nginx
sudo openssl req -x509 -nodes -days 365 -newkey rsa:2048 -keyout /etc/nginx/cert.key -out /etc/nginx/cert.crt
```
You will be prompted to enter some information about the certificate. Fill those as you like.

* Edit NGINX Config

```bash
sudo vim /etc/nginx/sites-enabled/default
server {
    listen 80;
    return 301 https://$host$request_uri;
}

server {

    listen 443;
    server_name atlantis.domain.com;

    ssl_certificate           /etc/nginx/cert.crt;
    ssl_certificate_key       /etc/nginx/cert.key;

    ssl on;
    ssl_session_cache  builtin:1000  shared:SSL:10m;
    ssl_protocols  TLSv1 TLSv1.1 TLSv1.2;
    ssl_ciphers HIGH:!aNULL:!eNULL:!EXPORT:!CAMELLIA:!DES:!MD5:!PSK:!RC4;
    ssl_prefer_server_ciphers on;

    access_log            /var/log/nginx/atlantis.access.log;

    location / {

      proxy_set_header        Host $host;
      proxy_set_header        X-Real-IP $remote_addr;
      proxy_set_header        X-Forwarded-For $proxy_add_x_forwarded_for;
      proxy_set_header        X-Forwarded-Proto $scheme;

      # Fixes the “It appears that your reverse proxy set up is broken" error.
      proxy_pass          http://localhost:4141;
      proxy_read_timeout  90;

      proxy_redirect      http://localhost:4141 https://atlantis.domain.com;
    }
  }
  ```

  * Restart NGINX

  ```bash
  sudo service nginx restart
  
  ```