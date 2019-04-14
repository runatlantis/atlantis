// This project sets up DNS entries for runatlantis.io. The site is hosted
// on Netlify.

provider "aws" {
  region = "us-east-1"
}

terraform {
  backend "s3" {
    bucket = "lkysow-terraform-states"
    key    = "runatlantis/atlantis/website"
    region = "us-east-1"
  }
}

variable "www_domain_name" {
  default = "www.runatlantis.io"
}

variable "root_domain_name" {
  default = "runatlantis.io"
}

resource "aws_route53_zone" "zone" {
  name = var.root_domain_name
}

resource "aws_route53_record" "www" {
  zone_id = aws_route53_zone.zone.zone_id
  name    = var.www_domain_name
  type    = "CNAME"
  ttl     = "300"
  records = ["runatlantis.netlify.com"]
}

resource "aws_route53_record" "root" {
  zone_id = aws_route53_zone.zone.zone_id

  // Note the name is blank here.
  name = ""
  type = "A"
  ttl  = "300"

  // This IP is for Netlify.
  records = ["104.198.14.52"]
}

// MailGun Records
resource "aws_route53_record" "mailgun_txt_0" {
  zone_id = aws_route53_zone.zone.zone_id
  name    = ""
  type    = "TXT"
  ttl     = "300"
  records = ["v=spf1 include:mailgun.org include:servers.mcsv.net ~all"]
}

resource "aws_route53_record" "mailgun_txt_1" {
  zone_id = aws_route53_zone.zone.zone_id
  name    = "krs._domainkey"
  type    = "TXT"
  ttl     = "300"
  records = ["k=rsa; p=MIGfMA0GCSqGSIb3DQEBAQUAA4GNADCBiQKBgQDW6rVlC11aSUQuUia02QRPkW2C1wLU/23Mx1PZHATpYSgLMo91MhVip1V1uVsC/rhqsvLiR6l0Cv/x7dG0lNQf3UPfn8Ld1qnjY66+HGt6crnuBJ6kpWYNRSVOlUU8tJrp6I0yNqvxDV689lI+HflyxCA1JP2SR5A9bL1oYJH64QIDAQAB"]
}

resource "aws_route53_record" "mailgun_mx" {
  zone_id = aws_route53_zone.zone.zone_id
  name    = ""
  type    = "MX"
  ttl     = "300"
  records = ["10 mxa.mailgun.org", "10 mxb.mailgun.org"]
}

resource "aws_route53_record" "mailgun_cname" {
  zone_id = aws_route53_zone.zone.zone_id
  name    = "email"
  type    = "CNAME"
  ttl     = "300"
  records = ["mailgun.org"]
}

resource "aws_route53_record" "mailchimp_cname" {
  zone_id = aws_route53_zone.zone.zone_id
  name    = "k1._domainkey"
  type    = "CNAME"
  ttl     = "300"
  records = ["dkim.mcsv.net"]
}

