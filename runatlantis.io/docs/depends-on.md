# Depends_on Argument
[[toc]]

## Description
The depends_on argument allow you to enforce dependencies between projects. Use the depends_on argument to handle cases
where require one project to be applied prior to the other.

## What Happens if one or more project's dependencies are not applied?
If there's one or more projects in the dependency list is not in an applied status, users will see an error if they try 
to run `atlantis apply`.

### Usage
1. In `atlantis.yaml` file specify the `depends_on` key under the project config:
   #### atlantis.yaml
   ```yaml
    version: 3
    projects:
    - dir: .
      name: project-2
      depends_on: [project-1]
    ```
