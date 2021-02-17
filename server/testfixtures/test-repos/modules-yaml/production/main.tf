module "null" {
  source = "../modules/null"
  var    = "production"
}
output "var" {
  value = module.null.var
}
