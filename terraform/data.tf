data "terraform_remote_state" "discord_bots_cluster" {
  backend = "s3"
  config = {
    profile = local.used_profile.name
    region  = "ca-central-1"
    bucket  = "stollenaar-terraform-states"
    key     = "infrastructure/terraform.tfstate"
  }
}

data "terraform_remote_state" "statisticsbot" {
  backend = "s3"
  config = {
    profile = local.used_profile.name
    region  = "ca-central-1"
    bucket  = "stollenaar-terraform-states"
    key     = "discordbots/statisticsBot.tfstate"
  }
}

data "awsprofiler_list" "list_profiles" {}

data "aws_region" "current" {}

data "aws_caller_identity" "current" {}

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
      "ssm:GetParameter",
      "ssm:DescribeParameters",
    ]
    resources = ["*"]
  }
}

# IAM policy document for the container to access the sqs queue
data "aws_iam_policy_document" "sqs_role_policy_document" {
  statement {
    sid    = "SQSSendMessage"
    effect = "Allow"
    actions = [
      "sqs:DeleteMessage",
      "sqs:GetQueueAttributes",
      "sqs:GetQueueUrl",
      "sqs:ReceiveMessage",
      "sqs:SendMessage",
    ]
    resources = [
      data.terraform_remote_state.statisticsbot.outputs.sqs.request.arn,
      data.terraform_remote_state.statisticsbot.outputs.sqs.response.arn,
    ]
  }
}


# IAM policy document for the Lambda to access the parameter store
data "aws_iam_policy_document" "lambda_execution_invocation_document" {
  statement {
    sid    = "InvokeLambdas"
    effect = "Allow"
    actions = [
      "lambda:InvokeFunction",
      "lambda:InvokeAsync",
    ]
    resources = [
      "arn:aws:lambda:${data.aws_region.current.name}:${data.aws_caller_identity.current.account_id}:function:*",
    ]
  }
}
