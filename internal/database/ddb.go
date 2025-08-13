package database

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
)

var ctx = context.Background()

type Client struct {
	client *dynamodb.Client
}

func NewDynamo(cfg *aws.Config) *Client {
	client := dynamodb.NewFromConfig(*cfg)

	return &Client{
		client: client,
	}
}

func (c *Client) Save(ctx context.Context, tableName string, data interface{}) error {
	item, err := attributevalue.MarshalMap(data)
	if err != nil {
		return err
	}
	_, err = c.client.PutItem(ctx, &dynamodb.PutItemInput{
		TableName: aws.String(tableName),
		Item:      item,
	})

	if err != nil {
		return err
	}

	return nil
}
