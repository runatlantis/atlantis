# Webhook Secrets

Atlantis uses Webhook secrets to validate that the webhooks it receives from your
Git host are legitimate.

One way to confirm this would be to whitelist requests
to only come from the IPs of your Git host but an easier way is to use a Webhook
Secret.

::: tip NOTE
Webhook secrets are actually optional. However they're highly recommended for
security.
:::

::: tip NOTE
Azure DevOps uses Basic authentication for webhooks rather than webhook secrets.
:::

::: warning
Bitbucket.org **does not** support webhook secrets.
To mitigate, use repo whitelists and IP whitelists. See [Security](security.html#bitbucket-cloud-bitbucket-org) for more information.
:::

## Generating A Webhook Secret
You can use any random string generator to create your Webhook secret. It should be > 24 characters.

For example:
* Generate via Ruby with `ruby -rsecurerandom -e 'puts SecureRandom.hex(20)'`
* Generate online with [https://www.random.org/passwords/?num=2&len=20&format=html&rnd=new](https://www.random.org/passwords/?num=2&len=20&format=html&rnd=new)

::: tip NOTE
You must use **the same** webhook secret for each repo.
:::

## Next Steps
* Record your secret
* You'll be using it later to [configure your webhooks](configuring-webhooks.html), however if you're
following the [Installation Guide](installation-guide.html) then your next step is to
[Deploy Atlantis](deployment.html)
