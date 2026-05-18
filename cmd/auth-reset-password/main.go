wpackage main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/haroldcamargo/auth-golang/internal/auth"
	"github.com/haroldcamargo/auth-golang/internal/handler"
	"github.com/haroldcamargo/auth-golang/internal/repository"
)

type ResetPasswordRequest struct {
	ResetToken  string `json:"reset_token"`
	NewPassword string `json:"new_password"`
}

var dbClient *repository.DynamoDBClient
var jwtService *auth.JWTService

func init() {
	region := os.Getenv("AWS_REGION")
	if region == "" {
		region = "us-east-1"
	}
	tableName := os.Getenv("DYNAMODB_TABLE")
	if tableName == "" {
		tableName = "auth_service_store"
	}

	var err error
	dbClient, err = repository.NewDynamoDBClient(context.Background(), region, tableName)
	if err != nil {
		log.Fatalf("Failed to initialize DynamoDB client: %v", err)
	}

	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		jwtSecret = "default-secret-for-dev"
	}
	jwtService = auth.NewJWTService(jwtSecret)
}

func handleRequest(ctx context.Context, request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	var req ResetPasswordRequest
	if err := json.Unmarshal([]byte(request.Body), &req); err != nil {
		return handler.ErrorResponse(http.StatusBadRequest, "Invalid request body")
	}

	if len(req.NewPassword) < 8 {
		return handler.ErrorResponse(http.StatusBadRequest, "Password must be at least 8 characters")
	}

	claims, err := jwtService.ValidateToken(req.ResetToken)
	if err != nil {
		return handler.ErrorResponse(http.StatusUnauthorized, "Invalid or expired reset token")
	}

	userEmail := claims.Email

	// Hash new password
	hashedPassword, err := auth.HashPassword(req.NewPassword)
	if err != nil {
		return handler.ErrorResponse(http.StatusInternalServerError, "Internal server error")
	}

	now := time.Now().Unix()
	if err := dbClient.UpdateUserPassword(ctx, userEmail, hashedPassword, now); err != nil {
		return handler.ErrorResponse(http.StatusInternalServerError, "Internal server error")
	}

	return handler.JSONResponse(http.StatusOK, map[string]string{
		"message": "Password updated successfully",
	})
}

func main() {
	lambda.Start(handleRequest)
}
