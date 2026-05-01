# FAQ

**Q: Does Atlantis affect Terraform [remote state](https://developer.hashicorp.com/terraform/language/state/remote)?**

A: No. Atlantis does not interfere with Terraform remote state in any way. Under the hood, Atlantis is simply executing `terraform plan` and `terraform apply`.

**Q: How does Atlantis locking interact with Terraform [locking](https://developer.hashicorp.com/terraform/language/state/locking)?**

A: Atlantis provides locking of pull requests that prevents concurrent modification of the same infrastructure (Terraform project) whereas Terraform locking only prevents two concurrent `terraform apply`'s from happening.

Terraform locking can be used alongside Atlantis locking since Atlantis is simply executing terraform commands.

**Q: How to run Atlantis in high availability mode? Does it need to be?**

A: Atlantis server can easily be run under the supervision of an init system like `upstart` or `systemd` to make sure `atlantis server` is always running.

Atlantis, by default, stores all locking and Terraform plans locally on disk under the `--data-dir` directory (defaults to `~/.atlantis`). If multiple Atlantis hosts are run by utilizing a shared redis backend, then it's important that the `data-dir` is using a shared filesystem between hosts.

However, if you were to lose the data, all you would need to do is run `atlantis plan` again on the pull requests that are open. If someone tries to run `atlantis apply` after the data has been lost then they will get an error back, so they will have to re-plan anyway.

**Q: How to add SSL to Atlantis server?**

A: First, you'll need to get a public/private key pair to serve over SSL.
These need to be in a directory accessible by Atlantis. Then start `atlantis server` with the `--ssl-cert-file` and `--ssl-key-file` flags.
See `atlantis server --help` for more information.

**Q: Can Atlantis detect infrastructure drift?**

A: Yes. When the `--enable-drift-detection` flag is set, Atlantis exposes API endpoints for drift detection, status, and remediation. Drift detection works by running `terraform plan` against the specified branch/ref (outside of a pull request context) and analyzing the plan output for changes. You can trigger drift detection via `POST /api/drift/detect` and retrieve cached results via `GET /api/drift/status`. See [API Endpoints](api-endpoints.md) for details.

**Q: How do I set up scheduled drift detection?**

A: Atlantis provides the drift detection API but does not include a built-in scheduler. You can use an external scheduler (e.g., cron, CI/CD pipelines, or a monitoring tool) to periodically call `POST /api/drift/detect`. Configure [drift webhooks](sending-notifications-via-webhooks.md) to receive Slack or HTTP notifications when drift is detected.

**Q: How can I get Atlantis up and running on AWS?**

A: There is [terraform-aws-atlantis](https://github.com/terraform-aws-modules/terraform-aws-atlantis) project where complete Terraform configurations for running Atlantis on AWS Fargate are hosted. Tested and maintained.
