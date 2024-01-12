resource "null_resource" "go_build" {
  for_each = { for key, v in var.functions : key => v if length(regexall("go1.*|provided.al2", v.runtime)) > 0 }

  triggers = {
    "filehash" = data.archive_file.lambda_source_zip[each.key].output_md5
  }
  provisioner "local-exec" {
    working_dir = "${path.root}/../cmd/${each.key}"
    command     = "CGO_ENABLED=0 GOOS=linux go build -o bootstrap -tags lambda.norpc"
  }
}

resource "aws_cloudwatch_log_group" "lambda_function_log_group" {
  for_each          = var.functions
  name              = "/aws/lambda/${each.key}"
  retention_in_days = 7
}

resource "aws_lambda_function" "lambda_function" {
  depends_on = [
    data.archive_file.lambda_zip
  ]
  for_each = var.functions

  filename                       = "${path.root}/../cmd/${each.key}/${each.key}.zip"
  description                    = each.value.description
  role                           = aws_iam_role.lambda_execution_role[each.key].arn
  function_name                  = each.key
  layers                         = concat(local.default_layers, try(each.value.layers, null))
  handler                        = each.value.handler
  source_code_hash               = try(each.value.override_zip_location, null) != null ? filebase64sha256(each.value.override_zip_location) : data.archive_file.lambda_zip[each.key].output_base64sha256
  runtime                        = each.value.runtime
  timeout                        = try(each.value.timeout, null) != null ? each.value.timeout : null
  memory_size                    = try(each.value.memory_size, null) != null ? each.value.memory_size : null
  reserved_concurrent_executions = try(each.value.reserved_concurrent_executions, null)
  tags                           = try(each.value.tags, null) != null ? each.value.tags : {}

  dynamic "environment" {
    for_each = try(each.value.environment_variables, null) != null ? [each.value.environment_variables] : []
    content {
      variables = each.value.environment_variables
    }
  }
}

# data "aws_sns_topic" "lambda_errors" {
#   name = "lambda-errors"
# }

# resource "aws_cloudwatch_metric_alarm" "lambda_error_alarm" {
#   for_each   = { for key, v in var.functions : key => v if v.enable_alarm }
#   alarm_name = "lambda_error_${each.key}"
#   namespace  = "AWS/Lambda"
#   dimensions = {
#     "FunctionName" = each.key
#   }

#   alarm_description   = "Errors in lambda ${each.key}"
#   comparison_operator = "GreaterThanThreshold"
#   period              = "300"
#   evaluation_periods  = 1
#   metric_name         = "Errors"
#   threshold           = 0
#   statistic           = "Sum"
#   datapoints_to_alarm = 1
#   treat_missing_data  = "notBreaching"

#   alarm_actions = [
#     data.aws_sns_topic.lambda_errors.arn
#   ]
# }
