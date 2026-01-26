# VCS Features

Below are the available features and which VCS providers support them.

Some are configurable by the user with flags, others are handled internally. Some are unimplemented because of inherent deficiencies with the VCS, whereas most are just due to lack of developer support.

### CommentEmojiReaction
[`--emoji-reaction`](/docs/server-configuration.html#emoji-reaction)

Adds an emoji onto a comment when Atlantis is processing it

| *VCS* | *Supported* | *Notes* |
|---|---------|-------|
| Github | ✔ | [Supported Emojis](https://docs.github.com/en/rest/reactions/reactions?apiVersion=2022-11-28#about-reactions) |
| Gitlab | ✔ | [Supported Emojis](https://gitlab.com/gitlab-org/gitlab/-/blob/master/fixtures/emojis/digests.json) |
| BitbucketCloud | ✘ |  |
| BitbucketServer | ✘ |  |
| AzureDevops | ✔ | [Supported Emojis](https://learn.microsoft.com/en-us/azure/devops/project/wiki/markdown-guidance?view=azure-devops#emoji) |
| Gitea | ✘ |  |

### DiscardApprovalOnPlan
[`--discard-approval-on-plan`](/docs/server-configuration.html#discard-approval-on-plan)

Discard approval if a new plan has been executed

| *VCS* | *Supported* | *Notes* |
|---|---------|-------|
| Github | ✔ |  |
| Gitlab | ✔ | A group or Project token is required for this feature, see [reset-approvals-of-a-merge-request](https://docs.gitlab.com/api/merge_request_approvals/#reset-approvals-of-a-merge-request) |
| BitbucketCloud | ✘ |  |
| BitbucketServer | ✘ |  |
| AzureDevops | ✘ |  |
| Gitea | ✘ |  |

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

### HidePreviousPlanComments
[`--hide-prev-plan-comments`](/docs/server-configuration.html#hide-prev-plan-comments)

Hide previous plan comments to declutter PRs

| *VCS* | *Supported* | *Notes* |
|---|---------|-------|
| Github | ✔ | Ensure the `--gh-user` is set appropriately or comments will not be hidden. When using the GitHub App, you need to set `--gh-app-slug` to enable this feature. |
| Gitlab | ✔ |  |
| BitbucketCloud | ✔ | comments are deleted rather than hidden as Bitbucket does not support hiding comments. |
| BitbucketServer | ✘ |  |
| AzureDevops | ✘ |  |
| Gitea | ✔ |  |

