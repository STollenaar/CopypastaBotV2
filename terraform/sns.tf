resource "aws_sns_topic" "router_sns" {
  name = "router-topic"
}

resource "aws_lambda_permission" "sns_lambda_invocation" {
  for_each = module.lambda_functions.lambda_functions

  function_name = each.value.function_name
  source_arn    = aws_sns_topic.router_sns.arn

  statement_id = "AllowExecutionFromSNS"
  action       = "lambda:InvokeFunction"
  principal    = "sns.amazonaws.com"
}

resource "aws_sns_topic_subscription" "router_sns_subscription" {
  for_each = module.lambda_functions.lambda_functions

  topic_arn = aws_sns_topic.router_sns.arn
  endpoint  = each.value.arn

  filter_policy = jsonencode({
    "function_name" = [each.key]
  })
  filter_policy_scope = "MessageAttributes"
  protocol            = "lambda"
}
