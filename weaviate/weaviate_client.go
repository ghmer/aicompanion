package weaviate

import (
	"context"
	"fmt"
	"log"

	"github.com/weaviate/weaviate-go-client/v4/weaviate"
	"github.com/weaviate/weaviate-go-client/v4/weaviate/auth"
	"github.com/weaviate/weaviate-go-client/v4/weaviate/graphql"
	"github.com/weaviate/weaviate/entities/models"
)

// WeaviateClient implements the VectorDB interface for Weaviate.
type WeaviateClient struct {
	client *weaviate.Client
}

// NewWeaviateClient initializes a WeaviateClient with the given endpoint and API key.
func NewWeaviateClient(endpoint, apiKey string) (*WeaviateClient, error) {
	cfg := weaviate.Config{
		Scheme: "https",
		Host:   endpoint,
		AuthConfig: auth.ClientCredentials{
			ClientSecret: apiKey,
			Scopes:       []string{"email"},
		},
	}
	client, err := weaviate.NewClient(cfg)
	if err != nil {
		log.Fatal(err)
		return &WeaviateClient{}, err
	}

	return &WeaviateClient{client: client}, nil
}

// AddDocument adds a document with metadata and embeddings to the vector database.
func (wc *WeaviateClient) AddDocument(ctx context.Context, classname, id string, data map[string]interface{}, vector []float32) error {
	_, err := wc.client.Data().Creator().WithClassName(classname).WithID(id).WithProperties(data).WithVector(vector).Do(ctx)
	return err
}

// UpdateDocument updates an existing document's metadata or vector in the vector database.
func (wc *WeaviateClient) UpdateDocument(ctx context.Context, id string, classname string, data map[string]interface{}, vector []float32) error {
	err := wc.client.Data().Updater().WithClassName(classname).WithID(id).WithProperties(data).WithVector(vector).Do(ctx)
	return err
}

// QueryDocuments retrieves documents most relevant to the provided query vector.
func (wc *WeaviateClient) QueryDocuments(ctx context.Context, className string, vector []float32, limit int) ([]map[string]interface{}, error) {
	title := graphql.Field{Name: "title"}
	content := graphql.Field{Name: "content"}
	source := graphql.Field{Name: "source"}
	_additional := graphql.Field{
		Name: "_additional", Fields: []graphql.Field{
			{Name: "certainty"}, // only supported if distance==cosine
			{Name: "distance"},  // always supported
		},
	}

	nearVector := wc.client.GraphQL().NearVectorArgBuilder().
		WithVector(vector) // Replace with a compatible vector

	result, err := wc.client.GraphQL().Get().
		WithClassName(className).
		WithLimit(limit).
		WithFields(title, content, source, _additional).
		WithNearVector(nearVector).
		Do(ctx)

	if err != nil {
		return nil, err
	}

	for _, err := range result.Errors {
		fmt.Println("error", err.Message)
	}

	var documents []map[string]interface{}
	for _, data := range result.Data["Get"].(map[string]interface{})[className].([]interface{}) {
		documents = append(documents, data.(map[string]interface{}))

	}

	fmt.Println("documents", documents)
	return documents, nil
}

// DeleteDocument removes a document from the vector database by its ID.
func (wc *WeaviateClient) DeleteDocument(ctx context.Context, className, id string) error {
	err := wc.client.Data().Deleter().WithClassName(className).WithID(id).Do(ctx)
	return err
}

// CreateSchema sets up the schema for storing documents in Weaviate.
func (wc *WeaviateClient) CreateSchema(ctx context.Context, class *models.Class) error {
	return wc.client.Schema().ClassCreator().WithClass(class).Do(ctx)
}

// CreateSchema sets up the schema for storing documents in Weaviate.
func (wc *WeaviateClient) GetSchema(ctx context.Context, className string) (*models.Class, error) {
	return wc.client.Schema().ClassGetter().WithClassName(className).Do(ctx)
}

// DeleteSchema removes a schema class from Weaviate.
func (wc *WeaviateClient) DeleteSchema(ctx context.Context, className string) error {
	return wc.client.Schema().ClassDeleter().WithClassName(className).Do(ctx)
}
