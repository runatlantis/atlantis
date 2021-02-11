resource "null_resource" "simple" {
  count = 1
}

resource "null_resource" "simple2" {}
resource "null_resource" "simple3" {}

variable "var" {
  default = "default"
}

output "var" {
  value = var.var
}

output "workspace" {
  value = terraform.workspace
}
