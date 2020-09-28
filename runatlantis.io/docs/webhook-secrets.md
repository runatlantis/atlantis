# Webhook Secrets

Atlantis uses Webhook secrets to validate that the webhooks it receives from your
Git host are legitimate.

One way to confirm this would be to allowlist requests
to only come from the IPs of your Git host but an easier way is to use a Webhook
Secret.

::: tip NOTE
Webhook secrets are actually optional. However they're highly recommended for
security.
:::

::: tip NOTE
Azure DevOps uses Basic authentication for webhooks rather than webhook secrets.
:::

::: tip NOTE
An app-wide token is generated during [Github App setup](access-credentials.html#github-app). You can recover it by navigating to the [Github app settings page](https://github.com/settings/apps) and selecting "Edit" next to your Atlantis app's name. Token appears after clicking "Edit" under the Webhook header.
:::

::: warning
Bitbucket.org **does not** support webhook secrets.
To mitigate, use repo allowlists and IP allowlists. See [Security](security.html#bitbucket-cloud-bitbucket-org) for more information.
:::

## Generating A Webhook Secret
You can use any random string generator to create your Webhook secret. It should be > 24 characters.

For example:
* Generate via Ruby with `ruby -rsecurerandom -e 'puts SecureRandom.hex(32)'`
* Generate online with [https://www.browserling.com/tools/random-string](https://www.browserling.com/tools/random-string)

::: tip NOTE
You must use **the same** webhook secret for each repo.
:::

## Next Steps
* Record your secret
* You'll be using it later to [configure your webhooks](configuring-webhooks.html), however if you're
following the [Installation Guide](installation-guide.html) then your next step is to
[Deploy Atlantis](deployment.html)
