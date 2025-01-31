// openai.go

package openaiclient

import (
	"context"

	"github.com/sashabaranov/go-openai"
)

type OpenAIClient struct {
	ctx    context.Context
	client *openai.Client
}

type Query struct {
	Prompt    string
	PrePrompt string
}

func NewOpenAIClient(ctx context.Context, apiKey string) *OpenAIClient {
	return &OpenAIClient{
		ctx:    ctx,
		client: openai.NewClient(apiKey),
	}
}

func (o *OpenAIClient) SendQuery(q *Query) (string, error) {
	resp, err := o.client.CreateChatCompletion(
		context.Background(),
		openai.ChatCompletionRequest{
			Model: openai.GPT4oMini,
			Messages: []openai.ChatCompletionMessage{
				{
					Role:    openai.ChatMessageRoleUser,
					Content: q.Prompt,
				},
				{
					Role:    openai.ChatMessageRoleAssistant,
					Content: q.PrePrompt,
				},
			},
		},
	)

	if err != nil {
		return "", err
	}

	return resp.Choices[0].Message.Content, nil
}

func (o *OpenAIClient) GenerateEmbedding(text string, dimension int) ([]float32, error) {
	resp, err := o.client.CreateEmbeddings(o.ctx, openai.EmbeddingRequest{
		Input: text,
		Model: openai.LargeEmbedding3,
	})
	if err != nil {
		return nil, err
	}

	var embedding []float32
	for _, emb := range resp.Data {
		embedding = append(embedding, emb.Embedding...)
	}

	return embedding, nil
}
