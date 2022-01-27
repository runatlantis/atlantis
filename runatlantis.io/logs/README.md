# Real-time logs

## Getting Started
Atlantis supports streaming terraform logs in real time by default. Currently, only two commands are supported

* terraform init
* terraform plan
* terraform apply

::: warning as of now, custom workflow outputs and other terraform commands are not supported

In order to view real-time terraform logs, a user can navigate through the *details* section of a given project's plan or apply status check.

![Plan Command](./images/plan.png)

This will link to the atlantis UI which provides real-time logging in addition to native terraform syntax highlighting.

![Plan Output](./images/plan_output.png)

::: warning As of now the logs are currently stored in memory and cleared when a given pull request is closed, so this link shouldn't be persisted anywhere.

## Future Work 
In the future, users will also be able to persist the logs into a storage backend of their choice. 

