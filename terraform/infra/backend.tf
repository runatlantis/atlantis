terraform {
  required_version = ">= 1.0"
  backend "s3" {
    bucket         = "plentiau-terraform-state"
    key            = "alantis/terraform/infra/terraform.tfstate"
    region         = "ap-southeast-2"
    encrypt        = true
    dynamodb_table = "terraform-state-locking"
    role_arn       = "arn:aws:iam::028287609508:role/cross-account/state-storage-plentiau-terraform-state"
  }
}
