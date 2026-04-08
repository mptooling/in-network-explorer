//go:build integration

package infrastructure

import (
	"context"
	"os"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/secretsmanager"
)

func integrationConfig(t *testing.T) Config {
	t.Helper()
	cfg := Config{
		AWSRegion:       envOrSkip(t, "AWS_REGION"),
		DynamoTableName: envOrSkip(t, "DYNAMO_TABLE"),
		DynamoEndpoint:  os.Getenv("DYNAMO_ENDPOINT"),
		SecretsEndpoint: os.Getenv("SECRETS_ENDPOINT"),
	}
	cfg.LinkedInCookiesSecret = envOrSkip(t, "LINKEDIN_COOKIES_SECRET")
	cfg.BedrockRegion = envOrDefault("BEDROCK_REGION", cfg.AWSRegion)
	return cfg
}

func envOrSkip(t *testing.T, key string) string {
	t.Helper()
	v := os.Getenv(key)
	if v == "" {
		t.Skipf("%s not set, skipping integration test", key)
	}
	return v
}

func TestNewDynamoClient_Connect(t *testing.T) {
	cfg := integrationConfig(t)

	client, err := NewDynamoClient(context.Background(), cfg)
	if err != nil {
		t.Fatalf("NewDynamoClient() error = %v", err)
	}

	out, err := client.ListTables(context.Background(), &dynamodb.ListTablesInput{
		Limit: aws.Int32(1),
	})
	if err != nil {
		t.Fatalf("ListTables() error = %v", err)
	}
	t.Logf("DynamoDB connected, found %d table(s)", len(out.TableNames))
}

func TestNewSecretsClient_ReadsSecret(t *testing.T) {
	cfg := integrationConfig(t)

	client, err := NewSecretsClient(context.Background(), cfg)
	if err != nil {
		t.Fatalf("NewSecretsClient() error = %v", err)
	}

	out, err := client.GetSecretValue(context.Background(), &secretsmanager.GetSecretValueInput{
		SecretId: aws.String(cfg.LinkedInCookiesSecret),
	})
	if err != nil {
		t.Fatalf("GetSecretValue() error = %v", err)
	}
	if out.SecretString == nil || *out.SecretString == "" {
		t.Fatal("secret string is empty")
	}
	t.Log("Secret read successfully")
}
