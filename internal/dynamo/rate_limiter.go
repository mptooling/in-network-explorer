// Package dynamo provides DynamoDB-backed implementations of the repository
// and rate-limiting interfaces defined in the explorer domain package.
package dynamo

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"

	explorer "github.com/pavlomaksymov/in-network-explorer/internal"
)

var _ explorer.RateLimiter = (*RateLimiter)(nil)

// RateLimiter enforces daily action caps backed by DynamoDB atomic counters.
// Each scope (e.g. "profile_views") gets a separate item per UTC day, keyed as
// RATE#<scope>#<YYYY-MM-DD>.
type RateLimiter struct {
	client *dynamodb.Client
	table  string
	caps   map[string]int
}

// NewRateLimiter creates a RateLimiter. caps maps scope names to their daily
// maximum (e.g. {"profile_views": 40, "connection_requests": 10}).
func NewRateLimiter(client *dynamodb.Client, table string, caps map[string]int) *RateLimiter {
	return &RateLimiter{client: client, table: table, caps: caps}
}

// Acquire atomically increments the counter for scope. If the daily cap has
// been reached it returns explorer.ErrRateLimitExceeded.
func (rl *RateLimiter) Acquire(ctx context.Context, scope string) error {
	max, ok := rl.caps[scope]
	if !ok {
		return fmt.Errorf("unknown rate limit scope: %q", scope)
	}

	_, err := rl.client.UpdateItem(ctx, &dynamodb.UpdateItemInput{
		TableName: aws.String(rl.table),
		Key: map[string]types.AttributeValue{
			"PK": &types.AttributeValueMemberS{Value: rateLimitPK(scope)},
		},
		UpdateExpression:    aws.String("SET #c = if_not_exists(#c, :zero) + :one, #ttl = :ttl"),
		ConditionExpression: aws.String("attribute_not_exists(#c) OR #c < :max"),
		ExpressionAttributeNames: map[string]string{
			"#c":   "Count",
			"#ttl": "ExpiresAt",
		},
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":zero": &types.AttributeValueMemberN{Value: "0"},
			":one":  &types.AttributeValueMemberN{Value: "1"},
			":max":  &types.AttributeValueMemberN{Value: strconv.Itoa(max)},
			":ttl":  &types.AttributeValueMemberN{Value: strconv.FormatInt(endOfDayTTL(), 10)},
		},
	})

	var condFailed *types.ConditionalCheckFailedException
	if errors.As(err, &condFailed) {
		return explorer.ErrRateLimitExceeded
	}
	if err != nil {
		return fmt.Errorf("rate limiter acquire %q: %w", scope, err)
	}
	return nil
}

// Current returns how many actions have been taken today for scope.
func (rl *RateLimiter) Current(ctx context.Context, scope string) (int, error) {
	if _, ok := rl.caps[scope]; !ok {
		return 0, fmt.Errorf("unknown rate limit scope: %q", scope)
	}

	out, err := rl.client.GetItem(ctx, &dynamodb.GetItemInput{
		TableName: aws.String(rl.table),
		Key: map[string]types.AttributeValue{
			"PK": &types.AttributeValueMemberS{Value: rateLimitPK(scope)},
		},
		ProjectionExpression:     aws.String("#c"),
		ExpressionAttributeNames: map[string]string{"#c": "Count"},
	})
	if err != nil {
		return 0, fmt.Errorf("rate limiter current %q: %w", scope, err)
	}
	if out.Item == nil {
		return 0, nil
	}

	n, ok := out.Item["Count"].(*types.AttributeValueMemberN)
	if !ok {
		return 0, nil
	}
	count, err := strconv.Atoi(n.Value)
	if err != nil {
		return 0, fmt.Errorf("rate limiter parse count for %q: %w", scope, err)
	}
	return count, nil
}

// rateLimitPK builds the partition key for today's counter in UTC.
func rateLimitPK(scope string) string {
	return "RATE#" + scope + "#" + time.Now().UTC().Format("2006-01-02")
}

// endOfDayTTL returns a Unix epoch 48 hours after midnight tonight (UTC),
// giving DynamoDB TTL a comfortable cleanup window.
func endOfDayTTL() int64 {
	now := time.Now().UTC()
	midnight := time.Date(now.Year(), now.Month(), now.Day()+1, 0, 0, 0, 0, time.UTC)
	return midnight.Add(48 * time.Hour).Unix()
}
