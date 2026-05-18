package main

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

type RequestResetRequest struct {
	Email string `json:"email"`
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
	var req RequestResetRequest
	if err := json.Unmarshal([]byte(request.Body), &req); err != nil {
		return handler.ErrorResponse(http.StatusBadRequest, "Invalid request body")
	}

	user, err := dbClient.GetUserByEmail(ctx, req.Email)
	if err != nil {
		return handler.ErrorResponse(http.StatusInternalServerError, "Internal server error")
	}

	// ALWAYS return success to prevent user enumeration
	if user == nil {
		return handler.JSONResponse(http.StatusOK, map[string]string{
			"message": "If that email is registered, you will receive reset instructions.",
		})
	}

	// Generate ephemeral recovery token (e.g., valid for 15 minutes)
	resetToken, err := jwtService.GenerateAccessToken(user.UUID, user.Email, 15*time.Minute)
	if err != nil {
		return handler.ErrorResponse(http.StatusInternalServerError, "Internal server error")
	}

	// TODO: Send resetToken via email
	// log.Printf("Reset Token (send via email): %s", resetToken)
	_ = resetToken

	return handler.JSONResponse(http.StatusOK, map[string]string{
		"message": "If that email is registered, you will receive reset instructions.",
	})
}

func main() {
	lambda.Start(handleRequest)
}
