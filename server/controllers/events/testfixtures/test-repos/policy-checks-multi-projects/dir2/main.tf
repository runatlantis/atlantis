resource "null_resource" "forbidden" {
  count = 1
}

output "workspace" {
  value = terraform.workspace
}
