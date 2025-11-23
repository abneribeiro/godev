package storage

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/google/uuid"
)

// Collection represents a folder/group of saved requests
type Collection struct {
	ID          string          `json:"id"`
	Name        string          `json:"name"`
	Description string          `json:"description"`
	CreatedAt   time.Time       `json:"created_at"`
	UpdatedAt   time.Time       `json:"updated_at"`
	Requests    []SavedRequest  `json:"requests"`
	SubCollections []Collection `json:"sub_collections,omitempty"`
}

// CollectionConfig holds all collections
type CollectionConfig struct {
	Version     string       `json:"version"`
	Collections []Collection `json:"collections"`
}

const collectionsFile = "collections.json"

// LoadCollections loads all collections from disk
func (s *Storage) LoadCollections() (*CollectionConfig, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get home directory: %w", err)
	}

	configDirPath := filepath.Join(homeDir, configDir)
	collectionsPath := filepath.Join(configDirPath, collectionsFile)

	// If file doesn't exist, return empty config
	if _, err := os.Stat(collectionsPath); os.IsNotExist(err) {
		return &CollectionConfig{
			Version:     version,
			Collections: []Collection{},
		}, nil
	}

	data, err := os.ReadFile(collectionsPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read collections file: %w", err)
	}

	var config CollectionConfig
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse collections file: %w", err)
	}

	return &config, nil
}

// SaveCollections saves all collections to disk
func (s *Storage) SaveCollections(config *CollectionConfig) error {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("failed to get home directory: %w", err)
	}

	configDirPath := filepath.Join(homeDir, configDir)
	collectionsPath := filepath.Join(configDirPath, collectionsFile)

	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal collections: %w", err)
	}

	// Use secure file permissions (0600 - only owner can read/write)
	if err := os.WriteFile(collectionsPath, data, 0600); err != nil {
		return fmt.Errorf("failed to write collections file: %w", err)
	}

	return nil
}

// CreateCollection creates a new collection
func CreateCollection(name, description string) Collection {
	now := time.Now()
	return Collection{
		ID:             uuid.New().String(),
		Name:           name,
		Description:    description,
		CreatedAt:      now,
		UpdatedAt:      now,
		Requests:       []SavedRequest{},
		SubCollections: []Collection{},
	}
}

// AddRequestToCollection adds a request to a collection
func AddRequestToCollection(collection *Collection, request SavedRequest) {
	collection.Requests = append(collection.Requests, request)
	collection.UpdatedAt = time.Now()
}

// RemoveRequestFromCollection removes a request from a collection
func RemoveRequestFromCollection(collection *Collection, requestID string) error {
	for i, req := range collection.Requests {
		if req.ID == requestID {
			collection.Requests = append(collection.Requests[:i], collection.Requests[i+1:]...)
			collection.UpdatedAt = time.Now()
			return nil
		}
	}
	return fmt.Errorf("request not found in collection: %s", requestID)
}

// AddSubCollection adds a sub-collection to a collection
func AddSubCollection(parent *Collection, child Collection) {
	parent.SubCollections = append(parent.SubCollections, child)
	parent.UpdatedAt = time.Now()
}

// FindCollectionByID recursively finds a collection by ID
func FindCollectionByID(collections []Collection, id string) *Collection {
	for i := range collections {
		if collections[i].ID == id {
			return &collections[i]
		}
		if found := FindCollectionByID(collections[i].SubCollections, id); found != nil {
			return found
		}
	}
	return nil
}

// ImportPostmanCollection imports a Postman collection format
type PostmanRequest struct {
	Name    string                 `json:"name"`
	Request PostmanRequestDetails `json:"request"`
}

type PostmanRequestDetails struct {
	Method string                   `json:"method"`
	URL    PostmanURL              `json:"url"`
	Header []PostmanHeader         `json:"header"`
	Body   PostmanBody             `json:"body,omitempty"`
}

type PostmanURL struct {
	Raw string `json:"raw"`
}

type PostmanHeader struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

type PostmanBody struct {
	Mode string `json:"mode"`
	Raw  string `json:"raw"`
}

type PostmanCollection struct {
	Info PostmanInfo      `json:"info"`
	Item []PostmanRequest `json:"item"`
}

type PostmanInfo struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

// ImportFromPostman imports a Postman collection JSON
func ImportFromPostman(data []byte) (*Collection, error) {
	var postman PostmanCollection
	if err := json.Unmarshal(data, &postman); err != nil {
		return nil, fmt.Errorf("failed to parse Postman collection: %w", err)
	}

	collection := CreateCollection(postman.Info.Name, postman.Info.Description)

	for _, item := range postman.Item {
		headers := make(map[string]string)
		for _, h := range item.Request.Header {
			headers[h.Key] = h.Value
		}

		body := ""
		if item.Request.Body.Mode == "raw" {
			body = item.Request.Body.Raw
		}

		now := time.Now()
		request := SavedRequest{
			ID:          uuid.New().String(),
			Name:        item.Name,
			Method:      item.Request.Method,
			URL:         item.Request.URL.Raw,
			Headers:     headers,
			Body:        body,
			QueryParams: make(map[string]string),
			CreatedAt:   now,
			LastUsed:    now,
		}

		AddRequestToCollection(&collection, request)
	}

	return &collection, nil
}

// ExportToPostman exports a collection to Postman format
func ExportToPostman(collection *Collection) ([]byte, error) {
	postman := PostmanCollection{
		Info: PostmanInfo{
			Name:        collection.Name,
			Description: collection.Description,
		},
		Item: []PostmanRequest{},
	}

	for _, req := range collection.Requests {
		headers := []PostmanHeader{}
		for k, v := range req.Headers {
			headers = append(headers, PostmanHeader{Key: k, Value: v})
		}

		body := PostmanBody{}
		if req.Body != "" {
			body.Mode = "raw"
			body.Raw = req.Body
		}

		postman.Item = append(postman.Item, PostmanRequest{
			Name: req.Name,
			Request: PostmanRequestDetails{
				Method: req.Method,
				URL:    PostmanURL{Raw: req.URL},
				Header: headers,
				Body:   body,
			},
		})
	}

	return json.MarshalIndent(postman, "", "  ")
}
