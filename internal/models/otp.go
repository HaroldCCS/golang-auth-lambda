package models

// OTP represents an OTP record for a user in DynamoDB.
// PK: USER#<email>
// SK: OTP#CURRENT
type OTP struct {
	PK        string `dynamodbav:"PK"`
	SK        string `dynamodbav:"SK"`
	OTPHash   string `dynamodbav:"otp_hash"`
	ExpiresAt int64  `dynamodbav:"expires_at"` // Used for TTL
}
