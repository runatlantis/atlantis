terraform {
  cloud {
    organization = "atlantis-test"
    workspaces {
      name = "tofu-workspace"
    }
  }
}
