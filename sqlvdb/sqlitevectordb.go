package sqlvdb

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"sort"
	"sync"

	_ "modernc.org/sqlite"

	"github.com/ghmer/aicompanion/models"
)

// SQLiteVectorDb represents a vector database using SQLite.
type SQLiteVectorDb struct {
	db              *sql.DB
	mutex           sync.RWMutex
	schemas         map[string]bool
	dbPath          string
	normalizeVector bool
}

// NewSQLiteVectorDb creates a new SQLite vector database instance.
func NewSQLiteVectorDb(dbPath string, normalize bool) (*SQLiteVectorDb, error) {
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, err
	}
	s := &SQLiteVectorDb{
		db:              db,
		schemas:         make(map[string]bool),
		dbPath:          dbPath,
		normalizeVector: normalize,
	}

	ctx := context.Background()
	if err := s.loadSchemas(ctx); err != nil {
		return nil, err
	}

	return s, nil
}

// loadSchemas loads all existing schemas from the database.
func (s *SQLiteVectorDb) loadSchemas(ctx context.Context) error {
	query := `SELECT name FROM sqlite_master WHERE type='table'`
	rows, err := s.db.QueryContext(ctx, query)
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			return err
		}
		s.schemas[name] = true
	}
	return nil
}

// schemaExists checks if a schema with the given class name exists in the database.
func (s *SQLiteVectorDb) schemaExists(ctx context.Context, classname string) (bool, error) {
	query := `SELECT name FROM sqlite_master WHERE type='table' AND name=?`
	var name string
	err := s.db.QueryRowContext(ctx, query, classname).Scan(&name)
	if err == sql.ErrNoRows {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return true, nil
}

// GetSchema retrieves the schema for storing documents with the given class name.
func (s *SQLiteVectorDb) GetSchema(ctx context.Context, classname string) (interface{}, error) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	if exists, err := s.schemaExists(ctx, classname); err != nil {
		return nil, err
	} else if !exists {
		return nil, errors.New("schema does not exist")
	}
	return classname, nil
}

// GetSchemaClassNames retrieves the class names of all schemas in the database.
func (s *SQLiteVectorDb) GetSchemaClassNames(ctx context.Context) ([]string, error) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	var result []string

	query := `SELECT name FROM sqlite_master WHERE type='table'`
	rows, err := s.db.QueryContext(ctx, query)
	if err != nil {
		return result, err
	}
	defer rows.Close()

	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			return result, err
		}
		result = append(result, name)
	}
	return result, nil
}

// CreateSchema creates a new schema for storing documents with the given class name.
func (s *SQLiteVectorDb) CreateSchema(ctx context.Context, classname interface{}) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	classnameStr := classname.(string)
	if exists, err := s.schemaExists(ctx, classnameStr); err != nil {
		return err
	} else if exists {
		return errors.New("schema already exists")
	}

	query := fmt.Sprintf(`CREATE TABLE IF NOT EXISTS %s (
		id TEXT PRIMARY KEY,
		metadata BLOB,
		embeddings BLOB
	)`, classnameStr)
	if _, err := s.db.ExecContext(ctx, query); err != nil {
		return fmt.Errorf("failed to create schema: %w", err)
	}

	s.schemas[classnameStr] = true
	return nil
}

// DeleteSchema deletes a schema from the database.
func (s *SQLiteVectorDb) DeleteSchema(ctx context.Context, classname string) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	if _, exists := s.schemas[classname]; !exists {
		return errors.New("schema does not exist")
	}

	query := fmt.Sprintf(`DROP TABLE IF EXISTS %s`, classname)
	if _, err := s.db.ExecContext(ctx, query); err != nil {
		return fmt.Errorf("failed to delete schema: %w", err)
	}

	delete(s.schemas, classname)
	return nil
}

// AddDocument adds a document with the given class name and ID to the database.
func (s *SQLiteVectorDb) AddDocument(ctx context.Context, classname, id string, document models.Document) error {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	if _, exists := s.schemas[classname]; !exists {
		return errors.New("schema does not exist")
	}

	normalizedVector := s.NormalizeVector(document.Embeddings)
	vectorBytes, err := json.Marshal(normalizedVector)
	if err != nil {
		return fmt.Errorf("failed to serialize vector: %w", err)
	}

	metadataBytes, err := json.Marshal(document.Metadata)
	if err != nil {
		return fmt.Errorf("failed to serialize metadata: %w", err)
	}

	query := fmt.Sprintf(`INSERT OR REPLACE INTO %s (id, metadata, embeddings) VALUES (?, ?, ?)`, classname)
	if _, err := s.db.ExecContext(ctx, query, id, metadataBytes, vectorBytes); err != nil {
		return fmt.Errorf("failed to add document: %w", err)
	}

	return nil
}

// AddDocuments adds multiple documents to the database.
func (s *SQLiteVectorDb) AddDocuments(ctx context.Context, classname string, documents []models.Document) error {
	for _, doc := range documents {
		if err := s.AddDocument(ctx, classname, doc.ID, doc); err != nil {
			return err
		}
	}
	return nil
}

