# Sending notifications via webhooks

It is possible to send notifications to external systems whenever a plan or apply is being done.

You can make requests to any HTTP endpoint or send messages directly to your Slack channel.

::: tip NOTE
Both `plan` and `apply` events are supported.
:::

## Configuration

Webhooks are configured in Atlantis [server-side configuration](server-configuration.md).
There can be many webhooks: sending notifications to different destinations or for different
workspaces/branches. Here is example configuration to send Slack messages for every plan and apply:

```yaml
webhooks:
- event: apply
  kind: slack
  channel: my-channel-id
- event: plan
  kind: slack
  channel: my-channel-id
```

If you are deploying Atlantis as a Helm chart, this can be implemented via the `config` parameter available for [chart customizations](https://github.com/runatlantis/helm-charts#customization):

```yaml
## Use Server Side Config,
## ref: https://www.runatlantis.io/docs/server-configuration.html
config: |
   ---
   webhooks:
     - event: apply
       kind: slack
       channel: my-channel-id
```

### Filter on workspace/branch

To limit notifications to particular workspaces or branches, use `workspace-regex` or `branch-regex` parameters.
If the workspace **and** branch matches respective regex, an event will be sent. Note that empty regular expression
(a result of unset parameter) matches every string.

## Using HTTP webhooks

You can send POST requests with JSON payload to any HTTP/HTTPS server.

### Configuring Atlantis

In your Atlantis [server-side configuration](server-configuration.md) you can add the following:

```yaml
webhooks:
- event: apply
  kind: http
  url: https://example.com/hooks
- event: plan
  kind: http
  url: https://example.com/hooks
```

The `apply` and `plan` event information will be POSTed to `https://example.com/hooks`.

You can supply any additional headers with `--webhook-http-headers` parameter (or environment variable),
for example for authentication purposes. See [webhook-http-headers](server-configuration.md#webhook-http-headers) for details.

### JSON payload

For apply events, the payload is a JSON-marshalled [ApplyResult](https://pkg.go.dev/github.com/runatlantis/atlantis/server/events/webhooks#ApplyResult) struct.

For plan events, the payload is a JSON-marshalled [PlanResult](https://pkg.go.dev/github.com/runatlantis/atlantis/server/events/webhooks#PlanResult) struct, which has the same structure as ApplyResult.

Example payload for an apply event:

```json
{
  "Workspace": "default",
  "Repo": {
    "FullName": "octocat/Hello-World",
    "Owner": "octocat",
    "Name": "Hello-World",
    "CloneURL": "https://:@github.com/octocat/Hello-World.git",
    "SanitizedCloneURL": "https://:<redacted>@github.com/octocat/Hello-World.git",
    "VCSHost": {
      "Hostname": "github.com",
      "Type": 0
    }
  },
  "Pull": {
    "Num": 2137,
    "HeadCommit": "7fd1a60b01f91b314f59955a4e4d4e80d8edf11d",
    "URL": "https://github.com/octocat/Hello-World/pull/2137",
    "HeadBranch": "feature/some-branch",
    "BaseBranch": "main",
    "Author": "octocat",
    "State": 0,
    "BaseRepo": {
      "FullName": "octocat/Hello-World",
      "Owner": "octocat",
      "Name": "Hello-World",
      "CloneURL": "https://:@github.com/octocat/Hello-World.git",
      "SanitizedCloneURL": "https://:<redacted>@github.com/octocat/Hello-World.git",
      "VCSHost": {
        "Hostname": "github.com",
        "Type": 0
      }
    }
  },
  "User": {
    "Username": "octocat",
    "Teams": null
  },
  "Success": true,
  "Directory": "terraform/example",
  "ProjectName": "example-project"
}
```

## Using Slack hooks

For this you'll need to:

* Create a Bot user in Slack
* Configure Atlantis to send notifications to Slack.

### Configuring Slack for Atlantis

* Go to [Slack: Apps](https://api.slack.com/apps)
* Click the `Create New App` button
* Select `From scratch` in the dialog that opens
* Give it a name, e.g. `atlantis-bot`.
* Select your Slack workspace
* Click `Create App`
* On the left go to `oAuth & Permissions`
* Scroll down to Scopes | Bot Token Scopes and add the following OAuth scopes:
  * `channels:read`
  * `chat:write`
  * `groups:read`
  * `incoming-webhook`
  * `mpim:read`
* Install the app onto your Slack workspace
* Copy the `Bot User OAuth Token` and provide it to Atlantis by using `--slack-token=xoxb-xxxxxxxxxxx` or via the environment `ATLANTIS_SLACK_TOKEN=xoxb-xxxxxxxxxxx`.
* Create a channel in your Slack workspace (e.g. `my-channel`) or use existing
* Add the app to Created channel or existing channel ( click channel name then tab integrations, there Click "Add apps"

### Configuring Atlantis

After following the above steps it is time to configure Atlantis. Assuming you have already provided the `slack-token` (via parameter or environment variable) you can now instruct Atlantis to send `apply` events to Slack.

In your Atlantis [server-side configuration](server-configuration.md) you can now add the following:

```yaml
webhooks:
- event: apply
  kind: slack
  channel: my-channel-id
- event: plan
  kind: slack
  channel: my-channel-id
```
