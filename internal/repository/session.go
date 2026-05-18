package repository

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/haroldcamargo/auth-golang/internal/models"
)

func (db *DynamoDBClient) SaveSession(ctx context.Context, session *models.Session) error {
	item, err := attributevalue.MarshalMap(session)
	if err != nil {
		return err
	}

	_, err = db.Client.PutItem(ctx, &dynamodb.PutItemInput{
		TableName: aws.String(db.TableName),
		Item:      item,
	})

	return err
}

func (db *DynamoDBClient) GetSession(ctx context.Context, refreshJTI string) (*models.Session, error) {
	pk := "TOKEN#" + refreshJTI
	sk := "SESSION"

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
		return nil, nil
	}

	var session models.Session
	if err := attributevalue.UnmarshalMap(out.Item, &session); err != nil {
		return nil, err
	}

	return &session, nil
}

func (db *DynamoDBClient) DeleteSession(ctx context.Context, refreshJTI string) error {
	pk := "TOKEN#" + refreshJTI
	sk := "SESSION"

	_, err := db.Client.DeleteItem(ctx, &dynamodb.DeleteItemInput{
		TableName: aws.String(db.TableName),
		Key: map[string]types.AttributeValue{
			"PK": &types.AttributeValueMemberS{Value: pk},
			"SK": &types.AttributeValueMemberS{Value: sk},
		},
	})

	return err
}
