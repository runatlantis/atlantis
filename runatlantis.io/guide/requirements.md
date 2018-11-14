# Requirements

[[toc]]

## Git Host
* GitHub (public, private or enterprise)
* GitLab (public, private or enterprise)
* Bitbucket Cloud aka bitbucket.org (public or private)
* Bitbucket Server aka Stash

## Remote State
Atlantis supports all remote state backends. It **does not** support local state
because it does not commit the modified state files back to version control.

## Repository Structure
Atlantis supports any Terraform project structures, for example:

### Single Terraform project at repo root
```
.
├── main.tf
└── ...
```

### Multiple project folders
```
.
├── project1
│   ├── main.tf
|   └── ...
└── project2
    ├── main.tf
    └── ...
```

### Modules
```
.
├── project1
│   ├── main.tf
|   └── ...
└── modules
    └── module1
        ├── main.tf
        └── ...
```
With modules, if you want `project1` automatically planned when `module1` is modified
you need to create an `atlantis.yaml` file. See [atlantis.yaml Use Cases](atlantis-yaml-use-cases.html#configuring-autoplanning) for more details.

###  Terraform Workspaces
::: tip
See [Terraform's docs](https://www.terraform.io/docs/state/workspaces.html) if you are unfamiliar with workspaces.
:::
If you're using a Terraform version >= 0.9.0, Atlantis supports workspaces through an
`atlantis.yaml` file that tells Atlantis the names of your workspaces
(see [atlantis.yaml Use Cases](atlantis-yaml-use-cases.html#supporting-terraform-workspaces) for more details)
or through the `-w` flag. For example:
```
atlantis plan -w staging
atlantis apply -w staging
```


### .tfvars Files
```
.
├── production.tfvars
│── staging.tfvars
└── main.tf
```
For Atlantis to be able to plan automatically with `.tfvars files`, you need to create
an `atlantis.yaml` file to tell it to use `-var-file={YOUR_FILE}`.
See [atlantis.yaml Use Cases](atlantis-yaml-use-cases.html#using-tfvars-files) for more details.

## Terraform Versions
By default, Atlantis will use the `terraform` executable that is in its path.
To use a specific version of Terraform:
1. Install the desired version of Terraform into the `$PATH` of where Atlantis is
 running and name it `terraform{version}`, ex. `terraform0.8.8`.
2. Create an `atlantis.yaml` file for your repo and set the `terraform_version` key.
See [atlantis.yaml Use Cases](atlantis-yaml-use-cases.html#terraform-versions) for more details.

## Next Steps
Check out our [full documentation](../docs/).
