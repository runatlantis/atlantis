variable "domain_name" {
  description = "The website endpoint of your S3 bucket."
}

variable "origin_id" {
  description = "Any string to name this origin."
}

variable "cnames" {
  description = "CNAME's for this distribution."
  type        = "list"
}

variable "acm_certificate_arn" {
  description = "ARN of ACM certificate used to provide SSL for this distribution's domain name."
}
