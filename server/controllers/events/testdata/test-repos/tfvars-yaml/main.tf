terraform {
  backend "local" {
  }
}

resource "null_resource" "simple" {
  count = 1
}

variable "var" {
}

output "var" {
  value = var.var
}

output "workspace" {
  value = terraform.workspace
}
