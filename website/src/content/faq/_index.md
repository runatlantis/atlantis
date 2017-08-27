+++
draft= false
title = "FAQ"
description = "All your Atlantis questions answered!"
+++

## Does Atlantis affect Terraform [remote state](https://www.terraform.io/docs/state/remote.html)?

No. Atlantis does not interfere with Terraform remote state in anyway. Under the hood, Atlantis is simply executing `terraform plan` and `terraform apply`.

## How does Atlantis locking interact with Terraform [locking](https://www.terraform.io/docs/state/locking.html)?

Atlantis provides locking of pull requests that prevents concurrent modification of the same infrastructure (Terraform project) whereas Terraform locking only prevents two concurrent `terraform apply`'s from happening. 

Terraform locking can be used alongside Atlantis locking since Atlantis is simply executing terraform commands.

## How to run Atlantis in high availability mode? Does it need to be?

Atlantis server can easily be run under the supervision of a init system like `upstart` or `systemd` to make sure `atlantis server` is always running. 

Atlantis currently stores all locking and Terraform plans locally on disk under the `--data-dir` directory (defaults to `~/.atlantis`). Because of this there is currently no way to run two or more Atlantis instances concurrently.

However, if you were to lose the data, all you would need to do is run `atlantis plan` again on the pull requests that are open. If someone tries to run `atlantis apply` after the data has been lost then they will get an error back, so they will have to re-plan anyway.

## How to add SSL to Atlantis server?

Atlantis currently only supports HTTP. In order to add SSL you will need to front Atlantis server with NGINX or HAProxy. Follow the document [here](https://github.com/hootsuite/atlantis/blob/master/docs/nginx-ssl-proxy.md) to use configure NGINX with SSL as a reverse proxy.