module "atlantis" {
  source  = "app.terraform.io/plentiau/ecr/aws"
  version = "1.0.1"

  ecr_repository_name = "buildtools/atlantis"
  allowed_read_account_ids = [
    data.aws_caller_identity.current.account_id,
    local.plentiau_security_account_id
  ]
  allowed_write_account_ids = [data.aws_caller_identity.current.account_id]
}
