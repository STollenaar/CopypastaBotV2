# Generated by Terragrunt. Sig: nIlQXj57tbuaRZEa
terraform {
  backend "s3" {
    bucket  = "stollenaar-terraform-states"
    encrypt = true
    key     = "discordbots/copypastabot/terraform.tfstate"
    region  = "ca-central-1"
  }
}
