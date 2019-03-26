# Configuring Webhooks
Atlantis needs to receive Webhooks from your Git host so that it can respond to pull request events.

:::tip Prerequisites
* You have created an [access credential](access-credentials.html)
* You have created a [webhook secret](webhook-secrets.html)
* You have [deployed](deployment.html) Atlantis and have a url for it
:::

See the instructions for your specific provider below.
[[toc]]

## GitHub/GitHub Enterprise
You can install your webhook at the [organization](https://help.github.com/articles/differences-between-user-and-organization-accounts/) level, or for each individual repository.

::: tip NOTE
If only some of the repos in your organization are to be managed by Atlantis, then you
may want to only install on specific repos for now.
:::

If you're installing on the organization, navigate to your organization's page and click **Settings**.
If installing on a single repository, navigate to the repository home page and click **Settings**.
- Select **Webhooks** or **Hooks** in the sidebar
- Click **Add webhook**
- set **Payload URL** to `http://$URL/events` (or `https://$URL/events` if you're using SSL) where `$URL` is where Atlantis is hosted. **Be sure to add `/events`**
- double-check you added `/events` to the end of your URL.
- set **Content type** to `application/json`
- set **Secret** to the Webhook Secret you generated previously
  - **NOTE** If you're adding a webhook to multiple repositories, each repository will need to use the **same** secret.
- select **Let me select individual events**
- check the boxes
	- **Pull request reviews**
	- **Pushes**
	- **Issue comments**
	- **Pull requests**
- leave **Active** checked
- click **Add webhook**
- See [Next Steps](#next-steps)

## GitLab
If you're using GitLab, navigate to your project's home page in GitLab
- Click **Settings > Integrations** in the sidebar
- set **URL** to `http://$URL/events` (or `https://$URL/events` if you're using SSL) where `$URL` is where Atlantis is hosted. **Be sure to add `/events`**
- double-check you added `/events` to the end of your URL.
- set **Secret Token** to the Webhook Secret you generated previously
  - **NOTE** If you're adding a webhook to multiple repositories, each repository will need to use the **same** secret.
- check the boxes
    - **Push events**
    - **Comments**
    - **Merge Request events**
- leave **Enable SSL verification** checked
- click **Add webhook**
- See [Next Steps](#next-steps)

## Bitbucket Cloud (bitbucket.org)
- Go to your repo's home page
- Click **Settings** in the sidebar
- Click **Webhooks** under the **WORKFLOW** section
- Click **Add webhook**
- Enter "Atlantis" for **Title**
- set **URL** to `http://$URL/events` (or `https://$URL/events` if you're using SSL) where `$URL` is where Atlantis is hosted. **Be sure to add `/events`**
- double-check you added `/events` to the end of your URL.
- Keep **Status** as Active
- Don't check **Skip certificate validation** because NGROK has a valid cert.
- Select **Choose from a full list of triggers**
- Under **Repository** **un**check everything
- Under **Issues** leave everything **un**checked
- Under **Pull Request**, select: Created, Updated, Merged, Declined and Comment created
- Click **Save**
<img src="../guide/images/bitbucket-webhook.png" alt="Bitbucket Webhook" style="max-height: 500px">
- See [Next Steps](#next-steps)

## Bitbucket Server (aka Stash)
- Go to your repo's home page
- Click **Settings** in the sidebar
- Click **Webhooks** under the **WORKFLOW** section
- Click **Create webhook**
- Enter "Atlantis" for **Name**
- set **URL** to `http://$URL/events` (or `https://$URL/events` if you're using SSL) where `$URL` is where Atlantis is hosted. **Be sure to add `/events`**
- Double-check you added `/events` to the end of your URL.
- Set **Secret** to the Webhook Secret you generated previously
  - **NOTE** If you're adding a webhook to multiple repositories, each repository will need to use the **same** secret.
- Under **Repository** select **Push**
- Under **Pull Request**, select: Opened, Modified, Merged, Declined, Deleted and Comment added
- Click **Save**<img src="../guide/images/bitbucket-server-webhook.png" alt="Bitbucket Webhook" style="max-height: 500px;">
- See [Next Steps](#next-steps)

## Next Steps
* To verify that Atlantis is receiving your webhooks, create a test pull request
  to your repo. You should see the request show up in the Atlantis logs at an `INFO` level.
* You'll now need to configure Atlantis to add your [Provider Credentials](provider-credentials.html)
