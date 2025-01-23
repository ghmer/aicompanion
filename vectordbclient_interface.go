package aicompanion

import (
	"context"

	"github.com/ghmer/aicompanion/weaviate"
	"github.com/weaviate/weaviate/entities/models"
)

// AICompanion defines the interface for interacting with AI models.
type VectorDbClient interface {
	AddDocument(ctx context.Context, classname, id string, data map[string]interface{}, vector []float32) error

	// UpdateDocument updates an existing document's metadata or vector in the vector database.
	UpdateDocument(ctx context.Context, id string, classname string, data map[string]interface{}, vector []float32) error

	// QueryDocuments retrieves documents most relevant to the provided query vector.
	QueryDocuments(ctx context.Context, className string, vector []float32, limit int) ([]map[string]interface{}, error)

	// DeleteDocument removes a document from the vector database by its ID.
	DeleteDocument(ctx context.Context, className, id string) error

	// CreateSchema sets up the schema for storing documents in Weaviate.
	CreateSchema(ctx context.Context, class *models.Class) error

	// CreateSchema sets up the schema for storing documents in Weaviate.
	GetSchema(ctx context.Context, className string) (*models.Class, error)

	// DeleteSchema removes a schema class from Weaviate.
	DeleteSchema(ctx context.Context, className string) error
}

func NewVectorDbClient(endpoint, apikey string) (VectorDbClient, error) {
	return weaviate.NewWeaviateClient(endpoint, apikey)
}
