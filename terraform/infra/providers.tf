provider "aws" {
  region = "ap-southeast-2"

  default_tags {
    tags = {
      env        = "infra"
      managed_by = "terraform"
      repo       = "atlantis"
    }
  }

  # test account only
  allowed_account_ids = ["028287609508"]
}
