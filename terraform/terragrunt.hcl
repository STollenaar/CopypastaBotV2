locals {
  name            = "copypastabot"
  kubeconfig_file = "/home/stollenaar/.kube/config"

  # Automatically load provider variables
  provider_vars = read_terragrunt_config("./provider.hcl")

  # Extract the variables we need for easy access
  providers = local.provider_vars.locals.providers
}

remote_state {
  backend = "s3"
  generate = {
    path      = "grunt_backend.tf"
    if_exists = "overwrite_terragrunt"
  }
  config = {
    bucket = "stollenaar-terraform-states"

    key     = "discordbots/${local.name}/terraform.tfstate"
    region  = "ca-central-1"
    encrypt = true
  }
}

generate "provider" {
  path      = "grunt_providers.tf"
  if_exists = "overwrite"
  contents  = <<EOF
        %{if contains(local.providers, "kubernetes")}
        provider "kubernetes" {
            config_path = "${local.kubeconfig_file}"
        }
        %{endif}
        %{if contains(local.providers, "hcp")}
        provider "hcp" {
            client_id     = data.aws_ssm_parameter.vault_client_id.value
            client_secret = data.aws_ssm_parameter.vault_client_secret.value
        }
        %{endif}
        %{if contains(local.providers, "vault")}
        provider "vault" {
            token   = data.hcp_vault_secrets_secret.vault_root.secret_value
            address = "http://localhost:8200"
        }
        %{endif}
    EOF
}

terraform_binary = "/usr/local/bin/tofu"

generate "versions" {
  path      = "grunt_versions.tf"
  if_exists = "overwrite_terragrunt"
  contents  = <<EOF
    terraform {
        required_providers {
            %{if contains(local.providers, "aws")}
            aws = {
            source  = "hashicorp/aws"
            version = "~> 5.24.0"
            }
            %{endif}
            %{if contains(local.providers, "hcp")}
            hcp = {
            version = "~> 0.76.0"
            source  = "hashicorp/hcp"
            }
            %{endif}
            %{if contains(local.providers, "kubernetes")}
            kubernetes = {
            version = "~> 2.23.0"
            source  = "hashicorp/kubernetes"
            }
            %{endif}
            %{if contains(local.providers, "vault")}
            vault = {
                source  = "hashicorp/vault"
                version = "~> 3.21.0"
            }
            %{endif}
        }
        required_version = ">= 1.2.2"
    }
    EOF
}

terraform {
  before_hook "before_hook" {
    commands = ["apply", "destroy", "plan"]
    execute  = ["./conf/start_service.sh", local.kubeconfig_file]
    # get_env("KUBES_ENDPOINT", "somedefaulturl") you can get env vars or pass in params as needed to the script
  }

  after_hook "after_hook" {
    commands     = ["apply", "destroy", "plan"]
    execute      = ["./conf/stop_service.sh"]
    run_on_error = true
  }
  extra_arguments "common_vars" {
    commands = get_terraform_commands_that_need_vars()
    arguments = [
      "-var=kubeconfig_file=${local.kubeconfig_file}"
    ]y
  }
}