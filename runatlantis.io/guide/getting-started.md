# Getting Started
These instructions are for running Atlantis locally so you can test it out against
your own repositories before deciding whether to install it more permanently.

::: tip
If you want to set up a production-ready Atlantis installation, read [Deployment](../docs/deployment.html).
:::

Steps:

[[toc]]

## Install Terraform
`terraform` needs to be in the `$PATH` for Atlantis.
Download from [https://www.terraform.io/downloads.html](https://www.terraform.io/downloads.html)
```
unzip path/to/terraform_*.zip -d /usr/local/bin
```

## Download Atlantis
Get the latest release from [https://github.com/runatlantis/atlantis/releases](https://github.com/runatlantis/atlantis/releases)
and unpackage it.

## Download Ngrok
Atlantis needs to be accessible somewhere that github.com/gitlab.com or your GitHub/GitLab Enterprise installation can reach.
One way to accomplish this is with ngrok, a tool that forwards your local port to a random
public hostname.

Go to [https://ngrok.com/download](https://ngrok.com/download), download ngrok and `unzip` it.

Start `ngrok` on port `4141` and take note of the hostname it gives you:
```bash
./ngrok http 4141
```

In a new tab (where you'll soon start Atlantis) create an environment variable with
ngrok's hostname:
```bash
URL=https://{YOUR_HOSTNAME}.ngrok.io
```

## Create a Webhook Secret
GitHub and GitLab use webhook secrets so clients can verify that the webhooks came
from them. Create a random string of any length (you can use [https://www.random.org/strings/](https://www.random.org/strings/))
and set an environment variable:
```
SECRET={YOUR_RANDOM_STRING}
```

## Add Webhook
Take the URL that ngrok output and create a webhook in your GitHub or GitLab repo:

### GitHub
- Go to your repo's settings
- Select **Webhooks** or **Hooks** in the sidebar
- Click **Add webhook**
- set **Payload URL** to your ngrok url with `/events` at the end. Ex. `https://c5004d84.ngrok.io/events`
- double-check you added `/events` to the end of your URL.
- set **Content type** to `application/json`
- set **Secret** to your random string
- select **Let me select individual events**
- check the boxes
	- **Pull request reviews**
	- **Pushes**
	- **Issue comments**
	- **Pull requests**
- leave **Active** checked
- click **Add webhook**

### GitLab
- Go to your repo's home page
- Click **Settings > Integrations** in the sidebar
- set **URL** to your ngrok url with `/events` at the end. Ex. `https://c5004d84.ngrok.io/events`
- double-check you added `/events` to the end of your URL.
- set **Secret Token** to your random string
- check the boxes
    - **Push events**
    - **Comments**
    - **Merge Request events**
- leave **Enable SSL verification** checked
- click **Add webhook**

## Create an access token for Atlantis
We recommend using a dedicated CI user or creating a new user named **@atlantis** that performs all API actions, however for testing,
you can use your own user. Here we'll create the access token that Atlantis uses to comment on the pull request and
set commit statuses.

### GitHub
- follow [https://help.github.com/articles/creating-a-personal-access-token-for-the-command-line/#creating-a-token](https://help.github.com/articles/creating-a-personal-access-token-for-the-command-line/#creating-a-token)
- create a token with **repo** scope
- set the token as an environment variable
```
TOKEN={YOUR_TOKEN}
```

### GitLab
- follow [https://docs.gitlab.com/ce/user/profile/personal_access_tokens.html#creating-a-personal-access-token](https://docs.gitlab.com/ce/user/profile/personal_access_tokens.html#creating-a-personal-access-token)
- create a token with **api** scope
- set the token as an environment variable
```
TOKEN={YOUR_TOKEN}
```

## Start Atlantis
You're almost ready to start Atlantis, just set one more variable:

```
USERNAME={the username of your GitHub or GitLab user}
```
Now you can start Atlantis, the exact command differs depending on your Git Host:

### GitHub
```
atlantis server --atlantis-url $URL --gh-user $USERNAME --gh-token $TOKEN --gh-webhook-secret $SECRET
```

### GitHub Enterprise
```
HOSTNAME=YOUR_GITHUB_ENTERPRISE_HOSTNAME # ex. github.runatlantis.io, without the scheme
atlantis server --atlantis-url $URL --gh-user $USERNAME --gh-token $TOKEN --gh-webhook-secret $SECRET --gh-hostname $HOSTNAME
```

### GitLab
```
atlantis server --atlantis-url $URL --gitlab-user $USERNAME --gitlab-token $TOKEN --gitlab-webhook-secret $SECRET
```

### GitLab Enterprise
```
HOSTNAME=YOUR_GITLAB_ENTERPRISE_HOSTNAME # ex. gitlab.runatlantis.io, without the scheme
atlantis server --atlantis-url $URL --gitlab-user $USERNAME --gitlab-token $TOKEN --gitlab-webhook-secret $SECRET --gitlab-hostname $HOSTNAME
```

## Create a pull request
Create a pull request so you can test Atlantis.
::: tip
You could add a null resource as a test:
```hcl
resource "null_resource" "example" {}
```
Or just modify the whitespace in a file.
:::

### Autoplan
You should see Atlantis logging about receiving the webhook and you should see the output of `terraform plan` on your repo.

Atlantis tries to figure out the directory to plan in based on the files modified.
If you need to customize the directories that that Atlantis runs in or the commands it runs if you're using workspaces
or `.tfvars` files, see [atlantis.yaml Reference](../docs/atlantis-yaml-reference.html).

### Manual Plan
To manually `plan` in a specific directory or workspace, comment on the pull request using the `-d` or `-w` flags:
```
atlantis plan -d mydir
atlantis plan -w staging
```

To add additional arguments to the underlying `terraform plan` you can use:
```
atlantis plan -- -target=resource -var 'foo=bar'
```

### Apply
If you'd like to `apply`, type a comment: `atlantis apply`. You can use the `-d` or `-w` flags to point
Atlantis at a specific plan. Otherwise it tries to apply the plan for the root directory.

## Next Steps
* You're done! Hopefully Atlantis is working with your repo and you're ready to move on to a [production-ready deployment](../docs/deployment.html).
* If it's not working as expected, you may need to customize how Atlantis runs with an `atlantis.yaml` file.
See [atlantis.yaml Reference](../docs/atlantis-yaml-reference.html).
* Check out our full documentation for more details: [Documentation](../docs/).
