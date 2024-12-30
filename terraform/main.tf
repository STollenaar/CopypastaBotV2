resource "kubernetes_namespace" "copypastabot" {
  metadata {
    name = local.name
  }
}


resource "kubernetes_deployment" "copypastabot" {
  metadata {
    name      = "copypastabot"
    namespace = kubernetes_namespace.copypastabot.metadata.0.name
    labels = {
      app = local.name
    }
  }
  spec {
    selector {
      match_labels = {
        app = local.name
      }
    }
    template {
      metadata {
        annotations = {
          "vault.hashicorp.com/agent-inject" = "true"
          "vault.hashicorp.com/role"         = "internal-app"
          "vault.hashicorp.com/aws-role"     = aws_iam_role.copypastabot_role.name
          "cache.spicedelver.me/cmtemplate"  = "vault-aws-agent"
        }
        labels = {
          app = local.name
        }
      }
      spec {
        image_pull_secrets {
          name = kubernetes_manifest.external_secret.manifest.spec.target.name
        }
        container {
          image = "${data.terraform_remote_state.discord_bots_cluster.outputs.discord_bots_repo.repository_url}:${local.name}-1.1.16-SNAPSHOT-26e1c53-amd64"
          name  = local.name
          env {
            name  = "AWS_REGION"
            value = data.aws_region.current.name
          }
          env {
            name  = "AWS_SHARED_CREDENTIALS_FILE"
            value = "/vault/secrets/aws/credentials"
          }
          dynamic "env" {
            for_each = local.environment_variables
            content {
              name  = env.key
              value = env.value
            }
          }
        }
      }
    }
  }
}


locals {
  name = "copypastabot"
  #   used_profile = data.awsprofiler_list.list_profiles.profiles[try(index(data.awsprofiler_list.list_profiles.profiles.*.name, "personal"), 0)]

  environment_variables = {
    AWS_PARAMETER_DISCORD_TOKEN        = "/discord_tokens/${local.name}"
    AWS_PARAMETER_REDDIT_USERNAME      = "/reddit/username"
    AWS_PARAMETER_REDDIT_PASSWORD      = "/reddit/password"
    AWS_PARAMETER_REDDIT_CLIENT_ID     = "/reddit/client_id"
    AWS_PARAMETER_REDDIT_CLIENT_SECRET = "/reddit/client_secret"
    OPENAI_KEY                         = "/openai/api_key"
    STATSBOT_URL                       = "statisticsbot-statisticsbot-ingress.tail88c07.ts.net"
  }
}

# module "lambda_functions" {
#   source    = "./templates/lambda"
#   functions = local.functions
#   project   = local.name
# }

# resource "aws_scheduler_schedule" "example" {
#   name       = "speak-interrupt"
#   group_name = "default"

#   flexible_time_window {
#     mode = "OFF"
#   }

#   schedule_expression = "cron(0 * ? * * *)"

#   target {
#     arn      = module.lambda_functions.lambda_functions["speakInterrupt"].arn
#     role_arn = aws_iam_role.scheduler.arn
#   }
# }

