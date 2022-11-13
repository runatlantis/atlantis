terraform {
  required_version = ">=1.2"
  cloud {
    organization = "atlantis-test"
    workspaces {
      name = "test-workspace"
    }
  }
}
