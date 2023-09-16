resource "aws_iam_role" "lambda_execution_role" {
  for_each = var.functions

  name = "${each.key}LambdaRole"

  assume_role_policy = data.aws_iam_policy_document.lambda_execution_role_document.json
}

resource "aws_iam_role_policy" "lambda_execution_role_policy" {
  for_each = var.functions
  role     = aws_iam_role.lambda_execution_role[each.key].id
  name     = "lambda-inline-policy"
  policy   = data.aws_iam_policy_document.lambda_execution_role_policy_document[each.key].json
}
