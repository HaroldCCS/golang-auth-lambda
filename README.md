# 🔐 golang-auth-lambda

Servicio de autenticación **100 % serverless** escrito en **Go**: cada operación es una AWS Lambda independiente detrás de API Gateway, con DynamoDB como almacén y toda la infraestructura definida en **Terraform**.

![Go](https://img.shields.io/badge/Go-00ADD8?style=flat-square&logo=go&logoColor=white)
![AWS Lambda](https://img.shields.io/badge/AWS_Lambda-FF9900?style=flat-square&logo=awslambda&logoColor=white)
![DynamoDB](https://img.shields.io/badge/DynamoDB-4053D6?style=flat-square&logo=amazondynamodb&logoColor=white)
![Terraform](https://img.shields.io/badge/Terraform-7B42BC?style=flat-square&logo=terraform&logoColor=white)
![GitHub Actions](https://img.shields.io/badge/GitHub_Actions-2088FF?style=flat-square&logo=githubactions&logoColor=white)

## ✨ Qué hace

| Endpoint | Lambda | Descripción |
|---|---|---|
| `POST /signup` | `auth-signup` | Registro con contraseña hasheada con **Argon2id** |
| `POST /login` | `auth-login` | Login con access token (15 min) y refresh token (30 días) |
| `POST /verify-otp` | `auth-verify-otp` | Segundo factor (MFA) por OTP enviado al correo |
| `POST /refresh-token` | `auth-refresh-token` | Rotación de refresh tokens agrupados por familia de sesión |
| `POST /request-reset` | `auth-request-reset` | Solicitud de recuperación de contraseña por correo |
| `POST /reset-password` | `auth-reset-password` | Cambio de contraseña con el código de recuperación |

## 🏗️ Arquitectura

```
Cliente ──► API Gateway (HTTP API) ──► Lambda por endpoint (Go, arm64)
                                            │
                                            ├──► DynamoDB (single-table: usuarios, sesiones, OTP con TTL)
                                            ├──► SSM Parameter Store (JWT secret, credenciales SMTP)
                                            └──► SMTP (envío de OTP y recuperación)
```

- **Single-table design** en DynamoDB (`PK` / `SK`) con TTL sobre `expires_at` para que sesiones y OTP caduquen solos.
- **Secretos fuera del código**: se leen de SSM Parameter Store.
- **Binarios `arm64`** con `lambda.norpc`: arranques en frío más rápidos y más baratos.

## 📁 Estructura

```
cmd/                    # un main.go por Lambda
internal/
  auth/                 # Argon2id + JWT
  email/                # envío SMTP
  handler/              # respuestas HTTP comunes
  models/               # User, Session, OTP
  repository/           # acceso a DynamoDB
infrastructure/terraform/   # API Gateway, Lambdas, DynamoDB, IAM, SSM
.github/workflows/deploy.yml
```

## 🚀 Despliegue

Cada push a `main` ejecuta el pipeline de GitHub Actions:
1. `go test ./...`
2. Compila las 6 Lambdas para `linux/arm64` y las empaqueta en `.zip`
3. Se autentica en AWS con **OIDC** (sin llaves de larga duración)
4. `terraform apply` y actualización del código de cada función

Para desplegar desde tu máquina:

```bash
cd infrastructure/terraform
terraform init
terraform apply
```

> Antes de usarlo, reemplaza `YOUR_AWS_ACCOUNT_ID` y el rol de GitHub Actions en `.github/workflows/deploy.yml`.

---

Hecho por [Harold Camargo](https://haroldsoftware.com) · [LinkedIn](https://www.linkedin.com/in/haroldcamargo2002/)
