package rag

import (
	"context"

	"github.com/ghmer/aicompanion/models"
	"github.com/ghmer/aicompanion/sqlvdb"
	"github.com/ghmer/aicompanion/weaviate"
)

// AICompanion defines the interface for interacting with AI models.
type VectorDbClient interface {
	AddDocument(ctx context.Context, classname, id string, document models.Document) error
	UpdateDocument(ctx context.Context, classname, id string, document models.Document) error
	QueryDocuments(ctx context.Context, classname string, vector []float32, limit int) ([]models.Document, error)
	DeleteDocument(ctx context.Context, classname, id string) error
	CreateSchema(ctx context.Context, classname interface{}) error
	GetSchema(ctx context.Context, classname string) (interface{}, error)
	DeleteSchema(ctx context.Context, classname string) error
	AddDocuments(ctx context.Context, classname string, documents []models.Document) error
	DeleteDocuments(ctx context.Context, classname string, ids []string) error
	QueryDocumentsWithFilter(ctx context.Context, classname string, vector []float32, limit int, filter map[string]interface{}) ([]models.Document, error)
}

func NewSQLiteVectorDb(dbpath string, normalize bool) (VectorDbClient, error) {
	return sqlvdb.NewSQLiteVectorDb(dbpath, normalize)
}

func NewWeaviateClient(endpoint, apiKey string) (VectorDbClient, error) {
	return weaviate.NewWeaviateClient(endpoint, apiKey)
}
