package models

// Session represents a refresh token session in DynamoDB.
// PK: TOKEN#<refresh_jti>
// SK: SESSION
type Session struct {
	PK        string `dynamodbav:"PK"`
	SK        string `dynamodbav:"SK"`
	UserID    string `dynamodbav:"user_id"`
	FamilyID  string `dynamodbav:"family_id"`
	ExpiresAt int64  `dynamodbav:"expires_at"` // Used for TTL
}
