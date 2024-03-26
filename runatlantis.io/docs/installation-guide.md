# Installation Guide
This guide is for installing a **production-ready** instance of Atlantis onto your
infrastructure:
1. First, ensure your Terraform setup meets the Atlantis **requirements**
    * See [Requirements](requirements.md)
1. Create **access credentials** for your Git host (GitHub, GitLab, Bitbucket, Azure DevOps)
    * See [Generating Git Host Access Credentials](access-credentials.md)
1. Create a **webhook secret** so Atlantis can validate webhooks
    * See [Creating a Webhook Secret](webhook-secrets.md)
1. **Deploy** Atlantis into your infrastructure
    * See [Deployment](deployment.md)
1. Configure **Webhooks** on your Git host so Atlantis can respond to your pull requests
    * See [Configuring Webhooks](configuring-webhooks.md)
1. Configure **provider credentials** so Atlantis can actually run Terraform commands
    * See [Provider Credentials](provider-credentials.md)

:::tip
If you want to test out Atlantis first, check out [Test Drive](../guide/test-drive.md)
and [Testing Locally](../guide/testing-locally.md).
:::
