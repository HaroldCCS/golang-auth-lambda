resource "aws_apigatewayv2_api" "auth_api" {
  name          = "auth-service-http-api"
  protocol_type = "HTTP"
}

resource "aws_apigatewayv2_stage" "prod" {
  api_id      = aws_apigatewayv2_api.auth_api.id
  name        = "$default"
  auto_deploy = true
}

resource "aws_apigatewayv2_integration" "lambda_integration" {
  for_each = aws_lambda_function.auth_functions

  api_id           = aws_apigatewayv2_api.auth_api.id
  integration_type = "AWS_PROXY"

  connection_type        = "INTERNET"
  description            = "Lambda integration for ${each.key}"
  integration_method     = "POST"
  integration_uri        = each.value.invoke_arn
  payload_format_version = "2.0"
}

locals {
  routes = {
    "auth-signup"         = "POST /signup"
    "auth-login"          = "POST /login"
    "auth-verify-otp"     = "POST /verify-otp"
    "auth-refresh-token"  = "POST /refresh-token"
    "auth-request-reset"  = "POST /request-reset"
    "auth-reset-password" = "POST /reset-password"
  }
}

resource "aws_apigatewayv2_route" "routes" {
  for_each = local.routes

  api_id    = aws_apigatewayv2_api.auth_api.id
  route_key = each.value
  target    = "integrations/${aws_apigatewayv2_integration.lambda_integration[each.key].id}"
}

resource "aws_lambda_permission" "apigw" {
  for_each = aws_lambda_function.auth_functions

  statement_id  = "AllowExecutionFromAPIGateway"
  action        = "lambda:InvokeFunction"
  function_name = each.value.function_name
  principal     = "apigateway.amazonaws.com"
  source_arn    = "${aws_apigatewayv2_api.auth_api.execution_arn}/*/*"
}

output "api_url" {
  value = aws_apigatewayv2_api.auth_api.api_endpoint
}
