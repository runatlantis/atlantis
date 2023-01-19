resource "random_id" "simple" {
  keepers     = {}
  byte_length = 1
}

resource "random_id" "for_each" {
  for_each    = toset([var.var])
  keepers     = {}
  byte_length = 1
}

resource "random_id" "count" {
  count       = 1
  keepers     = {}
  byte_length = 1
}

variable "var" {
  default = "default"
}
