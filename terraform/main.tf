terraform {
  backend "s3" {
    region  = "ca-central-1"
    profile = "personal"
    bucket  = "stollenaar-terraform-states"
    key     = "discordbots/copypastabot.tfstate"
  }
  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 5.13.1"
    }
    awsprofiler = {
      version = "~> 0.0.1"
      source  = "spices.dev/stollenaar/awsprofiler"
    }
  }
  required_version = ">= 1.0.0"
}
provider "aws" {
  profile = local.used_profile.name
}

locals {
  name         = "copypastabot"
  used_profile = data.awsprofiler_list.list_profiles.profiles[try(index(data.awsprofiler_list.list_profiles.profiles.*.name, "personal"), 0)]

  functions = {
    browse = {
      description                    = "Browse command for CopypastaBot"
      enable_alarm                   = false
      runtime                        = "provided.al2"
      handler                        = "bootstrap"
      timeout                        = 3
      memory_size                    = 128
      layers                         = []
      reserved_concurrent_executions = -1
      extra_permissions              = [data.aws_iam_policy_document.lambda_execution_role_policy_document.json, data.aws_iam_policy_document.browse_sqs_role_policy_document.json]
      environment_variables = {
        AWS_SQS_URL = aws_sqs_queue.browse_request.url
      }
    }
    browseReceiver = {
      description                    = "browseReceiver for CopypastaBot"
      enable_alarm                   = false
      runtime                        = "provided.al2"
      handler                        = "bootstrap"
      timeout                        = 60 * 2
      memory_size                    = 128
      layers                         = []
      reserved_concurrent_executions = -1
      extra_permissions              = [data.aws_iam_policy_document.lambda_execution_role_policy_document.json, data.aws_iam_policy_document.browse_sqs_role_policy_document.json]
      environment_variables = {
        AWS_PARAMETER_REDDIT_USERNAME      = "/reddit/username"
        AWS_PARAMETER_REDDIT_PASSWORD      = "/reddit/password"
        AWS_PARAMETER_REDDIT_CLIENT_ID     = "/reddit/client_id"
        AWS_PARAMETER_REDDIT_CLIENT_SECRET = "/reddit/client_secret"
      }
    }
    chat = {
      description                    = "Chat command for CopypastaBot"
      enable_alarm                   = false
      runtime                        = "provided.al2"
      handler                        = "bootstrap"
      timeout                        = 3
      memory_size                    = 128
      layers                         = []
      reserved_concurrent_executions = -1
      extra_permissions              = [data.aws_iam_policy_document.lambda_execution_role_policy_document.json, data.aws_iam_policy_document.chat_sqs_role_policy_document.json]
      environment_variables = {
        AWS_SQS_URL = aws_sqs_queue.chat_request.url
      }
    }
    chatReceiver = {
      description                    = "Chat receiver for CopypastaBot"
      enable_alarm                   = false
      runtime                        = "provided.al2"
      handler                        = "bootstrap"
      timeout                        = 60 * 5
      memory_size                    = 128
      layers                         = []
      reserved_concurrent_executions = -1
      extra_permissions              = [data.aws_iam_policy_document.lambda_execution_role_policy_document.json, data.aws_iam_policy_document.chat_sqs_role_policy_document.json, data.aws_iam_policy_document.speak_sqs_role_policy_document.json]
      environment_variables = {
        OPENAI_KEY  = "/openai/api_key"
        AWS_SQS_URL = aws_sqs_queue.speak_request.url
      }
    }
    markov = {
      description                    = "Markov command for CopypastaBot"
      enable_alarm                   = false
      runtime                        = "provided.al2"
      handler                        = "bootstrap"
      timeout                        = 3
      memory_size                    = 128
      layers                         = []
      reserved_concurrent_executions = -1
      extra_permissions              = [data.aws_iam_policy_document.lambda_execution_role_policy_document.json, data.aws_iam_policy_document.sqs_role_policy_document.json]
      environment_variables = {
        AWS_SQS_URL = data.terraform_remote_state.statisticsbot.outputs.sqs.request.url
      }
    }
    pasta = {
      description                    = "Pasta command for CopypastaBot"
      enable_alarm                   = false
      runtime                        = "provided.al2"
      handler                        = "bootstrap"
      timeout                        = 3
      memory_size                    = 128
      layers                         = []
      reserved_concurrent_executions = -1
      extra_permissions              = [data.aws_iam_policy_document.lambda_execution_role_policy_document.json]
      environment_variables = {
        AWS_PARAMETER_REDDIT_USERNAME      = "/reddit/username"
        AWS_PARAMETER_REDDIT_PASSWORD      = "/reddit/password"
        AWS_PARAMETER_REDDIT_CLIENT_ID     = "/reddit/client_id"
        AWS_PARAMETER_REDDIT_CLIENT_SECRET = "/reddit/client_secret"
      }
    }
    ping = {
      description                    = "Ping command for CopypastaBot"
      enable_alarm                   = false
      runtime                        = "provided.al2"
      handler                        = "bootstrap"
      timeout                        = 3
      memory_size                    = 128
      layers                         = []
      reserved_concurrent_executions = -1
      extra_permissions              = [data.aws_iam_policy_document.lambda_execution_role_policy_document.json]
      environment_variables          = {}
    }
    sqsReceiver = {
      description                    = "sqs receiver for CopypastaBot"
      enable_alarm                   = false
      runtime                        = "provided.al2"
      handler                        = "bootstrap"
      timeout                        = 60 * 5
      memory_size                    = 2048
      layers                         = []
      reserved_concurrent_executions = -1
      extra_permissions              = [data.aws_iam_policy_document.lambda_execution_role_policy_document.json, data.aws_iam_policy_document.sqs_role_policy_document.json, data.aws_iam_policy_document.speak_sqs_role_policy_document.json]
      environment_variables = {
        AWS_SQS_URL = aws_sqs_queue.speak_request.url
      }
    }
    router = {
      description                    = "Router lambda to handle all commands correctly"
      enable_alarm                   = false
      runtime                        = "provided.al2"
      handler                        = "bootstrap"
      timeout                        = 5
      memory_size                    = 128
      layers                         = []
      reserved_concurrent_executions = -1
      extra_permissions              = [data.aws_iam_policy_document.lambda_execution_role_policy_document.json, data.aws_iam_policy_document.lambda_execution_invocation_document.json, data.aws_iam_policy_document.browse_sqs_role_policy_document.json]
      environment_variables = {
        AWS_PARAMETER_PUBLIC_DISCORD_TOKEN = "/discord_tokens/${local.name}_public",
        AWS_SQS_URL                        = aws_sqs_queue.browse_request.url
      }
    }
    speak = {
      description                    = "Speak command for CopypastaBot"
      enable_alarm                   = false
      runtime                        = "provided.al2"
      handler                        = "bootstrap"
      timeout                        = 3
      memory_size                    = 128
      layers                         = []
      reserved_concurrent_executions = -1
      extra_permissions              = [data.aws_iam_policy_document.lambda_execution_role_policy_document.json, data.aws_iam_policy_document.sqs_role_policy_document.json, data.aws_iam_policy_document.speak_sqs_role_policy_document.json, data.aws_iam_policy_document.chat_sqs_role_policy_document.json]
      environment_variables = {
        AWS_SQS_URL       = data.terraform_remote_state.statisticsbot.outputs.sqs.request.url
        AWS_SQS_URL_OTHER = "${aws_sqs_queue.speak_request.url};${aws_sqs_queue.chat_request.url}"
      }
    }
    speakInterrupt = {
      description                    = "Speak Interrupt for CopypastaBot"
      enable_alarm                   = false
      runtime                        = "provided.al2"
      handler                        = "bootstrap"
      timeout                        = 60 * 10
      memory_size                    = 512
      layers                         = []
      reserved_concurrent_executions = -1
      extra_permissions              = [data.aws_iam_policy_document.lambda_execution_role_policy_document.json, data.aws_iam_policy_document.sqs_role_policy_document.json]
      environment_variables = {
        DEBUG_GUILD                 = "544911814886948865"
        AWS_PARAMETER_DISCORD_TOKEN = "/discord_tokens/${local.name}"
        AWS_SQS_URL                 = data.terraform_remote_state.statisticsbot.outputs.sqs.request.url
      }
    }
    speakReceiver = {
      description                    = "Speak Receiver for CopypastaBot"
      enable_alarm                   = false
      runtime                        = "provided.al2"
      handler                        = "bootstrap"
      timeout                        = 60 * 10
      memory_size                    = 512
      layers                         = [data.aws_lambda_layer_version.ffmpeg_layer.arn]
      reserved_concurrent_executions = 1
      extra_permissions              = [data.aws_iam_policy_document.lambda_execution_role_policy_document.json, data.aws_iam_policy_document.speak_sqs_role_policy_document.json, data.aws_iam_policy_document.polly_role_policy_document.json]
      environment_variables = {
        AWS_PARAMETER_DISCORD_TOKEN        = "/discord_tokens/${local.name}"
        AWS_PARAMETER_REDDIT_USERNAME      = "/reddit/username"
        AWS_PARAMETER_REDDIT_PASSWORD      = "/reddit/password"
        AWS_PARAMETER_REDDIT_CLIENT_ID     = "/reddit/client_id"
        AWS_PARAMETER_REDDIT_CLIENT_SECRET = "/reddit/client_secret"
      }
    }
  }
}

module "lambda_functions" {
  source    = "./templates/lambda"
  functions = local.functions
}

resource "aws_scheduler_schedule" "example" {
  name       = "speak-interrupt"
  group_name = "default"

  flexible_time_window {
    mode = "OFF"
  }

  schedule_expression = "rate(1 hours)"

  target {
    arn      = module.lambda_functions.lambda_functions["speakInterrupt"].arn
    role_arn = aws_iam_role.scheduler.arn
  }
}

