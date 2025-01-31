package pipeline

import (
	"kamos/internal/openaiclient"
	"log"
)

type Pipeline struct {
	Store    DocumentStore
	Embedder Embedder
	Config   Config
}

type Config struct {
	CollectionName  string
	VectorDimension int
}

type Document struct {
	ID   uint64
	Text string
}

func NewPipeline(store DocumentStore, embedder Embedder, config Config) *Pipeline {
	return &Pipeline{Store: store, Embedder: embedder, Config: config}
}

func (p *Pipeline) SetupCollection() error {
	err := p.Store.CreateCollection(p.Config.CollectionName, p.Config.VectorDimension)
	if err != nil {
		return err
	}
	return nil
}

func (p *Pipeline) ProcessDocuments(documents []Document) error {
	for _, doc := range documents {
		exists, err := p.Store.Exists(p.Config.CollectionName, doc.ID)
		if err != nil {
			return err
		}
		if exists {
			log.Printf("Document %d already exists, skipping", doc.ID)
			continue
		}

		embedding, err := p.Embedder.GenerateEmbedding(doc.Text, p.Config.VectorDimension)
		if err != nil {
			return err
		}

		if err := p.Store.AddDocument(p.Config.CollectionName, doc.ID, doc.Text, embedding); err != nil {
			return err
		}

		log.Printf("Inserted document %d", doc.ID)
	}
	return nil
}

func (p *Pipeline) Query(prompt string) {
	prePrompt := "You are an expert AWS DevOps assistant specialized in Amazon EKS. Your role is to provide clear, concise, and accurate answers using the retrieved knowledge base. If the retrieved information is insufficient, clarify that instead of hallucinating. Cite relevant sources when applicable."

	promptEmbed, err := p.Embedder.GenerateEmbedding(prompt, p.Config.VectorDimension)
	if err != nil {
		log.Fatalf("Failed to generate embedding for query: %v", err)
	}

	relatedVectors, err := p.Store.Search(p.Config.CollectionName, promptEmbed)
	if err != nil {
		log.Fatalf("Failed to retrieve related documents: %v", err)
	}

	query := &openaiclient.Query{
		Prompt:    relatedVectors + prompt,
		PrePrompt: prePrompt,
	}

	result, err := p.Embedder.SendQuery(query)
	if err != nil {
		log.Fatalf("Failed to get AI response: %v", err)
	}

	log.Println(result)
}
