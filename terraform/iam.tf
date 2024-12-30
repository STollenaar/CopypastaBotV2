resource "aws_iam_role" "copypastabot_role" {
  name               = "CopypastabotRole"
  description        = "Role for the copypastabot"
  assume_role_policy = data.aws_iam_policy_document.assume_policy_document.json
}

resource "aws_iam_role_policy" "copypastabot_role_policy" {
  role   = aws_iam_role.copypastabot_role.id
  name   = "inline-role"
  policy = data.aws_iam_policy_document.ssm_access_role_policy_document.json
}
