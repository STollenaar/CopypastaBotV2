output "lambda_functions" {
  value = tomap({ for k, v in aws_lambda_function.lambda_function : k => v })
}
