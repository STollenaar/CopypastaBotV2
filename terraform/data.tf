data "terraform_remote_state" "discord_bots_cluster" {
  backend = "s3"
  config = {
    profile = local.used_profile.name
    region  = "ca-central-1"
    bucket  = "stollenaar-terraform-states"
    key     = "infrastructure/terraform.tfstate"
  }
}
data "awsprofiler_list" "list_profiles" {}

data "aws_region" "current" {}

# IAM policy document for the Lambda to access the parameter store
data "aws_iam_policy_document" "lambda_execution_role_policy_document" {
  statement {
    sid    = "KMSDecryption"
    effect = "Allow"
    actions = [
      "kms:ListKeys",
      "kms:GetPublicKey",
      "kms:DescribeKey",
      "kms:Decrypt",
    ]
    resources = [
      "*"
    ]
  }
  statement {
    sid    = "SSMAccess"
    effect = "Allow"
    actions = [
      "ssm:GetParametersByPath",
      "ssm:GetParameters",
      "ssm:DescribeParameters",
    ]
    resources = ["*"]
  }
}
