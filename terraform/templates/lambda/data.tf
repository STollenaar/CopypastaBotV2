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

data "aws_ecr_image" "service_image" {
  for_each   = { for key, v in var.functions : key => v if length(regexall("go1.*|provided.al2", v.runtime)) > 0 && try(v.image_uri, null) != null }
  depends_on = [null_resource.go_build, null_resource.docker_build]

  repository_name = "lambdas"
  image_tag       = lower(each.key)
}

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
      "arn:aws:logs:ca-central-1:${data.aws_caller_identity.current.account_id}:log-group:/aws/lambda/${var.project}-${each.key}*"
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
