package infrastructure

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/bedrockruntime"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/secretsmanager"
)

// loadAWSConfig loads the base AWS SDK configuration for the given region.
// When localEndpoint is non-empty, static dummy credentials are injected so
// that DynamoDB Local and LocalStack accept requests without real AWS creds.
func loadAWSConfig(ctx context.Context, region, localEndpoint string) (aws.Config, error) {
	opts := []func(*awsconfig.LoadOptions) error{
		awsconfig.WithRegion(region),
	}
	if localEndpoint != "" {
		opts = append(opts, awsconfig.WithCredentialsProvider(
			credentials.NewStaticCredentialsProvider("local", "local", ""),
		))
	}
	return awsconfig.LoadDefaultConfig(ctx, opts...)
}

// NewDynamoClient creates a DynamoDB client configured from cfg. When
// cfg.DynamoEndpoint is non-empty the client connects to that URL instead
// of the default AWS endpoint — intended for DynamoDB Local in development.
func NewDynamoClient(ctx context.Context, cfg Config) (*dynamodb.Client, error) {
	awsCfg, err := loadAWSConfig(ctx, cfg.AWSRegion, cfg.DynamoEndpoint)
	if err != nil {
		return nil, fmt.Errorf("load aws config for dynamodb: %w", err)
	}

	var opts []func(*dynamodb.Options)
	if cfg.DynamoEndpoint != "" {
		opts = append(opts, func(o *dynamodb.Options) {
			o.BaseEndpoint = aws.String(cfg.DynamoEndpoint)
		})
	}
	return dynamodb.NewFromConfig(awsCfg, opts...), nil
}

// NewBedrockClient creates a Bedrock Runtime client configured from cfg. It
// uses cfg.BedrockRegion for the region since model availability varies by
// region. No endpoint override is supported — Bedrock has no local equivalent.
func NewBedrockClient(ctx context.Context, cfg Config) (*bedrockruntime.Client, error) {
	awsCfg, err := loadAWSConfig(ctx, cfg.BedrockRegion, "")
	if err != nil {
		return nil, fmt.Errorf("load aws config for bedrock: %w", err)
	}
	return bedrockruntime.NewFromConfig(awsCfg), nil
}

// NewSecretsClient creates a Secrets Manager client configured from cfg. When
// cfg.SecretsEndpoint is non-empty the client connects to that URL instead
// of the default AWS endpoint — intended for LocalStack in development.
func NewSecretsClient(ctx context.Context, cfg Config) (*secretsmanager.Client, error) {
	awsCfg, err := loadAWSConfig(ctx, cfg.AWSRegion, cfg.SecretsEndpoint)
	if err != nil {
		return nil, fmt.Errorf("load aws config for secrets manager: %w", err)
	}

	var opts []func(*secretsmanager.Options)
	if cfg.SecretsEndpoint != "" {
		opts = append(opts, func(o *secretsmanager.Options) {
			o.BaseEndpoint = aws.String(cfg.SecretsEndpoint)
		})
	}
	return secretsmanager.NewFromConfig(awsCfg, opts...), nil
}
