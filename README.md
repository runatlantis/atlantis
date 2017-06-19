# atlantis
[![CircleCI](https://circleci.com/gh/hootsuite/atlantis.svg?style=svg&circle-token=6a0c78c9b1fd77486c72a5e22512c7c9175f2aaf)](https://circleci.com/gh/hootsuite/atlantis)

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
