package main

import (
	"context"
	"log"
	"os"

	"kamos/internal/openaiclient"
	"kamos/internal/pipeline"
	"kamos/internal/qdrantclient"
)

func main() {
	ctx := context.Background()

	qclient, err := qdrantclient.NewQdrantClient(ctx)
	if err != nil {
		log.Fatalf("Failed to initialize Qdrant: %v", err)
	}

	oclient := openaiclient.NewOpenAIClient(ctx, os.Getenv("OPENAI_API_KEY"))

	p := pipeline.NewPipeline(qclient, oclient, pipeline.Config{
		CollectionName:  "my_documents",
		VectorDimension: 3072,
	})

	if err := p.SetupCollection(); err != nil {
		log.Fatalf("Error setting up collection: %v", err)
	}

	documents := []pipeline.Document{
		{ID: 1, Text: "Go is a statically typed language"},
		{ID: 2, Text: "Retrieval-Augmented Generation enhances LLMs"},
		{ID: 3, Text: "Vector databases store high-dimensional embeddings"},
	}

	if err := p.ProcessDocuments(documents); err != nil {
		log.Fatalf("Error processing documents: %v", err)
	}

	p.Query("What is a vector database?")
}
