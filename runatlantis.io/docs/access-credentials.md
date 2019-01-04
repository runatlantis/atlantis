# Git Host Access Credentials
This doc describes how to create credentials on your Git host (GitHub, GitLab or Bitbucket)
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
generate an access token. Read on for the instructions for your Git host.

### Create a GitHub Token
**NOTE: The Atlantis user must have "Write permissions" (for repos in an organization) or be a "Collaborator" (for repos in a user account) to be able to set commit statuses:**
![Atlantis status](./images/status.png)
- create a Personal Access Token by following [https://help.github.com/articles/creating-a-personal-access-token-for-the-command-line/#creating-a-token](https://help.github.com/articles/creating-a-personal-access-token-for-the-command-line/#creating-a-token)
- create the token with **repo** scope
- record the access token

### Create a GitLab Token
- follow [https://docs.gitlab.com/ce/user/profile/personal_access_tokens.html#creating-a-personal-access-token](https://docs.gitlab.com/ce/user/profile/personal_access_tokens.html#creating-a-personal-access-token)
- create a token with **api** scope
- record the access token

### Create a Bitbucket Cloud (bitbucket.org) App Password
- create an App Password by following [https://confluence.atlassian.com/bitbucket/app-passwords-828781300.html#Apppasswords-Createanapppassword](https://confluence.atlassian.com/bitbucket/app-passwords-828781300.html#Apppasswords-Createanapppassword)
- Label the password "atlantis"
- Select **Pull requests**: **Read** and **Write** so that Atlantis can read your pull requests and write comments to them
- record the access token

### Create a Bitbucket Server (aka Stash) Personal Access Token
- Click on your avatar in the top right and select **Manage account**
- Click **Personal access tokens** in the sidebar
- Click **Create a token**
- Name the token **atlantis**
- Give the token **Read** Project permissions and **Write** Pull request permissions
- Click **Create** and record the access token

## Next Steps
Once you've got your user and access token, you're ready to create a webhook secret. See [Creating a Webhook Secret](webhook-secrets.html).
