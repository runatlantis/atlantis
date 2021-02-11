resource "null_resource" "this" {
}
output "workspace" {
  value = terraform.workspace
}
