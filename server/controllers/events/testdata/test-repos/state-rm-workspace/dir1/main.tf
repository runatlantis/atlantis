resource "random_id" "dummy1" {
  count = terraform.workspace == "ops" ? 1 : 0

  byte_length = 1
}

output "workspace" {
  value = terraform.workspace
}
