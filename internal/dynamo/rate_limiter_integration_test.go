//go:build integration

package dynamo

import (
	"context"
	"errors"
	"os"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"

	explorer "github.com/pavlomaksymov/in-network-explorer/internal"
)

func integrationClient(t *testing.T) (*dynamodb.Client, string) {
	t.Helper()

	endpoint := os.Getenv("DYNAMO_ENDPOINT")
	if endpoint == "" {
		t.Skip("DYNAMO_ENDPOINT not set, skipping integration test")
	}
	table := os.Getenv("DYNAMO_TABLE")
	if table == "" {
		t.Skip("DYNAMO_TABLE not set, skipping integration test")
	}

	cfg, err := awsconfig.LoadDefaultConfig(context.Background(),
		awsconfig.WithRegion("eu-central-1"),
		awsconfig.WithCredentialsProvider(
			credentials.NewStaticCredentialsProvider("local", "local", ""),
		),
	)
	if err != nil {
		t.Fatalf("load aws config: %v", err)
	}

	client := dynamodb.NewFromConfig(cfg, func(o *dynamodb.Options) {
		o.BaseEndpoint = aws.String(endpoint)
	})
	return client, table
}

func TestIntegration_Acquire_IncrementsCounter(t *testing.T) {
	client, table := integrationClient(t)
	scope := "test_acquire_" + t.Name()
	rl := NewRateLimiter(client, table, map[string]int{scope: 5})
	ctx := context.Background()

	if err := rl.Acquire(ctx, scope); err != nil {
		t.Fatalf("first Acquire() error = %v", err)
	}

	got, err := rl.Current(ctx, scope)
	if err != nil {
		t.Fatalf("Current() error = %v", err)
	}
	if got != 1 {
		t.Fatalf("Current() = %d, want 1", got)
	}
}

func TestIntegration_Acquire_RespectsCapAndReturnsCurrent(t *testing.T) {
	client, table := integrationClient(t)
	scope := "test_cap_" + t.Name()
	cap := 3
	rl := NewRateLimiter(client, table, map[string]int{scope: cap})
	ctx := context.Background()

	for i := range cap {
		if err := rl.Acquire(ctx, scope); err != nil {
			t.Fatalf("Acquire() #%d error = %v", i+1, err)
		}
	}

	got, err := rl.Current(ctx, scope)
	if err != nil {
		t.Fatalf("Current() error = %v", err)
	}
	if got != cap {
		t.Fatalf("Current() = %d, want %d", got, cap)
	}

	err = rl.Acquire(ctx, scope)
	if !errors.Is(err, explorer.ErrRateLimitExceeded) {
		t.Fatalf("Acquire() over cap: got %v, want ErrRateLimitExceeded", err)
	}
}

func TestIntegration_Acquire_ScopesAreIndependent(t *testing.T) {
	client, table := integrationClient(t)
	scopeA := "test_scopeA_" + t.Name()
	scopeB := "test_scopeB_" + t.Name()
	rl := NewRateLimiter(client, table, map[string]int{scopeA: 1, scopeB: 1})
	ctx := context.Background()

	if err := rl.Acquire(ctx, scopeA); err != nil {
		t.Fatalf("Acquire(A) error = %v", err)
	}

	// Scope B should still be available.
	if err := rl.Acquire(ctx, scopeB); err != nil {
		t.Fatalf("Acquire(B) error = %v, want nil", err)
	}
}

func TestIntegration_Current_ReturnsZeroForNewScope(t *testing.T) {
	client, table := integrationClient(t)
	scope := "test_zero_" + t.Name()
	rl := NewRateLimiter(client, table, map[string]int{scope: 10})
	ctx := context.Background()

	got, err := rl.Current(ctx, scope)
	if err != nil {
		t.Fatalf("Current() error = %v", err)
	}
	if got != 0 {
		t.Fatalf("Current() = %d, want 0", got)
	}
}
