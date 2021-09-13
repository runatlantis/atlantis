module "null" {
  source = "../modules/null"
  var    = "staging"
}
output "var" {
  value = module.null.var
}
