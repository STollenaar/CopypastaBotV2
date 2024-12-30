
resource "vault_aws_secret_backend_role" "copypastabot_role" {
  backend         = data.terraform_remote_state.vault_setup.outputs.vault_aws_client
  name            = aws_iam_role.copypastabot_role.id
  credential_type = "assumed_role"
  role_arns       = [aws_iam_role.copypastabot_role.arn] #TODO: fetch dynamically
}
