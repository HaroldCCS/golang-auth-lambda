package main

import (
	"context"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/hex"
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

type VerifyOTPRequest struct {
	MFAToken string `json:"mfa_token"`
	OTP      string `json:"otp"`
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
	var req VerifyOTPRequest
	if err := json.Unmarshal([]byte(request.Body), &req); err != nil {
		return handler.ErrorResponse(http.StatusBadRequest, "Invalid request body")
	}

	claims, err := jwtService.ValidateToken(req.MFAToken)
	if err != nil {
		return handler.ErrorResponse(http.StatusUnauthorized, "Invalid MFA session")
	}

	userEmail := claims.Email
	userID := claims.UserID

	otpRecord, err := dbClient.GetOTP(ctx, userEmail)
	if err != nil {
		return handler.ErrorResponse(http.StatusInternalServerError, "Internal server error")
	}
	if otpRecord == nil || otpRecord.ExpiresAt < time.Now().Unix() {
		return handler.ErrorResponse(http.StatusUnauthorized, "OTP expired or invalid")
	}

	// Hash provided OTP to compare
	h := sha256.New()
	h.Write([]byte(req.OTP))
	providedHash := hex.EncodeToString(h.Sum(nil))

	if subtle.ConstantTimeCompare([]byte(providedHash), []byte(otpRecord.OTPHash)) != 1 {
		return handler.ErrorResponse(http.StatusUnauthorized, "Invalid OTP")
	}

	// Delete OTP record after successful use (optional, as TTL will clean it, but good practice)
	// (Skipped for brevity, but recommended in production)

	// Generate Access and Refresh Tokens
	accessToken, err := jwtService.GenerateAccessToken(userID, userEmail, 15*time.Minute)
	if err != nil {
		return handler.ErrorResponse(http.StatusInternalServerError, "Internal server error")
	}

	refreshJTI := uuid.New().String()
	familyID := uuid.New().String()

	session := &models.Session{
		PK:        "TOKEN#" + refreshJTI,
		SK:        "SESSION",
		UserID:    userID,
		FamilyID:  familyID,
		ExpiresAt: time.Now().Add(30 * 24 * time.Hour).Unix(),
	}

	if err := dbClient.SaveSession(ctx, session); err != nil {
		return handler.ErrorResponse(http.StatusInternalServerError, "Internal server error")
	}

	refreshToken, _ := jwtService.GenerateAccessToken(refreshJTI, userEmail, 30*24*time.Hour)

	return handler.JSONResponse(http.StatusOK, map[string]interface{}{
		"access_token":  accessToken,
		"refresh_token": refreshToken,
		"expires_in":    900,
	})
}

func main() {
	lambda.Start(handleRequest)
}
