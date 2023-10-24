locals {
  files = {
    for key, v in var.functions : key => {
      files    = fileset("${path.root}/../cmd/${key}", "*")
      zips     = [for f in fileset("${path.root}/../cmd/${key}", "*") : f if endswith(f, ".zip")]
      go_files = [for f in fileset("${path.root}/../cmd/${key}", "*") : f if endswith(f, ".go") || endswith(f, ".mod") || endswith(f, ".sum")]
    }
  }
}

data "archive_file" "lambda_source_zip" {
  for_each = var.functions

  type        = "zip"
  source_dir  = "${path.root}/../cmd/${each.key}"
  output_path = "${path.root}/../cmd/${each.key}/${each.key}_source.zip"
  excludes    = setsubtract(local.files[each.key].files, local.files[each.key].go_files)
}

data "archive_file" "lambda_zip" {
  depends_on = [
    null_resource.go_build
  ]
  for_each = var.functions

  type        = "zip"
  output_path = "${path.root}/../cmd/${each.key}/${each.key}.zip"
  source_file = "${path.root}/../cmd/${each.key}/bootstrap"
}

data "aws_iam_policy_document" "lambda_execution_role_document" {
  statement {
    actions = ["sts:AssumeRole"]
    principals {
      type        = "Service"
      identifiers = ["lambda.amazonaws.com"]
    }
  }
}

data "aws_caller_identity" "current" {}

# IAM policy document for the Lambda to access S3
data "aws_iam_policy_document" "lambda_execution_role_policy_document" {
  for_each = var.functions

  statement {
    sid    = "CloudwatchLogGroup"
    effect = "Allow"
    actions = [
      "logs:PutLogEvents",
      "logs:CreateLogStream",
      "logs:CreateLogGroup"
    ]
    resources = [
      "arn:aws:logs:ca-central-1:${data.aws_caller_identity.current.account_id}:log-group:/aws/lambda/${each.key}*"
    ]
  }
  statement {
    sid    = "CloudwatchLogGroupDescribe"
    effect = "Allow"
    actions = [
      "logs:DescribeLogGroups"
    ]
    resources = [
      "*"
    ]
  }
  source_policy_documents = try(each.value.extra_permissions, [])
}
