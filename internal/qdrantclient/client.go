package qdrantclient

import (
	"context"
	"fmt"
	"kamos/internal/pipeline"
	"log"
	"strings"

	"github.com/qdrant/go-client/qdrant"
)

// QdrantClient wraps the gRPC client for reusability
type QdrantClient struct {
	client *qdrant.Client
	ctx    context.Context
}

func NewQdrantClient(ctx context.Context) (*QdrantClient, error) {
	client, err := qdrant.NewClient(&qdrant.Config{
		Host:   "localhost",
		Port:   6334,
		UseTLS: false,
	})
	if err != nil {
		return nil, err
	}

	return &QdrantClient{
		client: client,
		ctx:    ctx,
	}, nil
}

// initializeQdrant initializes the Qdrant client and collection
func InitializeQdrant(ctx context.Context, config pipeline.Config) (*QdrantClient, error) {
	qclient, err := NewQdrantClient(ctx)
	if err != nil {
		return nil, err
	}

	err = qclient.CreateCollection(config.CollectionName, config.VectorDimension)
	if err != nil && !strings.Contains(err.Error(), "AlreadyExists") {
		return nil, err
	} else if err == nil {
		log.Printf("Collection %s created successfully", config.CollectionName)
	} else {
		log.Printf("Collection %s already exists, skipping creation", config.CollectionName)
	}

	return qclient, nil
}

// CreateCollection ensures a Qdrant collection exists
func (q *QdrantClient) CreateCollection(collectionName string, dimension int) error {
	err := q.client.CreateCollection(q.ctx, &qdrant.CreateCollection{
		CollectionName: collectionName,
		VectorsConfig: &qdrant.VectorsConfig{
			Config: &qdrant.VectorsConfig_Params{
				Params: &qdrant.VectorParams{
					Size:     uint64(dimension),
					Distance: qdrant.Distance_Cosine,
				},
			},
		},
	})

	if err != nil {
		if strings.Contains(err.Error(), "AlreadyExists") {
			log.Printf("Collection %s already exists, skipping creation", collectionName)
			return nil
		}
		log.Printf("Collection creation error (might already exist): %v", err)
	}
	log.Printf("Collection created successfully")
	return err
}

// AddDocument inserts a document with an embedding
func (q *QdrantClient) AddDocument(collectionName string, id uint64, text string, embedding []float32) error {
	_, err := q.client.Upsert(q.ctx, &qdrant.UpsertPoints{
		CollectionName: collectionName,
		Points: []*qdrant.PointStruct{
			{
				Id: qdrant.NewIDNum(id),
				Vectors: &qdrant.Vectors{
					VectorsOptions: &qdrant.Vectors_Vector{
						Vector: &qdrant.Vector{
							Data: embedding,
						},
					},
				},
				Payload: map[string]*qdrant.Value{
					"text": {Kind: &qdrant.Value_StringValue{StringValue: text}},
				},
			},
		},
	})

	if err != nil {
		return fmt.Errorf("failed to insert document %d: %v", id, err)
	}

	fmt.Printf("Inserted document: %s\n", text)
	return nil
}

func (q *QdrantClient) Exists(collectionName string, id uint64) (bool, error) {
	point, err := q.client.Get(q.ctx, &qdrant.GetPoints{
		CollectionName: collectionName,
		Ids: []*qdrant.PointId{
			qdrant.NewIDNum(id),
		},
	})
	if err != nil {
		return false, err
	}
	if point != nil {
		return true, nil
	}

	return false, nil
}

// Search retrieves the top-k most relevant documents based on a query embedding
func (q *QdrantClient) Search(collectionName string, embedding []float32) (string, error) {
	points, err := q.client.Query(q.ctx, &qdrant.QueryPoints{
		CollectionName: collectionName,
		Query:          qdrant.NewQuery(embedding...),
	})
	if err != nil {
		return "", fmt.Errorf("failed to search collection %s: %v", collectionName, err)
	}

	var result string
	for _, point := range points {
		if textValue, ok := point.Payload["text"].GetKind().(*qdrant.Value_StringValue); ok {
			result += textValue.StringValue + "\n"
		}
	}

	log.Printf("Retrieved %d relevant documents from collection %s", len(result), collectionName)
	return result, nil
}
