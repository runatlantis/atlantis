# VCS Features

Below are the available features and which VCS providers support them.

Some are configurable by the user with flags, others are handled internally. Some are unimplemented because of inherent deficiencies with the VCS, whereas most are just due to lack of developer support.

### CommentEmojiReaction
[`--emoji-reaction`](/docs/server-configuration.html#emoji-reaction)

Adds an emoji onto a comment when Atlantis is processing it

| *VCS* | *Supported* |
|---|---------|
| Github | ✔ |
| Gitlab | ✔ |
| BitbucketCloud | ✘ |
| BitbucketServer | ✘ |
| AzureDevops | ✔ |
| Gitea | ✘ |

### DiscardApprovalOnPlan
[`--discard-approval-on-plan`](/docs/server-configuration.html#discard-approval-on-plan)

Discard approval if a new plan has been executed

| *VCS* | *Supported* |
|---|---------|
| Github | ✔ |
| Gitlab | ✔ |
| BitbucketCloud | ✘ |
| BitbucketServer | ✘ |
| AzureDevops | ✘ |
| Gitea | ✘ |

### SingleFileDownload
Whether we can download a single file from the VCS

| *VCS* | *Supported* |
|---|---------|
| Github | ✔ |
| Gitlab | ✔ |
| BitbucketCloud | ✘ |
| BitbucketServer | ✘ |
| AzureDevops | ✘ |
| Gitea | ✔ |

### DetailedPullIsMergeable
Whether PullIsMergeable returns a detailed reason as to why it's unmergeable

| *VCS* | *Supported* |
|---|---------|
| Github | ✔ |
| Gitlab | ✔ |
| BitbucketCloud | ✘ |
| BitbucketServer | ✘ |
| AzureDevops | ✘ |
| Gitea | ✘ |

