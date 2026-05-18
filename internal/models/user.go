package models

// User represents the metadata of a user in DynamoDB.
// PK: USER#<email>
// SK: METADATA
type User struct {
	PK           string `dynamodbav:"PK"`
	SK           string `dynamodbav:"SK"`
	UUID         string `dynamodbav:"uuid"`
	Email        string `dynamodbav:"email"`
	PasswordHash string `dynamodbav:"password_hash"`
	IsMFAEnabled bool   `dynamodbav:"is_mfa_enabled"`
	MFASecret    string `dynamodbav:"mfa_secret,omitempty"`
	Phone        string `dynamodbav:"phone,omitempty"`
	Status       string `dynamodbav:"status"` // PENDING, ACTIVE, SUSPENDED
	CreatedAt    int64  `dynamodbav:"created_at"`
	UpdatedAt    int64  `dynamodbav:"updated_at"`
}
