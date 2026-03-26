resource "null_resource" "simple" {
  count = 1
}

output "workspace" {
  value = terraform.workspace
}
