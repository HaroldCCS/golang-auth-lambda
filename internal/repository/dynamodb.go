package repository

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
)

type DynamoDBClient struct {
	Client    *dynamodb.Client
	TableName string
}

func NewDynamoDBClient(ctx context.Context, region, tableName string) (*DynamoDBClient, error) {
	cfg, err := config.LoadDefaultConfig(ctx, config.WithRegion(region))
	if err != nil {
		return nil, fmt.Errorf("unable to load SDK config, %w", err)
	}

	client := dynamodb.NewFromConfig(cfg)

	return &DynamoDBClient{
		Client:    client,
		TableName: tableName,
	}, nil
}
