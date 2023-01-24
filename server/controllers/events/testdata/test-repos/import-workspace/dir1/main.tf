resource "random_id" "dummy1" {
  count = terraform.workspace == "ops" ? 1 : 0

  keepers     = {}
  byte_length = 1
}

resource "random_id" "dummy2" {
  count = terraform.workspace == "ops" ? 1 : 0

  keepers     = {}
  byte_length = 1
}

output "workspace" {
  value = terraform.workspace
}
