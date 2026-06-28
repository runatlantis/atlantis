terraform {
  cloud {
    organization = "atlantis-test"
    workspaces {
      name = "tf-workspace"
    }
  }
}
