# Git Host Access Credentials
This page describes how to create credentials for your Git host (GitHub, GitLab, Bitbucket, or Azure DevOps)

that Atlantis will use to make API calls.
[[toc]]

## Create an Atlantis user (optional)
We recommend creating a new user named **@atlantis** (or something close) or using a dedicated CI user.

This isn't required (you can use an existing user), however all the comments that Atlantis writes
will come from that user so it might be confusing if its coming from a personal account.

![Example Comment](./images/example-comment.png)
<p align="center"><i>An example comment coming from the @atlantisbot user</i></p>

## Generating an Access Token
Once you've created a new user (or decided to use an existing one), you need to
generate an access token. Read on for the instructions for your specific Git host:
* [GitHub](#github)
* [GitLab](#gitlab)
* [Bitbucket Cloud (bitbucket.org)](#bitbucket-cloud-bitbucket-org)
* [Bitbucket Server (aka Stash)](#bitbucket-server-aka-stash)
* [Azure DevOps](#azure-devops)

### GitHub
- Create a Personal Access Token by following: [https://help.github.com/articles/creating-a-personal-access-token-for-the-command-line/#creating-a-token](https://help.github.com/articles/creating-a-personal-access-token-for-the-command-line/#creating-a-token)
- Create the token with **repo** scope
- Record the access token
::: warning
Your Atlantis user must also have "Write permissions" (for repos in an organization) or be a "Collaborator" (for repos in a user account) to be able to set commit statuses:
![Atlantis status](./images/status.png)
:::

### GitLab
- Follow: [https://docs.gitlab.com/ce/user/profile/personal_access_tokens.html#creating-a-personal-access-token](https://docs.gitlab.com/ce/user/profile/personal_access_tokens.html#creating-a-personal-access-token)
- Create a token with **api** scope
- Record the access token

### Bitbucket Cloud (bitbucket.org)
- Create an App Password by following [https://confluence.atlassian.com/bitbucket/app-passwords-828781300.html#Apppasswords-Createanapppassword](https://confluence.atlassian.com/bitbucket/app-passwords-828781300.html#Apppasswords-Createanapppassword)
- Label the password "atlantis"
- Select **Pull requests**: **Read** and **Write** so that Atlantis can read your pull requests and write comments to them
- Record the access token

### Bitbucket Server (aka Stash)
- Click on your avatar in the top right and select **Manage account**
- Click **Personal access tokens** in the sidebar
- Click **Create a token**
- Name the token **atlantis**
- Give the token **Read** Project permissions and **Write** Pull request permissions
- Click **Create** and record the access token

### Azure DevOps
- Create a Personal access token by following [https://docs.microsoft.com/en-us/azure/devops/organizations/accounts/use-personal-access-tokens-to-authenticate?view=azure-devops](https://docs.microsoft.com/en-us/azure/devops/organizations/accounts/use-personal-access-tokens-to-authenticate?view=azure-devops)
- Label the password "atlantis"
- The minimum scopes required for this token are:
  - Code (Read & Write)
  - Code (Status)
- Record the access token

## Next Steps
Once you've got your user and access token, you're ready to create a webhook secret. See [Creating a Webhook Secret](webhook-secrets.html).
