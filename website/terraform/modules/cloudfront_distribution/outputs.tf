output "domain_name" {
  description = "The domain name for this distribution. Point your Route53 records here."
  value = "${aws_cloudfront_distribution.website_distribution.domain_name}"
}

output "hosted_zone_id" {
  description = "Zone ID of the CloudFront distribution. Use this for your Route53 record."
  value = "${aws_cloudfront_distribution.website_distribution.hosted_zone_id}"
}
