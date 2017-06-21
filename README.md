# atlantis
[![CircleCI](https://circleci.com/gh/hootsuite/atlantis/tree/master.svg?style=svg&circle-token=08bf5b34233b0e168a9dd73e01cafdcf7dc4bf16)](https://circleci.com/gh/hootsuite/atlantis/tree/master)

A [terraform](https://www.terraform.io/) collaboration tool that enables you to collaborate on infrastructure safely and securely.

## Locking
When a **Run** is triggered, the set of infrastructure that is being modified is locked against any modification or planning until the **Run** has been
completed by an `apply` and the pull request has been merged

```
{
  "data_dir": "/var/lib/atlantis",
  "locking": {
    "backend": "file"
  }
}

{
  "locking": {
    "backend": "dynamodb"
  }
}
```
