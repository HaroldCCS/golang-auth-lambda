resource "aws_ssm_parameter" "jwt_secret" {
  name        = "/auth-service/jwt-secret"
  description = "JWT Secret Key for Auth Service"
  type        = "SecureString"
  value       = "CHANGE_ME_IN_CONSOLE" # Placeholder value
  
  lifecycle {
    ignore_changes = [value]
  }
}

resource "aws_ssm_parameter" "smtp_password" {
  name        = "/auth-service/smtp-password"
  description = "SMTP Password for sending emails"
  type        = "SecureString"
  value       = "CHANGE_ME_IN_CONSOLE" # Placeholder value
  
  lifecycle {
    ignore_changes = [value]
  }
}
