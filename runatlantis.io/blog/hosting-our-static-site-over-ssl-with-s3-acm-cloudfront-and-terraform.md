---
title: Hosting Our Static Site over SSL with S3, ACM, CloudFront and Terraform
lang: en-US
---

# Hosting Our Static Site over SSL with S3, ACM, CloudFront and Terraform

::: info
This post was originally written on March 4, 2018

Original post: <https://medium.com/runatlantis/hosting-our-static-site-over-ssl-with-s3-acm-cloudfront-and-terraform-513b799aec0f>
:::

In this post I cover how I hosted <www.runatlantis.io> using

- S3 — for storing the static site
- CloudFront — for serving the static site over SSL
- AWS Certificate Manager — for generating the SSL certificates
- Route53 — for routing the domain name <www.runatlantis.io> to the correct location

I chose Terraform in this case because Atlantis is a tool for automating and collaborating on Terraform in a team (see github.com/runatlantis/atlantis)–and so obviously it made sense to host our homepage using Terraform–but also because it's now much easier to manage. I don't have to go into the AWS console and click around to find what settings I want to change. Instead I can just look at ~100 lines of code, make a change, and run `terraform apply`.

::: info
NOTE: 4 months after this writing, I moved the site to [Netlify](https://www.netlify.com/) because it automatically builds from my master branch on any change, updates faster since I don't need to wait for the Cloudfront cache to expire and gives me [deploy previews](https://www.netlify.com/blog/2016/07/20/introducing-deploy-previews-in-netlify/) of changes. The DNS records are still hosted on AWS.
:::

# Overview

There's a surprising number of components required to get all this working so I'm going to start with an overview of what they're all needed for. Here's what the final architecture looks like:

![](/blog/hosting-our-static-site/pic1.webp)

That's what the final product looks like, but lets start with the steps required to get there.

## Step 1 — Generate The Site

The first step is to have a site generated. Our site uses [Hugo](https://gohugo.io/), a Golang site generator. Once it's set up, you just need to run `hugo` and it will generate a directory with HTML and all your content ready to host.

## Step 2 — Host The Content

Once you've got a website, you need it to be accessible on the internet. I used S3 for this because it's dirt cheap and it integrates well with all the other necessary components. I simply upload my website folder to the S3 bucket.

## Step 3 — Generate an SSL Certificate

I needed to generate an SSL certificate for <https://www.runatlantis.io>. I used the AWS Certificate Manager for this because it's free and is easily integrated with the rest of the system.

## Step 4 — Set up DNS

Because I'm going to host the site on AWS services, I need requests to <www.runatlantis.io> to be routed to those services. Route53 is the obvious solution.

## Step 5 — Host with CloudFront

At this point, we've generated an SSL certificate for <www.runatlantis.io> and our website is available on the internet via its [S3 url](http://www.runatlantis.io.s3-website-us-east-1.amazonaws.com/) so can't we just CNAME to the S3 bucket and call it a day? Unfortunately not.

Since we generated our own certificate, we would need S3 to sign its responses using our certificiate. S3 doesn't support this and thus we need CloudFront. CloudFront supports using our own SSL cert and will just pull its data from the S3 bucket.

# Terraform Time

Now that we know what our architecture should look like, it's simply a matter of writing the Terraform.

## Initial Setup

Create a new file `main.tf`:

<<< @/public/blog/hosting-our-static-site/code/main.tf

## S3 Bucket

Assuming we've generated our site content already, we need to create an S3 bucket to host the content.

<<< @/public/blog/hosting-our-static-site/code/s3-bucket.tf

We should be able to run Terraform now to create the S3 bucket

```sh
terraform init
`terraform apply`
```

![](/blog/hosting-our-static-site/pic2.webp)

Now we want to upload our content to the S3 bucket:

```sh
$ cd dir/with/website
# generate the HTML
$ hugo -d generated
$ cd generated
# send it to our S3 bucket
$ aws s3 sync . s3://www.runatlantis.io/ # change this to your bucket
```

Now we need the S3 url to see our content:

```sh
$ terraform state show aws_s3_bucket.www | grep website_endpoint
website_endpoint                       = www.runatlantis.io.s3-website-us-east-1.amazonaws.com
```

You should see your site hosted at that url!

## SSL Certificate

Let's use the AWS Certificate Manager to create our SSL certificate.

<<< @/public/blog/hosting-our-static-site/code/ssl-cert.tf

Before you run `terraform apply`, ensure you're forwarding any of

- `administrator@your_domain_name`
- `hostmaster@your_domain_name`
- `postmaster@your_domain_name`
- `webmaster@your_domain_name`
- `admin@your_domain_name`

To an email address you can access. Then, run `terraform apply` and you should get an email from AWS to confirm you own this domain where you'll need to click on the link.

## CloudFront

Now we're ready for CloudFront to host our website using the S3 bucket for the content and using our SSL certificate. Warning! There's a lot of code ahead but most of it is just defaults.

<<< @/public/blog/hosting-our-static-site/code/cloudfront.tf

Apply the changes with `terraform apply` and then find the domain name that CloudFront gives us:

```sh
$ terraform state show aws_cloudfront_distribution.www_distribution | grep ^domain_name
domain_name                                                                                          = d1l8j8yicxhafq.cloudfront.net
```

You'll probably get an error if you go to that URL right away. You need to wait a couple minutes for CloudFront to set itself up. It took me 10 minutes. You can view its progress in the console: <https://console.aws.amazon.com/cloudfront/home>

## DNS

We're almost done! We've got CloudFront hosting our site, now we need to point our DNS at it.

<<< @/public/blog/hosting-our-static-site/code/dns.tf

If you bought your domain from somewhere else like Namecheap, you'll need to point your DNS at the nameservers listed in the state for the Route53 zone you created. First `terraform apply` (which may take a while), then find out your nameservers.

```sh
$ terraform state show aws_route53_zone.zone
id             = Z2FNAJGFW912JG
comment        = Managed by Terraform
force_destroy  = false
name           = runatlantis.io
name_servers.# = 4
name_servers.0 = ns-1349.awsdns-40.org
name_servers.1 = ns-1604.awsdns-08.co.uk
name_servers.2 = ns-412.awsdns-51.com
name_servers.3 = ns-938.awsdns-53.net
tags.%         = 0
zone_id        = Z2FNAJGFW912JG
```

Then look at your domain's docs for how to change your nameservers to all 4 listed.

## That's it...?

Once the DNS propagates you should see your site at `https://www.yourdomain`! But what about `https://yourdomain`? i.e. without the `www.`? Shouldn't this redirect to `https://www.yourdomain`?

## Root Domain

It turns out, we need to create a whole new S3 bucket, CloudFront distribution and Route53 record just to get this to happen. That's because although S3 can serve up a redirect to the www version of your site, it can't host SSL certs and so you need CloudFront. I've included all the terraform necessary for that below.

Congrats! You're done!

<iframe src="https://cdn.embedly.com/widgets/media.html?src=https%3A%2F%2Fgiphy.com%2Fembed%2Fl0MYt5jPR6QX5pnqM%2Ftwitter%2Fiframe&amp;display_name=Giphy&amp;url=https%3A%2F%2Fmedia.giphy.com%2Fmedia%2Fl0MYt5jPR6QX5pnqM%2Fgiphy.gif&amp;image=https%3A%2F%2Fi.giphy.com%2Fmedia%2Fl0MYt5jPR6QX5pnqM%2Fgiphy.gif&amp;key=d04bfffea46d4aeda930ec88cc64b87c&amp;type=text%2Fhtml&amp;schema=giphy" allowfullscreen="" frameborder="0" height="244" width="435" title="The Office Party Hard GIF - Find &amp; Share on GIPHY" class="fr n gh dv bg" scrolling="no"></iframe>

If you're using Terraform in a team, check out Atlantis: <https://github.com/runatlantis/atlantis> for automation and collaboration to make your team happier!

Here's the Terraform needed to redirect your root domain:

<<< @/public/blog/hosting-our-static-site/code/full.tf
