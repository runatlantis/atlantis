terraform {
  required_version = ">=1.2"
  cloud {
    organization = "atlantis-test"
    workspaces {
      # Atlantis uses the workspace name as a path component, so a name that
      # escapes the directory must be rejected rather than used.
      name = "../../evil"
    }
  }
}
