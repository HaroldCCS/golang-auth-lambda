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
	"github.com/google/uuid"
	"github.com/haroldcamargo/auth-golang/internal/auth"
	"github.com/haroldcamargo/auth-golang/internal/handler"
	"github.com/haroldcamargo/auth-golang/internal/models"
	"github.com/haroldcamargo/auth-golang/internal/repository"
)

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
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
	var req LoginRequest
	if err := json.Unmarshal([]byte(request.Body), &req); err != nil {
		return handler.ErrorResponse(http.StatusBadRequest, "Invalid request body")
	}

	user, err := dbClient.GetUserByEmail(ctx, req.Email)
	if err != nil {
		log.Println("Error fetching user:", err)
		return handler.ErrorResponse(http.StatusInternalServerError, "Internal server error")
	}
	if user == nil {
		return handler.ErrorResponse(http.StatusUnauthorized, "Invalid credentials")
	}

	// Verify password
	valid, err := auth.VerifyPassword(req.Password, user.PasswordHash)
	if err != nil {
		log.Println("Error verifying password:", err)
		return handler.ErrorResponse(http.StatusInternalServerError, "Internal server error")
	}
	if !valid {
		return handler.ErrorResponse(http.StatusUnauthorized, "Invalid credentials")
	}

	if user.Status == "PENDING" {
		return handler.ErrorResponse(http.StatusForbidden, "Email not verified")
	}

	// MFA flow
	if user.IsMFAEnabled {
		mfaToken, err := jwtService.GenerateAccessToken(user.UUID, user.Email, 15*time.Minute)
		if err != nil {
			return handler.ErrorResponse(http.StatusInternalServerError, "Internal server error")
		}
		return handler.JSONResponse(http.StatusOK, map[string]interface{}{
			"mfa_required": true,
			"mfa_token":    mfaToken,
			"message":      "Please verify OTP sent to your email/phone",
		})
	}

	// Generate Access and Refresh Tokens
	accessToken, err := jwtService.GenerateAccessToken(user.UUID, user.Email, 15*time.Minute)
	if err != nil {
		return handler.ErrorResponse(http.StatusInternalServerError, "Internal server error")
	}

	refreshJTI := uuid.New().String()
	familyID := uuid.New().String()

	expiresAt := time.Now().Add(30 * 24 * time.Hour).Unix()
	
	session := &models.Session{
		PK:        "TOKEN#" + refreshJTI,
		SK:        "SESSION",
		UserID:    user.UUID,
		FamilyID:  familyID,
		ExpiresAt: expiresAt,
	}

	if err := dbClient.SaveSession(ctx, session); err != nil {
		log.Println("Error saving session:", err)
		return handler.ErrorResponse(http.StatusInternalServerError, "Internal server error")
	}

	refreshToken, _ := jwtService.GenerateAccessToken(refreshJTI, user.Email, 30*24*time.Hour)

	return handler.JSONResponse(http.StatusOK, map[string]interface{}{
		"access_token":  accessToken,
		"refresh_token": refreshToken,
		"expires_in":    900,
	})
}

func main() {
	lambda.Start(handleRequest)
}