// UpdateDocument updates a document with the given class name and ID in the database.
func (s *SQLiteVectorDb) UpdateDocument(ctx context.Context, classname, id string, document models.Document) error {
	return s.AddDocument(ctx, classname, id, document)
}

// QueryDocuments queries documents based on a vector and returns the top result.
func (s *SQLiteVectorDb) QueryDocuments(ctx context.Context, classname string, vector []float32, limit int) ([]models.Document, error) {
	documents, err := s.QueryDocumentsWithFilter(ctx, classname, vector, limit, nil)
	if err != nil || len(documents) == 0 {
		return nil, err
	}
	return documents, nil
}

// QueryDocumentsWithFilter queries documents based on a vector and returns the top results with an optional filter.
func (s *SQLiteVectorDb) QueryDocumentsWithFilter(ctx context.Context, classname string, vector []float32, limit int, filter map[string]interface{}) ([]models.Document, error) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	if _, exists := s.schemas[classname]; !exists {
		return nil, errors.New("schema does not exist")
	}

	rows, err := s.db.QueryContext(ctx, fmt.Sprintf(`SELECT id, metadata, embeddings FROM %s`, classname))
	if err != nil {
		return nil, fmt.Errorf("failed to query documents: %w", err)
	}
	defer rows.Close()

	results := []struct {
		ID    string
		Score float64
		Data  models.Document
	}{}

	queryVector := s.NormalizeVector(vector)

	for rows.Next() {
		var id string
		var metadataJSON []byte
		var embeddingBytes []byte
		if err := rows.Scan(&id, &metadataJSON, &embeddingBytes); err != nil {
			return nil, fmt.Errorf("failed to scan row: %w", err)
		}

		var embeddings []float32
		if err := json.Unmarshal(embeddingBytes, &embeddings); err != nil {
			return nil, fmt.Errorf("failed to deserialize embeddings: %w", err)
		}

		var metadata map[string]interface{}
		if err := json.Unmarshal(metadataJSON, &metadata); err != nil {
			return nil, fmt.Errorf("failed to deserialize metadata: %w", err)
		}

		if filter == nil || matchesFilter(metadata, filter) {
			score := cosineSimilarity(queryVector, embeddings)
			results = append(results, struct {
				ID    string
				Score float64
				Data  models.Document
			}{ID: id, Score: score, Data: models.Document{
				ID:         id,
				ClassName:  classname,
				Embeddings: embeddings,
				Metadata:   metadata,
			}})
		}
	}

	sort.Slice(results, func(i, j int) bool {
		return results[i].Score > results[j].Score
	})

	output := []models.Document{}
	for i := 0; i < len(results) && i < limit; i++ {
		output = append(output, results[i].Data)
	}

	return output, nil
}

// DeleteDocument deletes a document from the database.
func (s *SQLiteVectorDb) DeleteDocument(ctx context.Context, classname, id string) error {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	if _, exists := s.schemas[classname]; !exists {
		return errors.New("schema does not exist")
	}

	query := fmt.Sprintf(`DELETE FROM %s WHERE id = ?`, classname)
	if _, err := s.db.ExecContext(ctx, query, id); err != nil {
		return fmt.Errorf("failed to delete document: %w", err)
	}

	return nil
}

// DeleteDocuments deletes multiple documents from the database.
func (s *SQLiteVectorDb) DeleteDocuments(ctx context.Context, classname string, ids []string) error {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	if _, exists := s.schemas[classname]; !exists {
		return errors.New("schema does not exist")
	}

	query := fmt.Sprintf(`DELETE FROM %s WHERE id = ?`, classname)
	for _, id := range ids {
		if _, err := s.db.ExecContext(ctx, query, id); err != nil {
			return fmt.Errorf("failed to delete document %s: %w", id, err)
		}
	}
	return nil
}

// NormalizeVector normalizes a vector if required.
func (s *SQLiteVectorDb) NormalizeVector(vector []float32) []float32 {
	if !s.normalizeVector {
		return vector
	}

	var magnitude float64
	for _, v := range vector {
		magnitude += float64(v * v)
	}
	if magnitude == 0 {
		return vector
	}
	magnitude = math.Sqrt(magnitude)
	for i := range vector {
		vector[i] /= float32(magnitude)
	}
	return vector
}

// matchesFilter checks if the metadata matches the filter.
func matchesFilter(metadata, filter map[string]interface{}) bool {
	for k, v := range filter {
		if metadata[k] != v {
			return false
		}
	}
	return true
}

// cosineSimilarity calculates the cosine similarity between two vectors.
func cosineSimilarity(v1, v2 []float32) float64 {
	var dotProduct, mag1, mag2 float64
	for i := range v1 {
		dotProduct += float64(v1[i] * v2[i])
		mag1 += float64(v1[i] * v1[i])
		mag2 += float64(v2[i] * v2[i])
	}
	if mag1 == 0 || mag2 == 0 {
		return 0
	}
	return dotProduct / (math.Sqrt(mag1) * math.Sqrt(mag2))
}
