# Requirements

[[toc]]

## Git Host
* GitHub (public, private or enterprise)
* GitLab (public, private or enterprise)

If you would like support for BitBucket, please add a :+1: to [this ticket](https://github.com/runatlantis/atlantis/issues/30)
and click "Subscribe" to be notified when support is available.

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
you need to create an `atlantis.yaml` file. See [atlantis.yaml Reference](../docs/atlantis-yaml-reference.html) for more details.

###  Terraform Workspaces
This refers to [Terraform Workspaces](https://www.terraform.io/docs/state/workspaces.html). You need
to tell Atlantis the names of your workspaces with an `atlantis.yaml` file. See [atlantis.yaml Reference](../docs/atlantis-yaml-reference.html).

### .tfvars Files
```
.
├── production.tfvars
│── staging.tfvars
└── main.tf
```
With .tfvars files, for Atlantis to be able to plan automatically, you need to create
an `atlantis.yaml` file to tell it to use `-var-file`.
See [atlantis.yaml Reference](../docs/atlantis-yaml-reference.html) for more details.

## Terraform Workspaces
Terraform introduced [Workspaces](https://www.terraform.io/docs/state/workspaces.html) in Terraform v0.9. They allow for
> a single directory of Terraform configuration to be used to manage multiple distinct sets of infrastructure resources

If you're using a Terraform version >= 0.9.0, Atlantis supports workspaces through an
`atlantis.yaml` file that tells Atlantis the names of your workspaces or through the
the `-w` flag. For example:
```
atlantis plan -w staging
```

## Terraform Versions
By default, Atlantis will use the `terraform` executable that is in its path.
To use a specific version of Terraform:
1. Install the desired version of Terraform into the `$PATH` of where Atlantis is
 running and name it `terraform{version}`, ex. `terraform0.8.8`.
2. Create an `atlantis.yaml` file for your repo and set the `terraform_version` key.
See [atlantis.yaml Reference](../docs/atlantis-yaml-reference.html) for more details.
