# Automerging
Atlantis can be configured to automatically merge a PR after all plans have
been successfully applied. Automerging can be enabled either by passing the
`--automerge` flag to the `atlantis server` command, or it can be specified
using `atlantis.yaml` at the top level:

```yaml
version: 2
automerge: true
projects:
- dir: project1
  autoplan:
    when_modified: ["../modules/**/*.tf", "*.tf*"]
```

The automerge setting is global, and if specified on the command line it will
override any `atlantis.yaml` settings. You may need to adjust the permissions
for your git provider to enable merging via the API.

When automerge is enabled, the changes will only be merged if all plan and
apply stages have succeeded.
