# Configuring Webhooks
Atlantis needs to receive Webhooks from your Git host so that it can respond to pull request events.

:::tip Prerequisites
* You have created an [access credential](access-credentials.md)
* You have created a [webhook secret](webhook-secrets.md)
* You have [deployed](deployment.md) Atlantis and have a url for it
:::

See the instructions for your specific provider below.
[[toc]]

## GitHub/GitHub Enterprise
You can install your webhook at the [organization](https://docs.github.com/en/get-started/learning-about-github/types-of-github-accounts) level, or for each individual repository.

::: tip NOTE
If only some of the repos in your organization are to be managed by Atlantis, then you
may want to only install on specific repos for now.
:::

When authenticating as a Github App, Webhooks are automatically created and need no additional setup, beyond being installed to your organization/user account after creation. Refer to the [Github App setup](access-credentials.md#github-app) section for instructions on how to do so.

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
- Click **Settings > Webooks** in the sidebar
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
- Under **Pull Request**, select: Opened, Source branch updated, Merged, Declined, Deleted and Comment added
- Click **Save**<img src="../guide/images/bitbucket-server-webhook.png" alt="Bitbucket Webhook" style="max-height: 600px;">
- See [Next Steps](#next-steps)

## Azure DevOps
Webhooks are installed at the [team project](https://docs.microsoft.com/en-us/azure/devops/organizations/projects/about-projects?view=azure-devops) level, but may be restricted to only fire based on events pertaining to [specific repos](https://docs.microsoft.com/en-us/azure/devops/service-hooks/services/webhooks?view=azure-devops) within the team project.

- Navigate anywhere within a team project, ie: `https://dev.azure.com/orgName/projectName/_git/repoName`
- Select **Project settings** in the lower-left corner
- Select **Service hooks**
  - If you see the message "You do not have sufficient permissions to view or configure subscriptions." you need to ensure your user is a member of either the organization's "Project Collection Administrators" group or the project's "Project Administrators" group.
  - To add your user to the Project Collection Build Administrators group, navigate to the organization level, click **Organization Settings** and then click **Permissions**. You should be at `https://dev.azure.com/<organization>/_settings/groups`. Now click on the **<organization>/Project Collection Administrators** group and add your user as a member.
  - To add your user to the Project Administrators group, navigate to the project level, click **Project Settings** and then click **Permissions**. You should be at `https://dev.azure.com/<organization>/<project>/_settings/permissions`. Now click on the **[<project>]/Project Administrators** group and add your user as a member.
- Click **Create subscription** or the green plus icon to add a new webhook
- Scroll to the bottom of the list and select **Web Hooks**
- Click **Next**
- Under "Trigger on this type of event", select **Pull request created**
  - Optionally, select a repository under **Filters** to restrict the scope of this webhook subscription to a specific repository
- Click **Next**
- Set **URL** to `http://$URL/events` where `$URL` is where Atlantis is hosted. Note that SSL, or `https://$URL/events`, is required if you set a Basic username and password for the webhook). **Be sure to add `/events`**
- It is strongly recommended to set a Basic Username and Password for all webhooks
- Leave all three drop-down menus for `...to send` set to **All**
- Resource version should be set to **1.0** for `Pull request created` and `Pull request updated` event types and **2.0** for `Pull request commented on`
- **NOTE** If you're adding a webhook to multiple team projects or repositories (using filters), each repository will need to use the **same** basic username and password.
- Click **Finish**

Repeat the process above until you have webhook subscriptions for the following event types that will trigger on all repositories Atlantis will manage:

- Pull request created (you just added this one)
- Pull request updated
- Pull request commented on

- See [Next Steps](#next-steps)

## Next Steps
* To verify that Atlantis is receiving your webhooks, create a test pull request
  to your repo.
* You should see the request show up in the Atlantis logs at an `INFO` level.
* You'll now need to configure Atlantis to add your [Provider Credentials](provider-credentials.md)
