resource "null_resource" "simple" {
  count = "1"
}

variable "var" {
  default = "default"
}

output "var" {
  value = var.var
}

output "workspace" {
  value = terraform.workspace
}
