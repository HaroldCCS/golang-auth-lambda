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

type RefreshRequest struct {
	RefreshToken string `json:"refresh_token"`
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
	var req RefreshRequest
	if err := json.Unmarshal([]byte(request.Body), &req); err != nil {
		return handler.ErrorResponse(http.StatusBadRequest, "Invalid request body")
	}

	claims, err := jwtService.ValidateToken(req.RefreshToken)
	if err != nil {
		return handler.ErrorResponse(http.StatusUnauthorized, "Invalid refresh token")
	}

	oldJTI := claims.UserID // Subject of refresh token is JTI
	userEmail := claims.Email

	// Fetch Session
	session, err := dbClient.GetSession(ctx, oldJTI)
	if err != nil || session == nil {
		return handler.ErrorResponse(http.StatusUnauthorized, "Invalid or expired session")
	}

	// Delete old session (Rotation)
	_ = dbClient.DeleteSession(ctx, oldJTI)

	// Create new Session
	newJTI := uuid.New().String()
	newSession := &models.Session{
		PK:        "TOKEN#" + newJTI,
		SK:        "SESSION",
		UserID:    session.UserID,
		FamilyID:  session.FamilyID,
		ExpiresAt: time.Now().Add(30 * 24 * time.Hour).Unix(),
	}

	if err := dbClient.SaveSession(ctx, newSession); err != nil {
		return handler.ErrorResponse(http.StatusInternalServerError, "Internal server error")
	}

	// Generate new tokens
	newAccessToken, _ := jwtService.GenerateAccessToken(session.UserID, userEmail, 15*time.Minute)
	newRefreshToken, _ := jwtService.GenerateAccessToken(newJTI, userEmail, 30*24*time.Hour)

	return handler.JSONResponse(http.StatusOK, map[string]interface{}{
		"access_token":  newAccessToken,
		"refresh_token": newRefreshToken,
		"expires_in":    900,
	})
}

func main() {
	lambda.Start(handleRequest)
}
