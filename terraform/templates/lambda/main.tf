resource "null_resource" "go_build" {
  for_each = { for key, v in var.functions : key => v if length(regexall("go1.*|provided.al2", v.runtime)) > 0 }

  triggers = {
    "filehash"     = data.archive_file.lambda_source_zip[each.key].output_md5
    "buildcommand" = each.value.buildArgs
  }
  provisioner "local-exec" {
    working_dir = "${path.root}/../cmd/${each.key}"
    command     = "go mod tidy && CGO_ENABLED=0 GOOS=linux go build -o bootstrap -tags lambda.norpc ${each.value.buildArgs}"
  }
}

resource "null_resource" "docker_build" {
  for_each   = { for key, v in var.functions : key => v if length(regexall("go1.*|provided.al2", v.runtime)) > 0 && try(v.image_uri, null) != null }
  depends_on = [null_resource.go_build]

  triggers = {
    "filehash"      = data.archive_file.lambda_source_zip[each.key].output_md5
    "binary_hash"   = filemd5("${path.root}/../cmd/${each.key}/main")
    "lambda_docker" = filemd5("${path.module}/docker/lambda.Dockerfile")
    "app_docker"    = filemd5("${path.module}/docker/app.Dockerfile")
    "bootstrap"     = filemd5("${path.module}/docker/bootstrap.sh")
    "image_uri"     = each.value.image_uri
  }
  provisioner "local-exec" {
    command = "docker build -t ${lower(each.key)}-app --no-cache -f ${path.module}/docker/app.Dockerfile ${path.root}/../cmd/${each.key} && docker build --push --build-arg COMMAND=${lower(each.key)} --no-cache -t ${self.triggers.image_uri} -f ${path.module}/docker/lambda.Dockerfile ${path.module}/docker"
  }
}

resource "aws_cloudwatch_log_group" "lambda_function_log_group" {
  for_each          = var.functions
  name              = "/aws/lambda/${var.project}-${each.key}"
  retention_in_days = 7
}

resource "aws_lambda_function" "lambda_function" {
  depends_on = [
    data.archive_file.lambda_zip,
    null_resource.docker_build
  ]
  for_each = var.functions

  filename                       = try(each.value.image_uri, null) == null ? "${path.root}/../cmd/${each.key}/${each.key}.zip" : null
  image_uri                      = try(each.value.image_uri, null) != null ? "${split(":", each.value.image_uri)[0]}@${data.aws_ecr_image.service_image[each.key].id}" : null
  package_type                   = try(each.value.image_uri, null) == null ? "Zip" : "Image"
  description                    = each.value.description
  role                           = aws_iam_role.lambda_execution_role[each.key].arn
  function_name                  = "${var.project}-${each.key}"
  layers                         = concat(local.default_layers, try(each.value.layers, null))
  handler                        = try(each.value.image_uri, null) == null ? each.value.handler : null
  source_code_hash               = try(each.value.image_uri, null) == null ? try(each.value.override_zip_location, null) != null ? filebase64sha256(each.value.override_zip_location) : data.archive_file.lambda_zip[each.key].output_base64sha256 : null
  runtime                        = try(each.value.image_uri, null) == null ? each.value.runtime : null
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
