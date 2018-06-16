resource "null_resource" "simple" {
  count = 1
}

variable "var" {
  default = "default"
}

output "this" {
  value = "${var.var}"
}