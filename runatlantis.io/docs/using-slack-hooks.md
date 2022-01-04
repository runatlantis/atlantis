# Using Slack hooks

It is possible to use Slack to send notifications to your Slack channel whenever an apply is being done.

::: tip NOTE
Currently only `apply` events are supported.
:::

For this you'll need to:

* Create a Bot user in Slack
* Configure Atlantis to send notifications to Slack.

## Configuring Slack for Atlantis

* Go to [https://api.slack.com/apps](https://api.slack.com/apps)
* Click the `Create New App` button
* Select `From scratch` in the dialog that opens
* Give it a name, e.g. `atlantis-bot`.
* Select your Slack workspace
* Click `Create App`
* On the left go to `oAuth & Permissions`
* Copy the `Bot User OAuth Token` and provide it to Atlantis by using `--slack-token=xoxb-xxxxxxxxxxx` or via the environment `ATLANTIS_SLACK_TOKEN=xoxb-xxxxxxxxxxx`.
* Scroll down to scopes and add the following:
  * `channels:read`
  * `chat:write`
  * `groups:read`
  * `incoming-webhook`
  * `mpim:read`
* Install the app onto your Slack workspace

## Configuring Atlantis

After following the above steps it is time to configure Atlantis. Assuming you have already provided the `slack-token` (via parameter or environment variable) you can now instruct Atlantis to send `apply` events to Slack.

In your Atlantis configuration you can now add the following:

```yaml
webhooks:
- event: apply
  workspace-regex: .*
  kind: slack
  channel: my-channel
```

The `apply` event information will be sent to the `my-channel` Slack channel.
