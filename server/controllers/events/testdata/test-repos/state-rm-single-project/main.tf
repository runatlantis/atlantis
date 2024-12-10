resource "random_id" "simple" {
  byte_length = 1
}

resource "random_id" "for_each" {
  for_each    = toset([var.var])
  byte_length = 1
}

resource "random_id" "count" {
  count       = 1
  byte_length = 1
}

variable "var" {
  default = "default"
}
