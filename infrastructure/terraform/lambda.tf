locals {
  lambdas = [
    "auth-signup",
    "auth-login",
    "auth-verify-otp",
    "auth-refresh-token",
    "auth-request-reset",
    "auth-reset-password"
  ]
}

resource "aws_lambda_function" "auth_functions" {
  for_each = toset(local.lambdas)

  function_name = each.key
  role          = aws_iam_role.auth_lambda_role.arn
  handler       = "bootstrap" # Required for provided.al2023 / Graviton
  runtime       = "provided.al2023"
  architectures = ["arm64"]
  
  timeout       = 10 # seconds

  filename         = "dummy.zip"
  source_code_hash = fileexists("dummy.zip") ? filebase64sha256("dummy.zip") : null

  environment {
    variables = {
      DYNAMODB_TABLE = aws_dynamodb_table.auth_service_store.name
      AWS_REGION     = var.aws_region
      JWT_SECRET     = aws_ssm_parameter.jwt_secret.value
    }
  }

  lifecycle {
    ignore_changes = [
      filename,
      source_code_hash,
      environment # so if we change JWT_SECRET in console, Terraform doesn't overwrite it immediately
    ]
  }
}
