package bedrock

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/bedrockruntime"
	"github.com/aws/aws-sdk-go-v2/service/bedrockruntime/types"

	explorer "github.com/pavlomaksymov/in-network-explorer/internal"
)

var _ explorer.LLMClient = (*Client)(nil)

// Client implements LLMClient using Amazon Bedrock's Converse API.
type Client struct {
	client  *bedrockruntime.Client
	modelID string
}

// NewClient creates a Bedrock LLM client for the given model.
func NewClient(client *bedrockruntime.Client, modelID string) *Client {
	return &Client{client: client, modelID: modelID}
}

// ScoreAndDraft scores the prospect against the target persona and drafts a
// personalized connection message. The response includes a self-critique score.
func (c *Client) ScoreAndDraft(ctx context.Context, p *explorer.Prospect, examples []explorer.Prospect) (explorer.ScoreResult, error) {
	prompt := buildScorePrompt(p, examples)
	raw, err := c.converse(ctx, prompt)
	if err != nil {
		return explorer.ScoreResult{}, fmt.Errorf("score and draft: %w", err)
	}
	return parseScoreResponse(raw)
}

// Critique evaluates a drafted connection message and returns a composite
// score from 3 (poor) to 15 (excellent).
func (c *Client) Critique(ctx context.Context, message string) (int, error) {
	prompt := buildCritiquePrompt(message)
	raw, err := c.converse(ctx, prompt)
	if err != nil {
		return 0, fmt.Errorf("critique: %w", err)
	}
	return parseCritiqueResponse(raw)
}

// converse sends a single-turn message to Bedrock and returns the text response.
func (c *Client) converse(ctx context.Context, prompt string) (string, error) {
	out, err := c.client.Converse(ctx, &bedrockruntime.ConverseInput{
		ModelId: aws.String(c.modelID),
		Messages: []types.Message{
			{
				Role: types.ConversationRoleUser,
				Content: []types.ContentBlock{
					&types.ContentBlockMemberText{Value: prompt},
				},
			},
		},
	})
	if err != nil {
		return "", fmt.Errorf("bedrock converse: %w", err)
	}
	return extractText(out)
}

func extractText(out *bedrockruntime.ConverseOutput) (string, error) {
	msg, ok := out.Output.(*types.ConverseOutputMemberMessage)
	if !ok {
		return "", fmt.Errorf("unexpected output type: %T", out.Output)
	}
	for _, block := range msg.Value.Content {
		if tb, ok := block.(*types.ContentBlockMemberText); ok {
			return tb.Value, nil
		}
	}
	return "", fmt.Errorf("no text block in response")
}
