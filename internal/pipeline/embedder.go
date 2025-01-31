package pipeline

import "kamos/internal/openaiclient"

type Embedder interface {
	GenerateEmbedding(text string, dimension int) ([]float32, error)
	SendQuery(query *openaiclient.Query) (string, error)
}
