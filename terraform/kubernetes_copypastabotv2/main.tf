locals {
  name  = "copypastabotv2"
  image = var.docker_image != null ? var.docker_image : "${data.terraform_remote_state.discord_bots_cluster.outputs.discord_bots_repo.repository_url}:${local.name}-0.0.9-SNAPSHOT-79b344a"

  environment_variables = {
    AWS_PARAMETER_DISCORD_TOKEN        = "/discord_tokens/${local.name}"
    AWS_PARAMETER_REDDIT_USERNAME      = "/reddit/username"
    AWS_PARAMETER_REDDIT_PASSWORD      = "/reddit/password"
    AWS_PARAMETER_REDDIT_CLIENT_ID     = "/reddit/client_id"
    AWS_PARAMETER_REDDIT_CLIENT_SECRET = "/reddit/client_secret"
    AWS_PARAMETER_OPENAI_KEY           = "/openai/api_key"
    STATSBOT_URL                       = "statisticsbot.statisticsbot.svc.cluster.local"
    OLLAMA_URL                         = "ollama.ollama.svc.cluster.local:11434"
    AWS_OLLAMA_AUTH_USERNAME           = "/ollama/dan_username"
    AWS_OLLAMA_AUTH_PASSWORD           = "/ollama/dan_password"
  }
}

resource "kubernetes_deployment" "copypastabot" {
  metadata {
    name      = "copypastabotv2"
    namespace = data.terraform_remote_state.kubernetes.outputs.namespace.metadata.0.name
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
          "vault.hashicorp.com/aws-role"     = data.terraform_remote_state.iam_role.outputs.iam.copypastabot_role.name
          "cache.spicedelver.me/cmtemplate"  = "vault-aws-agent"
        }
        labels = {
          app = local.name
        }
      }
      spec {
        image_pull_secrets {
          name = data.terraform_remote_state.kubernetes.outputs.external_secret.spec.target.name
        }
        container {
          image = local.image
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
