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

## Configuring Atlantis

After following the above steps it is time to configure Atlantis. Assuming you have already provided the `slack-token` (via parameter or environment variable) you can now instruct Atlantis to send `apply` events to Slack.

In your Atlantis configuration you can now add the following:

```yaml
webhooks:
- event: apply
  workspace-regex: .*
  branch-regex: .*
  kind: slack
  channel: my-channel
```

If you are deploying Atlantis as a Helm chart, this can be implemented via the `config` parameter available for [chart customizations](https://github.com/runatlantis/helm-charts#customization):

```

## Use Server Side Config,
## ref: https://www.runatlantis.io/docs/server-configuration.html
config: |
   ---
   webhooks:
     - event: apply
       workspace-regex: .*
       branch-regex: .*
       kind: slack
       channel: my-channel
```



The `apply` event information will be sent to the `my-channel` Slack channel.
