package storage

import (
	"encoding/json"
	"testing"
	"time"
)

func TestCreateCollection(t *testing.T) {
	collection := CreateCollection("Test Collection", "A test collection")

	if collection.Name != "Test Collection" {
		t.Errorf("Expected name 'Test Collection', got '%s'", collection.Name)
	}

	if collection.Description != "A test collection" {
		t.Errorf("Expected description 'A test collection', got '%s'", collection.Description)
	}

	if len(collection.Requests) != 0 {
		t.Errorf("Expected 0 requests, got %d", len(collection.Requests))
	}

	if collection.ID == "" {
		t.Error("Expected non-empty ID")
	}
}

func TestAddRequestToCollection(t *testing.T) {
	collection := CreateCollection("Test", "Test")

	request := SavedRequest{
		ID:     "req1",
		Name:   "Test Request",
		Method: "GET",
		URL:    "https://api.example.com",
	}

	AddRequestToCollection(&collection, request)

	if len(collection.Requests) != 1 {
		t.Errorf("Expected 1 request, got %d", len(collection.Requests))
	}

	if collection.Requests[0].ID != "req1" {
		t.Errorf("Expected request ID 'req1', got '%s'", collection.Requests[0].ID)
	}
}

func TestRemoveRequestFromCollection(t *testing.T) {
	collection := CreateCollection("Test", "Test")

	request1 := SavedRequest{ID: "req1", Name: "Request 1"}
	request2 := SavedRequest{ID: "req2", Name: "Request 2"}

	AddRequestToCollection(&collection, request1)
	AddRequestToCollection(&collection, request2)

	err := RemoveRequestFromCollection(&collection, "req1")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	if len(collection.Requests) != 1 {
		t.Errorf("Expected 1 request, got %d", len(collection.Requests))
	}

	if collection.Requests[0].ID != "req2" {
		t.Errorf("Expected remaining request to be 'req2', got '%s'", collection.Requests[0].ID)
	}
}

func TestRemoveNonExistentRequest(t *testing.T) {
	collection := CreateCollection("Test", "Test")

	err := RemoveRequestFromCollection(&collection, "nonexistent")
	if err == nil {
		t.Error("Expected error when removing non-existent request")
	}
}

func TestAddSubCollection(t *testing.T) {
	parent := CreateCollection("Parent", "Parent collection")
	child := CreateCollection("Child", "Child collection")

	AddSubCollection(&parent, child)

	if len(parent.SubCollections) != 1 {
		t.Errorf("Expected 1 sub-collection, got %d", len(parent.SubCollections))
	}

	if parent.SubCollections[0].Name != "Child" {
		t.Errorf("Expected sub-collection name 'Child', got '%s'", parent.SubCollections[0].Name)
	}
}

func TestFindCollectionByID(t *testing.T) {
	parent := CreateCollection("Parent", "Parent")
	child1 := CreateCollection("Child1", "Child 1")
	child2 := CreateCollection("Child2", "Child 2")

	// Add child2 to child1 first
	AddSubCollection(&child1, child2)
	// Then add child1 to parent
	AddSubCollection(&parent, child1)

	collections := []Collection{parent}

	// Find parent
	found := FindCollectionByID(collections, parent.ID)
	if found == nil {
		t.Error("Expected to find parent collection")
	} else if found.ID != parent.ID {
		t.Errorf("Expected to find collection with ID '%s', got '%s'", parent.ID, found.ID)
	}

	// Find child
	found = FindCollectionByID(collections, child1.ID)
	if found == nil {
		t.Error("Expected to find child collection")
	} else if found.ID != child1.ID {
		t.Errorf("Expected to find collection with ID '%s', got '%s'", child1.ID, found.ID)
	}

	// Find nested child
	found = FindCollectionByID(collections, child2.ID)
	if found == nil {
		t.Error("Expected to find nested child collection")
	} else if found.ID != child2.ID {
		t.Errorf("Expected to find collection with ID '%s', got '%s'", child2.ID, found.ID)
	}

	// Find non-existent
	found = FindCollectionByID(collections, "nonexistent")
	if found != nil {
		t.Error("Expected nil for non-existent collection")
	}
}

