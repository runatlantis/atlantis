variable "var" {}
resource "null_resource" "this" {
}
output "var" {
  value = var.var
}

output "workspace" {
  value = terraform.workspace
}
