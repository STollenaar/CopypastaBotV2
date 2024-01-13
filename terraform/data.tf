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

data "aws_lambda_layer_version" "ffmpeg_layer" {
  layer_name = "ffmpeg"
}

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
# IAM policy document for the container to access the sqs queue
data "aws_iam_policy_document" "browse_sqs_role_policy_document" {
  statement {
    sid    = "SQSBrowseSendMessage"
    effect = "Allow"
    actions = [
      "sqs:DeleteMessage",
      "sqs:GetQueueAttributes",
      "sqs:GetQueueUrl",
      "sqs:ReceiveMessage",
      "sqs:SendMessage",
    ]
    resources = [
      aws_sqs_queue.browse_request.arn
    ]
  }
}
# IAM policy document for the container to access the sqs queue
data "aws_iam_policy_document" "speak_sqs_role_policy_document" {
  statement {
    sid    = "SQSSpeakSendMessage"
    effect = "Allow"
    actions = [
      "sqs:DeleteMessage",
      "sqs:GetQueueAttributes",
      "sqs:GetQueueUrl",
      "sqs:ReceiveMessage",
      "sqs:SendMessage",
    ]
    resources = [
      aws_sqs_queue.speak_request.arn
    ]
  }
}

# IAM policy document for the container to access the sqs queue
data "aws_iam_policy_document" "chat_sqs_role_policy_document" {
  statement {
    sid    = "SQSChatSendMessage"
    effect = "Allow"
    actions = [
      "sqs:DeleteMessage",
      "sqs:GetQueueAttributes",
      "sqs:GetQueueUrl",
      "sqs:ReceiveMessage",
      "sqs:SendMessage",
    ]
    resources = [
      aws_sqs_queue.chat_request.arn
    ]
  }
}

# IAM policy document for the container to access the sqs queue
data "aws_iam_policy_document" "help_sqs_role_policy_document" {
  statement {
    sid    = "SQSHelpSendMessage"
    effect = "Allow"
    actions = [
      "sqs:DeleteMessage",
      "sqs:GetQueueAttributes",
      "sqs:GetQueueUrl",
      "sqs:ReceiveMessage",
      "sqs:SendMessage",
    ]
    resources = [
      aws_sqs_queue.help_request.arn
    ]
  }
}

# IAM policy document for the container to access the sqs queue
data "aws_iam_policy_document" "polly_role_policy_document" {
  statement {
    sid    = "PollySynthSpeech"
    effect = "Allow"
    actions = [
      "polly:SynthesizeSpeech",
    ]
    resources = [
      "*"
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

data "aws_iam_policy_document" "event_bridge_execution_role_document" {
  statement {
    actions = ["sts:AssumeRole"]
    principals {
      type        = "Service"
      identifiers = ["scheduler.amazonaws.com"]
    }
  }
}
