package bedrock

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/bedrockruntime"

	explorer "github.com/pavlomaksymov/in-network-explorer/internal"
)

var _ explorer.EmbeddingClient = (*EmbeddingClient)(nil)

// EmbeddingClient converts text to vectors using Amazon Titan Embeddings
// via the Bedrock InvokeModel API.
type EmbeddingClient struct {
	client  *bedrockruntime.Client
	modelID string
}

// NewEmbeddingClient creates an EmbeddingClient. modelID should be a Titan
// Embeddings model (e.g. "amazon.titan-embed-text-v2:0").
func NewEmbeddingClient(client *bedrockruntime.Client, modelID string) *EmbeddingClient {
	return &EmbeddingClient{client: client, modelID: modelID}
}

// Embed returns the embedding vector for the given text.
func (c *EmbeddingClient) Embed(ctx context.Context, text string) ([]float32, error) {
	payload, err := json.Marshal(titanEmbedRequest{InputText: text})
	if err != nil {
		return nil, fmt.Errorf("marshal embed request: %w", err)
	}

	out, err := c.client.InvokeModel(ctx, &bedrockruntime.InvokeModelInput{
		ModelId:     aws.String(c.modelID),
		ContentType: aws.String("application/json"),
		Body:        payload,
	})
	if err != nil {
		return nil, fmt.Errorf("invoke embedding model: %w", err)
	}

	var resp titanEmbedResponse
	if err := json.Unmarshal(out.Body, &resp); err != nil {
		return nil, fmt.Errorf("unmarshal embed response: %w", err)
	}
	return resp.Embedding, nil
}

type titanEmbedRequest struct {
	InputText string `json:"inputText"`
}

type titanEmbedResponse struct {
	Embedding []float32 `json:"embedding"`
}
