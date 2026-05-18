package main

import (
	"context"
	"encoding/json"
	"fmt"
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

type SignupRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

var dbClient *repository.DynamoDBClient

func init() {
	// Initialize DynamoDB client on cold start
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
}

func handleRequest(ctx context.Context, request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	var req SignupRequest
	if err := json.Unmarshal([]byte(request.Body), &req); err != nil {
		return handler.ErrorResponse(http.StatusBadRequest, "Invalid request body")
	}

	if req.Email == "" || len(req.Password) < 8 {
		return handler.ErrorResponse(http.StatusBadRequest, "Invalid email or weak password")
	}

	// Check if user exists
	existingUser, err := dbClient.GetUserByEmail(ctx, req.Email)
	if err != nil {
		log.Println("Error fetching user:", err)
		return handler.ErrorResponse(http.StatusInternalServerError, "Internal server error")
	}
	if existingUser != nil {
		return handler.ErrorResponse(http.StatusConflict, "User already exists")
	}

	// Hash password
	hashedPassword, err := auth.HashPassword(req.Password)
	if err != nil {
		log.Println("Error hashing password:", err)
		return handler.ErrorResponse(http.StatusInternalServerError, "Internal server error")
	}

	// Store user
	now := time.Now().Unix()
	user := &models.User{
		PK:           "USER#" + req.Email,
		SK:           "METADATA",
		UUID:         uuid.New().String(),
		Email:        req.Email,
		PasswordHash: hashedPassword,
		Status:       "PENDING", // Wait for email verification
		CreatedAt:    now,
		UpdatedAt:    now,
	}

	if err := dbClient.CreateUser(ctx, user); err != nil {
		log.Println("Error creating user:", err)
		return handler.ErrorResponse(http.StatusInternalServerError, "Internal server error")
	}

	// TODO: Send verification email with OTP or Link here

	return handler.JSONResponse(http.StatusCreated, map[string]string{
		"message": "User created successfully. Please verify your email.",
		"user_id": user.UUID,
	})
}

func main() {
	lambda.Start(handleRequest)
}
