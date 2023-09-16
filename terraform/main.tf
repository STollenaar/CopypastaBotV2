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
      description       = "Browse command for CopypastaBot"
      enable_alarm      = false
      runtime           = "provided.al2"
      handler           = "bootstrap"
      timeout           = 60 * 2
      extra_permissions = [data.aws_iam_policy_document.lambda_execution_role_policy_document.json]
      environment_variables = {
        AWS_PARAMETER_NAME                 = "/discord_tokens/${local.name}"
        AWS_PARAMETER_REDDIT_USERNAME      = "/reddit/username"
        AWS_PARAMETER_REDDIT_PASSWORD      = "/reddit/password"
        AWS_PARAMETER_REDDIT_CLIENT_ID     = "/reddit/client_id"
        AWS_PARAMETER_REDDIT_CLIENT_SECRET = "/reddit/client_secret"
        STATSBOT_URL                       = "statisticsbot"
      }
    }
    markov = {
      description       = "Markov command for CopypastaBot"
      enable_alarm      = false
      runtime           = "provided.al2"
      handler           = "bootstrap"
      timeout           = 60 * 2
      extra_permissions = [data.aws_iam_policy_document.lambda_execution_role_policy_document.json]
      environment_variables = {
        AWS_PARAMETER_NAME                 = "/discord_tokens/${local.name}"
        AWS_PARAMETER_REDDIT_USERNAME      = "/reddit/username"
        AWS_PARAMETER_REDDIT_PASSWORD      = "/reddit/password"
        AWS_PARAMETER_REDDIT_CLIENT_ID     = "/reddit/client_id"
        AWS_PARAMETER_REDDIT_CLIENT_SECRET = "/reddit/client_secret"
        STATSBOT_URL                       = "statisticsbot"
      }
    }
    pasta = {
      description       = "Pasta command for CopypastaBot"
      enable_alarm      = false
      runtime           = "provided.al2"
      handler           = "bootstrap"
      timeout           = 60 * 2
      extra_permissions = [data.aws_iam_policy_document.lambda_execution_role_policy_document.json]
      environment_variables = {
        AWS_PARAMETER_NAME                 = "/discord_tokens/${local.name}"
        AWS_PARAMETER_REDDIT_USERNAME      = "/reddit/username"
        AWS_PARAMETER_REDDIT_PASSWORD      = "/reddit/password"
        AWS_PARAMETER_REDDIT_CLIENT_ID     = "/reddit/client_id"
        AWS_PARAMETER_REDDIT_CLIENT_SECRET = "/reddit/client_secret"
        STATSBOT_URL                       = "statisticsbot"
      }
    }
    ping = {
      description       = "Ping command for CopypastaBot"
      enable_alarm      = false
      runtime           = "provided.al2"
      handler           = "bootstrap"
      timeout           = 60 * 2
      extra_permissions = [data.aws_iam_policy_document.lambda_execution_role_policy_document.json]
      environment_variables = {
        AWS_PARAMETER_NAME                 = "/discord_tokens/${local.name}"
        AWS_PARAMETER_REDDIT_USERNAME      = "/reddit/username"
        AWS_PARAMETER_REDDIT_PASSWORD      = "/reddit/password"
        AWS_PARAMETER_REDDIT_CLIENT_ID     = "/reddit/client_id"
        AWS_PARAMETER_REDDIT_CLIENT_SECRET = "/reddit/client_secret"
        STATSBOT_URL                       = "statisticsbot"
      }
    }
    speak = {
      description       = "Speak command for CopypastaBot"
      enable_alarm      = false
      runtime           = "provided.al2"
      handler           = "bootstrap"
      timeout           = 60 * 2
      extra_permissions = [data.aws_iam_policy_document.lambda_execution_role_policy_document.json]
      environment_variables = {
        AWS_PARAMETER_NAME                 = "/discord_tokens/${local.name}"
        AWS_PARAMETER_REDDIT_USERNAME      = "/reddit/username"
        AWS_PARAMETER_REDDIT_PASSWORD      = "/reddit/password"
        AWS_PARAMETER_REDDIT_CLIENT_ID     = "/reddit/client_id"
        AWS_PARAMETER_REDDIT_CLIENT_SECRET = "/reddit/client_secret"
        STATSBOT_URL                       = "statisticsbot"
      }
    }
  }
}

module "lambda_functions" {
  source    = "./templates/lambda"
  functions = local.functions
}
