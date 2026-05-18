package repository

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/haroldcamargo/auth-golang/internal/models"
)

func (db *DynamoDBClient) SaveOTP(ctx context.Context, otp *models.OTP) error {
	item, err := attributevalue.MarshalMap(otp)
	if err != nil {
		return err
	}

	_, err = db.Client.PutItem(ctx, &dynamodb.PutItemInput{
		TableName: aws.String(db.TableName),
		Item:      item,
	})

	return err
}

func (db *DynamoDBClient) GetOTP(ctx context.Context, email string) (*models.OTP, error) {
	pk := "USER#" + email
	sk := "OTP#CURRENT"

	out, err := db.Client.GetItem(ctx, &dynamodb.GetItemInput{
		TableName: aws.String(db.TableName),
		Key: map[string]types.AttributeValue{
			"PK": &types.AttributeValueMemberS{Value: pk},
			"SK": &types.AttributeValueMemberS{Value: sk},
		},
	})
	if err != nil {
		return nil, err
	}

	if len(out.Item) == 0 {
		return nil, nil // Not found
	}

	var otp models.OTP
	if err := attributevalue.UnmarshalMap(out.Item, &otp); err != nil {
		return nil, err
	}

	return &otp, nil
}
