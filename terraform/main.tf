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

locals {
  name         = "copypastabot"
  used_profile = data.awsprofiler_list.list_profiles.profiles[try(index(data.awsprofiler_list.list_profiles.profiles.*.name, "personal"), 0)]
}


provider "aws" {
  profile = local.used_profile.name
}

resource "aws_ecs_service" "copypastaBot_service" {
  name            = local.name
  cluster         = data.terraform_remote_state.discord_bots_cluster.outputs.discord_bots_cluster.id
  task_definition = aws_ecs_task_definition.copypastaBot_service.arn
  desired_count   = 1

  capacity_provider_strategy {
    capacity_provider = data.terraform_remote_state.discord_bots_cluster.outputs.discord_bots_capacity_providers[0].name
    weight            = 100
  }

  service_connect_configuration {
    enabled   = true
    namespace = data.terraform_remote_state.discord_bots_cluster.outputs.discord_bots_namespace.arn
  }


  deployment_circuit_breaker {
    enable   = true
    rollback = true
  }
}

resource "aws_ecs_task_definition" "copypastaBot_service" {
  family                   = local.name
  requires_compatibilities = ["EC2"]
  execution_role_arn       = data.terraform_remote_state.discord_bots_cluster.outputs.spices_role.arn

  cpu          = 256
  memory       = 400
  network_mode = "bridge"

  runtime_platform {
    cpu_architecture        = "ARM64"
    operating_system_family = "LINUX"
  }
  container_definitions = jsonencode([
    {
      name      = local.name
      image     = "${data.terraform_remote_state.discord_bots_cluster.outputs.discord_bots_repo.repository_url}:${local.name}-latest-arm64"
      cpu       = 256
      memory    = 400
      essential = true

      environment = [
        {
          name  = "AWS_REGION"
          value = data.aws_region.current.name
        },
        {
          name  = "AWS_PARAMETER_NAME"
          value = "/discord_tokens/${local.name}"
        },
        {
          name  = "AWS_PARAMETER_REDDIT_USERNAME"
          value = "/reddit/username"
        },
        {
          name  = "AWS_PARAMETER_REDDIT_PASSWORD"
          value = "/reddit/password"
        },
        {
          name  = "AWS_PARAMETER_REDDIT_CLIENT_ID"
          value = "/reddit/client_id"
        },
        {
          name  = "AWS_PARAMETER_REDDIT_CLIENT_SECRET"
          value = "/reddit/client_secret"
        },
        {
          name  = "STATSBOT_URL"
          value = "statisticsbot"
        },
      ]
    }
  ])
}
