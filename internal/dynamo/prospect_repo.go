package dynamo

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"

	explorer "github.com/pavlomaksymov/in-network-explorer/internal"
)

var _ explorer.ProspectRepository = (*ProspectRepository)(nil)

// ProspectRepository persists Prospect aggregates in DynamoDB. Items are keyed
// by ProfileURL (PK). State-based queries use Scan with FilterExpression,
// which is acceptable for v1 volumes (< 1000 prospects).
type ProspectRepository struct {
	client *dynamodb.Client
	table  string
}

// NewProspectRepository creates a ProspectRepository.
func NewProspectRepository(client *dynamodb.Client, table string) *ProspectRepository {
	return &ProspectRepository{client: client, table: table}
}

// Save creates or fully replaces the prospect record.
func (r *ProspectRepository) Save(ctx context.Context, p *explorer.Prospect) error {
	item, err := marshalProspect(p)
	if err != nil {
		return fmt.Errorf("marshal prospect: %w", err)
	}
	_, err = r.client.PutItem(ctx, &dynamodb.PutItemInput{
		TableName: aws.String(r.table),
		Item:      item,
	})
	if err != nil {
		return fmt.Errorf("put prospect %q: %w", p.ProfileURL, err)
	}
	return nil
}

// Get returns the prospect for profileURL or explorer.ErrNotFound.
func (r *ProspectRepository) Get(ctx context.Context, profileURL string) (*explorer.Prospect, error) {
	out, err := r.client.GetItem(ctx, &dynamodb.GetItemInput{
		TableName: aws.String(r.table),
		Key:       prospectKey(profileURL),
	})
	if err != nil {
		return nil, fmt.Errorf("get prospect %q: %w", profileURL, err)
	}
	if out.Item == nil {
		return nil, explorer.ErrNotFound
	}
	return unmarshalProspect(out.Item)
}

// InsertIfNew writes the prospect only when no record exists for the URL.
func (r *ProspectRepository) InsertIfNew(ctx context.Context, p *explorer.Prospect) (bool, error) {
	item, err := marshalProspect(p)
	if err != nil {
		return false, fmt.Errorf("marshal prospect: %w", err)
	}
	_, err = r.client.PutItem(ctx, &dynamodb.PutItemInput{
		TableName:           aws.String(r.table),
		Item:                item,
		ConditionExpression: aws.String("attribute_not_exists(PK)"),
	})
	var condFailed *types.ConditionalCheckFailedException
	if err != nil {
		if isConditionFailed(err, &condFailed) {
			return false, nil
		}
		return false, fmt.Errorf("insert prospect %q: %w", p.ProfileURL, err)
	}
	return true, nil
}

// ListByState returns prospects in state whose NextActionAt is at or before dueBy.
func (r *ProspectRepository) ListByState(ctx context.Context, state explorer.State, dueBy time.Time) ([]*explorer.Prospect, error) {
	out, err := r.client.Scan(ctx, &dynamodb.ScanInput{
		TableName:        aws.String(r.table),
		FilterExpression: aws.String("#s = :state AND #na <= :dueBy"),
		ExpressionAttributeNames: map[string]string{
			"#s":  "State",
			"#na": "NextActionAt",
		},
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":state": &types.AttributeValueMemberS{Value: state.String()},
			":dueBy": &types.AttributeValueMemberS{Value: dueBy.UTC().Format(time.RFC3339)},
		},
	})
	if err != nil {
		return nil, fmt.Errorf("scan by state %s: %w", state, err)
	}
	return unmarshalProspects(out.Items)
}

// ListByStateOrderedByScore returns up to limit prospects in state, sorted by
// WorthinessScore descending. Uses Scan + client-side sort (adequate for v1).
func (r *ProspectRepository) ListByStateOrderedByScore(ctx context.Context, state explorer.State, limit int) ([]*explorer.Prospect, error) {
	out, err := r.client.Scan(ctx, &dynamodb.ScanInput{
		TableName:        aws.String(r.table),
		FilterExpression: aws.String("#s = :state"),
		ExpressionAttributeNames: map[string]string{
			"#s": "State",
		},
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":state": &types.AttributeValueMemberS{Value: state.String()},
		},
	})
	if err != nil {
		return nil, fmt.Errorf("scan by state %s: %w", state, err)
	}
	prospects, err := unmarshalProspects(out.Items)
	if err != nil {
		return nil, err
	}
	sort.Slice(prospects, func(i, j int) bool {
		return prospects[i].WorthinessScore > prospects[j].WorthinessScore
	})
	if limit > 0 && len(prospects) > limit {
		prospects = prospects[:limit]
	}
	return prospects, nil
}

// ── marshaling ──────────────────────────────────────────────────────────────

