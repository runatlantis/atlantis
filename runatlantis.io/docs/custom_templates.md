# Custom Comment Templates

Custom Comment Templates can be defined to override the default VCS comment templates used by Atlantis. 

[[toc]]

##Usage
Paths to custom templates for the supported steps can be specified in the Server Side Repo Config. The `template_overrides` key for a repo maps the template to override with the path the the template on the local filesystem. 
```
repos:
  - id: /github.com/*/
    apply_requirements: [mergeable, undiverged]
    template_overrides: 
      project_plan_success: server/events/templates/planSuccessWrappedCustomTmpl.tmpl
```
##Supported Overrides
The following Templates can be overriden: 
1. `project_err`: Error template for a single project within a repo. 
2. `project_failure`: Failure template for a single project within a repo.
3. `project_plan_success`: Plan success template for a single project within a repo. 
4. `project_policy_check_success`: Policy Check success template for a single project within a repo. 
5. `project_apply_success`: Apply success template for a single project within a repo. 
6. `project_version_success`: Version success template for a single project within a repo.
7. `plan`: Overall `plan` template for single and multi project plans. This is also used for the policy check. 
8. `apply`: Overall `apply` template for single and multi project applyies.
9. `version`: Overall `version` template for single and multi project version commands.

::: The `plan`, `apply`, and `version` templates support the extended template function library described [here] ( https://pkg.go.dev/github.com/masterminds/sprig)