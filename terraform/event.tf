# CloudWatch Event Rule
resource "aws_cloudwatch_event_rule" "schedule_rule" {
  name                = "daily_lambda_trigger"
  description         = "Trigger Lambda function daily"
  schedule_expression = "cron(30 8 * * ? *)"
}

# CloudWatch Event Target
resource "aws_cloudwatch_event_target" "lambda_target" {
  rule      = aws_cloudwatch_event_rule.schedule_rule.name
  target_id = "lambda_target"
  arn       = module.lambda_functions.lambda_functions["eggReceiver"].arn
}

# Lambda Permission
resource "aws_lambda_permission" "allow_cloudwatch" {
  statement_id  = "AllowExecutionFromCloudWatch"
  action        = "lambda:InvokeFunction"
  function_name = module.lambda_functions.lambda_functions["eggReceiver"].function_name
  principal     = "events.amazonaws.com"
  source_arn    = aws_cloudwatch_event_rule.schedule_rule.arn
}