// prospectItem is the DynamoDB persistence representation of a Prospect.
type prospectItem struct {
	PK              string   `dynamodbav:"PK"`
	Slug            string   `dynamodbav:"Slug,omitempty"`
	Name            string   `dynamodbav:"Name,omitempty"`
	Headline        string   `dynamodbav:"Headline,omitempty"`
	Location        string   `dynamodbav:"Location,omitempty"`
	About           string   `dynamodbav:"About,omitempty"`
	RecentPosts     []string `dynamodbav:"RecentPosts,omitempty"`
	WorthinessScore int      `dynamodbav:"WorthinessScore"`
	ScoreReasoning  string   `dynamodbav:"ScoreReasoning,omitempty"`
	DraftedMessage  string   `dynamodbav:"DraftedMessage,omitempty"`
	CritiqueScore   int      `dynamodbav:"CritiqueScore"`
	EmbeddingID     string   `dynamodbav:"EmbeddingID,omitempty"`
	State           string   `dynamodbav:"State"`
	LastActionAt    string   `dynamodbav:"LastActionAt,omitempty"`
	NextActionAt    string   `dynamodbav:"NextActionAt,omitempty"`
	CreatedAt       string   `dynamodbav:"CreatedAt,omitempty"`
}

func marshalProspect(p *explorer.Prospect) (map[string]types.AttributeValue, error) {
	item := prospectItem{
		PK:              p.ProfileURL,
		Slug:            p.Slug,
		Name:            p.Name,
		Headline:        p.Headline,
		Location:        p.Location,
		About:           p.About,
		RecentPosts:     p.RecentPosts,
		WorthinessScore: p.WorthinessScore,
		ScoreReasoning:  p.ScoreReasoning,
		DraftedMessage:  p.DraftedMessage,
		CritiqueScore:   p.CritiqueScore,
		EmbeddingID:     p.EmbeddingID,
		State:           p.State.String(),
		LastActionAt:    formatTime(p.LastActionAt),
		NextActionAt:    formatTime(p.NextActionAt),
		CreatedAt:       formatTime(p.CreatedAt),
	}
	return attributevalue.MarshalMap(item)
}

func unmarshalProspect(av map[string]types.AttributeValue) (*explorer.Prospect, error) {
	var item prospectItem
	if err := attributevalue.UnmarshalMap(av, &item); err != nil {
		return nil, fmt.Errorf("unmarshal prospect: %w", err)
	}
	state, err := parseState(item.State)
	if err != nil {
		return nil, err
	}
	return &explorer.Prospect{
		ProfileURL:      item.PK,
		Slug:            item.Slug,
		Name:            item.Name,
		Headline:        item.Headline,
		Location:        item.Location,
		About:           item.About,
		RecentPosts:     item.RecentPosts,
		WorthinessScore: item.WorthinessScore,
		ScoreReasoning:  item.ScoreReasoning,
		DraftedMessage:  item.DraftedMessage,
		CritiqueScore:   item.CritiqueScore,
		EmbeddingID:     item.EmbeddingID,
		State:           state,
		LastActionAt:    parseTime(item.LastActionAt),
		NextActionAt:    parseTime(item.NextActionAt),
		CreatedAt:       parseTime(item.CreatedAt),
	}, nil
}

func unmarshalProspects(items []map[string]types.AttributeValue) ([]*explorer.Prospect, error) {
	out := make([]*explorer.Prospect, 0, len(items))
	for _, av := range items {
		p, err := unmarshalProspect(av)
		if err != nil {
			return nil, err
		}
		out = append(out, p)
	}
	return out, nil
}

func prospectKey(profileURL string) map[string]types.AttributeValue {
	return map[string]types.AttributeValue{
		"PK": &types.AttributeValueMemberS{Value: profileURL},
	}
}

func formatTime(t time.Time) string {
	if t.IsZero() {
		return ""
	}
	return t.UTC().Format(time.RFC3339)
}

func parseTime(s string) time.Time {
	if s == "" {
		return time.Time{}
	}
	t, _ := time.Parse(time.RFC3339, s)
	return t
}

// parseState converts a string like "SCANNED" to the corresponding State value.
func parseState(s string) (explorer.State, error) {
	switch s {
	case "SCANNED":
		return explorer.StateScanned, nil
	case "LIKED":
		return explorer.StateLiked, nil
	case "DRAFTED":
		return explorer.StateDrafted, nil
	case "SENT":
		return explorer.StateSent, nil
	case "ACCEPTED":
		return explorer.StateAccepted, nil
	case "REJECTED":
		return explorer.StateRejected, nil
	case "SKIPPED":
		return explorer.StateSkipped, nil
	default:
		return 0, fmt.Errorf("unknown state: %q", s)
	}
}

func isConditionFailed(err error, target **types.ConditionalCheckFailedException) bool {
	return err != nil && errors.As(err, target)
}
