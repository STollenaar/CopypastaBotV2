# resource "aws_sqs_queue" "browse_request" {
#   name                       = "${local.name}-browse-request"
#   message_retention_seconds  = 60 * 10
#   receive_wait_time_seconds  = 10
#   visibility_timeout_seconds = 60 * 5
# }

# resource "aws_sqs_queue" "speak_request" {
#   name                       = "${local.name}-speak-request"
#   message_retention_seconds  = 60 * 10
#   receive_wait_time_seconds  = 10
#   visibility_timeout_seconds = 60 * 10
# }

# resource "aws_sqs_queue" "chat_request" {
#   name                       = "${local.name}-chat-request"
#   message_retention_seconds  = 60 * 10
#   receive_wait_time_seconds  = 10
#   visibility_timeout_seconds = 60 * 10
# }

# resource "aws_sqs_queue" "help_request" {
#   name                       = "${local.name}-help-request"
#   message_retention_seconds  = 60 * 10
#   receive_wait_time_seconds  = 10
#   visibility_timeout_seconds = 60 * 10
# }

# resource "aws_lambda_permission" "sqs_receiver_lambda_invocation" {
#   statement_id  = "AllowExecutionFromSQS"
#   action        = "lambda:InvokeFunction"
#   function_name = module.lambda_functions.lambda_functions["sqsReceiver"].function_name
#   principal     = "sqs.amazonaws.com"

#   source_arn = data.terraform_remote_state.statisticsbot.outputs.sqs.response.arn
# }

# # Event source from SQS
# resource "aws_lambda_event_source_mapping" "sqs_receiver_lambda_source" {
#   event_source_arn = data.terraform_remote_state.statisticsbot.outputs.sqs.response.arn
#   enabled          = true
#   function_name    = module.lambda_functions.lambda_functions["sqsReceiver"].function_name
#   batch_size       = 1
# }

# resource "aws_lambda_permission" "browse_receiver_lambda_invocation" {
#   statement_id  = "AllowExecutionFromSQS"
#   action        = "lambda:InvokeFunction"
#   function_name = module.lambda_functions.lambda_functions["browseReceiver"].function_name
#   principal     = "sqs.amazonaws.com"

#   source_arn = aws_sqs_queue.browse_request.arn
# }

# # Event source from SQS
# resource "aws_lambda_event_source_mapping" "browse_receiver_lambda_source" {
#   event_source_arn = aws_sqs_queue.browse_request.arn
#   enabled          = true
#   function_name    = module.lambda_functions.lambda_functions["browseReceiver"].function_name
#   batch_size       = 1
# }

# resource "aws_lambda_permission" "speak_receiver_lambda_invocation" {
#   statement_id  = "AllowExecutionFromSQS"
#   action        = "lambda:InvokeFunction"
#   function_name = module.lambda_functions.lambda_functions["speakReceiver"].function_name
#   principal     = "sqs.amazonaws.com"

#   source_arn = aws_sqs_queue.speak_request.arn
# }

# # Event source from SQS
# resource "aws_lambda_event_source_mapping" "speak_receiver_lambda_source" {
#   event_source_arn = aws_sqs_queue.speak_request.arn
#   enabled          = true
#   function_name    = module.lambda_functions.lambda_functions["speakReceiver"].function_name
#   batch_size       = 1
# }

# resource "aws_lambda_permission" "chat_receiver_lambda_invocation" {
#   statement_id  = "AllowExecutionFromSQS"
#   action        = "lambda:InvokeFunction"
#   function_name = module.lambda_functions.lambda_functions["chatReceiver"].function_name
#   principal     = "sqs.amazonaws.com"

#   source_arn = aws_sqs_queue.chat_request.arn
# }

# # Event source from SQS
# resource "aws_lambda_event_source_mapping" "chat_receiver_lambda_source" {
#   event_source_arn = aws_sqs_queue.chat_request.arn
#   enabled          = true
#   function_name    = module.lambda_functions.lambda_functions["chatReceiver"].function_name
#   batch_size       = 1
# }

# # Event source from SQS
# resource "aws_lambda_event_source_mapping" "help_receiver_lambda_source" {
#   event_source_arn = aws_sqs_queue.help_request.arn
#   enabled          = true
#   function_name    = module.lambda_functions.lambda_functions["helpReceiver"].function_name
#   batch_size       = 1
# }

# resource "aws_lambda_permission" "help_receiver_lambda_invocation" {
#   statement_id  = "AllowExecutionFromSQS"
#   action        = "lambda:InvokeFunction"
#   function_name = module.lambda_functions.lambda_functions["helpReceiver"].function_name
#   principal     = "sqs.amazonaws.com"

#   source_arn = aws_sqs_queue.help_request.arn
# }

# # Event source from SQS
# resource "aws_lambda_event_source_mapping" "dunce_receiver_lambda_source" {
#   event_source_arn = data.terraform_remote_state.statisticsbot.outputs.sqs.request.arn
#   enabled          = true
#   function_name    = module.lambda_functions.lambda_functions["dunceReceiver"].function_name
#   batch_size       = 1
# }

# resource "aws_lambda_permission" "dunce_receiver_lambda_invocation" {
#   statement_id  = "AllowExecutionFromSQS"
#   action        = "lambda:InvokeFunction"
#   function_name = module.lambda_functions.lambda_functions["dunceReceiver"].function_name
#   principal     = "sqs.amazonaws.com"

#   source_arn = data.terraform_remote_state.statisticsbot.outputs.sqs.request.arn
# }
