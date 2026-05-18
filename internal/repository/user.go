package repository

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/haroldcamargo/auth-golang/internal/models"
)

func (db *DynamoDBClient) CreateUser(ctx context.Context, user *models.User) error {
	item, err := attributevalue.MarshalMap(user)
	if err != nil {
		return err
	}

	_, err = db.Client.PutItem(ctx, &dynamodb.PutItemInput{
		TableName:           aws.String(db.TableName),
		Item:                item,
		ConditionExpression: aws.String("attribute_not_exists(PK)"),
	})

	return err
}

func (db *DynamoDBClient) GetUserByEmail(ctx context.Context, email string) (*models.User, error) {
	pk := "USER#" + email
	sk := "METADATA"

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

	var user models.User
	if err := attributevalue.UnmarshalMap(out.Item, &user); err != nil {
		return nil, err
	}

	return &user, nil
}

func (db *DynamoDBClient) UpdateUserPassword(ctx context.Context, email, newPasswordHash string, updatedAt int64) error {
	pk := "USER#" + email
	sk := "METADATA"

	_, err := db.Client.UpdateItem(ctx, &dynamodb.UpdateItemInput{
		TableName: aws.String(db.TableName),
		Key: map[string]types.AttributeValue{
			"PK": &types.AttributeValueMemberS{Value: pk},
			"SK": &types.AttributeValueMemberS{Value: sk},
		},
		UpdateExpression: aws.String("SET password_hash = :p, updated_at = :u"),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":p": &types.AttributeValueMemberS{Value: newPasswordHash},
			":u": &types.AttributeValueMemberN{Value: fmt.Sprintf("%d", updatedAt)},
		},
	})
	return err
}
