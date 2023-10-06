# Custom Policy Checks
If you want to run custom policy tools or scripts instead of the built-in Conftest integration, you can do so by setting the `custom_policy_check` option and running it in a custom workflow.  Note: custom policy tool output is simply parsed for "fail" substrings to determine if the policy set passed. 

This option can be configured either at the server-level in a [repos.yaml config file](server-configuration.md) or at the repo-level in an [atlantis.yaml file.](repo-level-atlantis-yaml.md). 

## Server-side config example
Set the `policy_check` and `custom_policy_check` options to true, and run the custom tool in the policy check steps as seen below.  No

```yaml
repos:
  - id: /.*/
    branch: /^main$/
    apply_requirements: [mergeable, undiverged, approved]
    policy_check: true
    custom_policy_check: true
    workflow: custom
workflows:
  custom:
    policy_check:
      steps:
        - show
        - run: cnspec scan terraform plan $SHOWFILE --policy-bundle example-cnspec-policies.mql.yaml 
policies:
  owners:
    users:
      - example_ghuser
  policy_sets:
    - name: example-set
      path: example-cnspec-policies.mql.yaml 
      source: local
```


## Repo-level atlantis.yaml example
First, you will need to ensure `custom_policy_check` is within the `allowed_overrides` field of the server-side config.  Next, just set the custom option to true on the specific project you want as shown in the example `atlantis.yaml` below:

```yaml
version: 3
projects:
  - name: example
    dir: ./example
    custom_policy_check: true
    autoplan:
      when_modified: ["*.tf"]
```