package config

import (
	"context"
	"testing"
)

func TestLoadAWSConfig_UsesRegion(t *testing.T) {
	cfg, err := loadAWSConfig(context.Background(), "ap-southeast-1", "")
	if err != nil {
		t.Fatalf("loadAWSConfig() error = %v", err)
	}
	if cfg.Region != "ap-southeast-1" {
		t.Errorf("Region = %q, want %q", cfg.Region, "ap-southeast-1")
	}
}

func TestLoadAWSConfig_InjectsStaticCreds_WhenEndpointSet(t *testing.T) {
	cfg, err := loadAWSConfig(context.Background(), "eu-central-1", "http://localhost:8000")
	if err != nil {
		t.Fatalf("loadAWSConfig() error = %v", err)
	}

	creds, err := cfg.Credentials.Retrieve(context.Background())
	if err != nil {
		t.Fatalf("Retrieve() error = %v", err)
	}
	if creds.AccessKeyID != "local" {
		t.Errorf("AccessKeyID = %q, want %q", creds.AccessKeyID, "local")
	}
	if creds.SecretAccessKey != "local" {
		t.Errorf("SecretAccessKey = %q, want %q", creds.SecretAccessKey, "local")
	}
}

func TestLoadAWSConfig_DefaultCredentials_WhenNoEndpoint(t *testing.T) {
	cfg, err := loadAWSConfig(context.Background(), "eu-central-1", "")
	if err != nil {
		t.Fatalf("loadAWSConfig() error = %v", err)
	}
	// When no endpoint override is set, the default credential chain is used.
	// We cannot retrieve credentials without real AWS config, but we verify
	// the config was built without error and region is correct.
	if cfg.Region != "eu-central-1" {
		t.Errorf("Region = %q, want %q", cfg.Region, "eu-central-1")
	}
}

func TestNewDynamoClient_ReturnsClient(t *testing.T) {
	cfg := Config{
		AWSRegion:      "eu-central-1",
		DynamoEndpoint: "http://localhost:8000",
	}
	client, err := NewDynamoClient(context.Background(), cfg)
	if err != nil {
		t.Fatalf("NewDynamoClient() error = %v", err)
	}
	if client == nil {
		t.Fatal("NewDynamoClient() returned nil client")
	}
}

func TestNewBedrockClient_ReturnsClient(t *testing.T) {
	cfg := Config{
		AWSRegion:     "eu-central-1",
		BedrockRegion: "us-east-1",
	}
	client, err := NewBedrockClient(context.Background(), cfg)
	if err != nil {
		t.Fatalf("NewBedrockClient() error = %v", err)
	}
	if client == nil {
		t.Fatal("NewBedrockClient() returned nil client")
	}
}

func TestNewSecretsClient_ReturnsClient(t *testing.T) {
	cfg := Config{
		AWSRegion:       "eu-central-1",
		SecretsEndpoint: "http://localhost:4566",
	}
	client, err := NewSecretsClient(context.Background(), cfg)
	if err != nil {
		t.Fatalf("NewSecretsClient() error = %v", err)
	}
	if client == nil {
		t.Fatal("NewSecretsClient() returned nil client")
	}
}
