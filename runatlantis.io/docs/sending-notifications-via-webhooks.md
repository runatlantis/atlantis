# Sending notifications via webhooks

It is possible to send notifications to external systems whenever an apply is being done or drift is detected.

You can make requests to any HTTP endpoint or send messages directly to your Slack channel.

::: tip NOTE
The `apply` and `drift` events are supported.
:::

## Configuration

Webhooks are configured in Atlantis [server-side configuration](server-configuration.md).
There can be many webhooks: sending notifications to different destinations or for different
workspaces/branches. Here is example configuration to send Slack messages for every apply:

```yaml
webhooks:
- event: apply
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
```

The `apply` event information will be POSTed to `https://example.com/hooks`.

You can supply any additional headers with `--webhook-http-headers` parameter (or environment variable),
for example for authentication purposes. See [webhook-http-headers](server-configuration.md#webhook-http-headers) for details.

### JSON payload

The payload is a JSON-marshalled [ApplyResult](https://pkg.go.dev/github.com/runatlantis/atlantis/server/events/webhooks#ApplyResult) struct.

Example payload:

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
```

## Drift detection webhooks

When [drift detection](api-endpoints.md#post-apidriftdetect) is enabled (`--enable-drift-detection`), you can configure webhooks to be notified whenever infrastructure drift is detected. Drift webhooks are sent automatically after a `POST /api/drift/detect` request finds projects with drift.

::: tip NOTE
Drift webhooks use `event: drift` in the webhook configuration. They are independent from `event: apply` webhooks.
:::

### Configuring drift webhooks

Drift webhooks are configured alongside apply webhooks in the same `webhooks` configuration block. You can send drift notifications to Slack, HTTP endpoints, or both:

```yaml
webhooks:
# Apply webhook (existing)
- event: apply
  kind: slack
  channel: apply-notifications

# Drift webhooks
- event: drift
  kind: slack
  channel: drift-alerts
- event: drift
  kind: http
  url: https://example.com/drift-webhook
```

::: tip NOTE
Unlike apply webhooks, drift webhooks do not support `workspace-regex` or `branch-regex` filtering because drift detection operates at the repository level, not in the context of a pull request.
:::

### Slack drift message format

When drift is detected, the Slack message includes:

* **Color**: Red if drift was found, green if no drift
* **Text**: "Drift detected in owner/repo" or "No drift in owner/repo"
* **Fields**: Repository, Ref, Projects with drift (count), Detection ID

### HTTP drift webhook payload

The HTTP webhook sends a POST request with the following JSON payload:

```json
{
  "repository": "octocat/Hello-World",
  "ref": "main",
  "detection_id": "550e8400-e29b-41d4-a716-446655440000",
  "projects_with_drift": 1,
  "total_projects": 2,
  "projects": [
    {
      "project_name": "vpc",
      "path": "modules/vpc",
      "workspace": "production",
      "has_drift": true,
      "to_add": 1,
      "to_change": 2,
      "to_destroy": 0,
      "summary": "Plan: 1 to add, 2 to change, 0 to destroy."
    },
    {
      "project_name": "ec2",
      "path": "modules/ec2",
      "workspace": "production",
      "has_drift": false,
      "to_add": 0,
      "to_change": 0,
      "to_destroy": 0,
      "summary": ""
    }
  ]
}
```

The same `--webhook-http-headers` headers configured for apply webhooks are also sent with drift webhook requests.
