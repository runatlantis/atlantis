module "pipeline_atlantis_role" {
  source  = "app.terraform.io/plentiau/iam-role-wizard/aws"
  version = "1.20.0"

  name              = "pipeline-atlantis-role"
  oidc_provider_arn = "arn:aws:iam::028287609508:oidc-provider/token.actions.githubusercontent.com"
  github_repo_names = ["atlantis:*"]
  allow_push_docker_image_to_ecr = [
    module.atlantis.ecr_arn
  ]
  allow_pull_docker_images_from_ecr = [
    module.atlantis.ecr_arn
  ]
}

output "pipeline_atlantis_role_arn" {
  value = module.pipeline_atlantis_role.iam_role_arn
}