func TestImportFromPostman(t *testing.T) {
	postmanJSON := `{
		"info": {
			"name": "Test API",
			"description": "A test API collection"
		},
		"item": [
			{
				"name": "Get Users",
				"request": {
					"method": "GET",
					"url": {
						"raw": "https://api.example.com/users"
					},
					"header": [
						{
							"key": "Authorization",
							"value": "Bearer token123"
						}
					]
				}
			},
			{
				"name": "Create User",
				"request": {
					"method": "POST",
					"url": {
						"raw": "https://api.example.com/users"
					},
					"header": [
						{
							"key": "Content-Type",
							"value": "application/json"
						}
					],
					"body": {
						"mode": "raw",
						"raw": "{\"name\": \"John Doe\"}"
					}
				}
			}
		]
	}`

	collection, err := ImportFromPostman([]byte(postmanJSON))
	if err != nil {
		t.Fatalf("Failed to import Postman collection: %v", err)
	}

	if collection.Name != "Test API" {
		t.Errorf("Expected name 'Test API', got '%s'", collection.Name)
	}

	if collection.Description != "A test API collection" {
		t.Errorf("Expected description 'A test API collection', got '%s'", collection.Description)
	}

	if len(collection.Requests) != 2 {
		t.Fatalf("Expected 2 requests, got %d", len(collection.Requests))
	}

	// Check first request
	req1 := collection.Requests[0]
	if req1.Name != "Get Users" {
		t.Errorf("Expected request name 'Get Users', got '%s'", req1.Name)
	}
	if req1.Method != "GET" {
		t.Errorf("Expected method 'GET', got '%s'", req1.Method)
	}
	if req1.URL != "https://api.example.com/users" {
		t.Errorf("Expected URL 'https://api.example.com/users', got '%s'", req1.URL)
	}
	if req1.Headers["Authorization"] != "Bearer token123" {
		t.Errorf("Expected Authorization header 'Bearer token123', got '%s'", req1.Headers["Authorization"])
	}

	// Check second request
	req2 := collection.Requests[1]
	if req2.Name != "Create User" {
		t.Errorf("Expected request name 'Create User', got '%s'", req2.Name)
	}
	if req2.Method != "POST" {
		t.Errorf("Expected method 'POST', got '%s'", req2.Method)
	}
	if req2.Body != "{\"name\": \"John Doe\"}" {
		t.Errorf("Expected body '{\"name\": \"John Doe\"}', got '%s'", req2.Body)
	}
}

func TestExportToPostman(t *testing.T) {
	collection := CreateCollection("Test API", "Test collection")

	request := SavedRequest{
		ID:     "req1",
		Name:   "Get Users",
		Method: "GET",
		URL:    "https://api.example.com/users",
		Headers: map[string]string{
			"Authorization": "Bearer token",
		},
		Body:        "",
		QueryParams: make(map[string]string),
		CreatedAt:   time.Now(),
		LastUsed:    time.Now(),
	}

	AddRequestToCollection(&collection, request)

	postmanJSON, err := ExportToPostman(&collection)
	if err != nil {
		t.Fatalf("Failed to export to Postman: %v", err)
	}

	// Parse the JSON to verify structure
	var postman PostmanCollection
	if err := json.Unmarshal(postmanJSON, &postman); err != nil {
		t.Fatalf("Failed to parse exported JSON: %v", err)
	}

	if postman.Info.Name != "Test API" {
		t.Errorf("Expected name 'Test API', got '%s'", postman.Info.Name)
	}

	if len(postman.Item) != 1 {
		t.Fatalf("Expected 1 item, got %d", len(postman.Item))
	}

	item := postman.Item[0]
	if item.Name != "Get Users" {
		t.Errorf("Expected item name 'Get Users', got '%s'", item.Name)
	}
	if item.Request.Method != "GET" {
		t.Errorf("Expected method 'GET', got '%s'", item.Request.Method)
	}
	if item.Request.URL.Raw != "https://api.example.com/users" {
		t.Errorf("Expected URL 'https://api.example.com/users', got '%s'", item.Request.URL.Raw)
	}
}

func TestImportInvalidPostman(t *testing.T) {
	invalidJSON := `{"invalid": "json"`

	_, err := ImportFromPostman([]byte(invalidJSON))
	if err == nil {
		t.Error("Expected error when importing invalid JSON")
	}
}
