provider "random" {}
terraform {
  required_providers {
    random = {
      source  = "hashicorp/random"
      version = "3.8.0"
    }
  }
}
