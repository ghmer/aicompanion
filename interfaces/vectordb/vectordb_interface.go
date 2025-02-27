package vectordb

import (
	"context"

	"github.com/ghmer/aicompanion/models"
)

type QueryOptions struct {
	Limit               int            `json:"limit"`
	Filter              map[string]any `json:"filter"`
	SimilarityThreshold float64        `json:"similarity_threshold"`
}

// AICompanion defines the interface for interacting with AI models.
type VectorDb interface {
	AddDocument(ctx context.Context, classname, id string, document models.Document) error
	AddDocuments(ctx context.Context, classname string, documents []models.Document) error
	UpdateDocument(ctx context.Context, classname, id string, document models.Document) error
	UpdateDocuments(ctx context.Context, classname string, documents []models.Document) error
	QueryDocuments(ctx context.Context, classname string, vector []float32, queryOptions QueryOptions) ([]models.Document, error)
	DeleteDocument(ctx context.Context, classname, id string) error
	DeleteDocuments(ctx context.Context, classname string, ids []string) error
	CreateSchema(ctx context.Context, classname any) error
	GetSchema(ctx context.Context, classname string) (any, error)
	GetSchemas(ctx context.Context) ([]string, error)
	DeleteSchema(ctx context.Context, classname string) error
	DeleteSchemas(ctx context.Context, classnames []string) error
}
