package database

import (
	"strings"
	"testing"
)

func TestQueryBuilderSelect(t *testing.T) {
	qb := NewQueryBuilder()
	query, err := qb.Select("id", "name", "email").
		From("users").
		Build()

	if err != nil {
		t.Fatalf("Build failed: %v", err)
	}

	expected := "SELECT id, name, email\nFROM users;"
	if query != expected {
		t.Errorf("Expected query:\n%s\nGot:\n%s", expected, query)
	}
}

func TestQueryBuilderSelectAll(t *testing.T) {
	qb := NewQueryBuilder()
	query, err := qb.Select().From("users").Build()

	if err != nil {
		t.Fatalf("Build failed: %v", err)
	}

	if !containsQueryStr(query, "SELECT *") {
		t.Error("Expected SELECT *")
	}
}

func TestQueryBuilderWhere(t *testing.T) {
	qb := NewQueryBuilder()
	query, err := qb.Select("*").
		From("users").
		Where("age", ">", 18).
		Where("active", "=", true).
		Build()

	if err != nil {
		t.Fatalf("Build failed: %v", err)
	}

	if !containsQueryStr(query, "WHERE") {
		t.Error("Expected WHERE clause")
	}

	if !containsQueryStr(query, "age > 18") {
		t.Error("Expected age > 18 condition")
	}

	if !containsQueryStr(query, "active = TRUE") {
		t.Error("Expected active = TRUE condition")
	}

	if !containsQueryStr(query, "AND") {
		t.Error("Expected AND operator between conditions")
	}
}

func TestQueryBuilderOrWhere(t *testing.T) {
	qb := NewQueryBuilder()
	query, err := qb.Select("*").
		From("users").
		Where("role", "=", "admin").
		OrWhere("role", "=", "moderator").
		Build()

	if err != nil {
		t.Fatalf("Build failed: %v", err)
	}

	if !containsQueryStr(query, "WHERE") {
		t.Error("Expected WHERE clause")
	}

	if !containsQueryStr(query, "OR") {
		t.Error("Expected OR operator")
	}
}

func TestQueryBuilderWhereNull(t *testing.T) {
	qb := NewQueryBuilder()
	query, err := qb.Select("*").
		From("users").
		WhereNull("deleted_at").
		Build()

	if err != nil {
		t.Fatalf("Build failed: %v", err)
	}

	if !containsQueryStr(query, "deleted_at IS NULL") {
		t.Error("Expected 'deleted_at IS NULL' condition")
	}
}

func TestQueryBuilderWhereNotNull(t *testing.T) {
	qb := NewQueryBuilder()
	query, err := qb.Select("*").
		From("users").
		WhereNotNull("email").
		Build()

	if err != nil {
		t.Fatalf("Build failed: %v", err)
	}

	if !containsQueryStr(query, "email IS NOT NULL") {
		t.Error("Expected 'email IS NOT NULL' condition")
	}
}

func TestQueryBuilderWhereIn(t *testing.T) {
	qb := NewQueryBuilder()
	query, err := qb.Select("*").
		From("users").
		WhereIn("id", []interface{}{1, 2, 3}).
		Build()

	if err != nil {
		t.Fatalf("Build failed: %v", err)
	}

	if !containsQueryStr(query, "id IN (1, 2, 3)") {
		t.Error("Expected 'id IN (1, 2, 3)' condition")
	}
}

func TestQueryBuilderJoin(t *testing.T) {
	qb := NewQueryBuilder()
	query, err := qb.Select("users.name", "orders.total").
		From("users").
		Join("orders", "users.id", "orders.user_id").
		Build()

	if err != nil {
		t.Fatalf("Build failed: %v", err)
	}

	if !containsQueryStr(query, "INNER JOIN orders") {
		t.Error("Expected INNER JOIN")
	}

	if !containsQueryStr(query, "ON users.id = orders.user_id") {
		t.Error("Expected ON clause with correct join condition")
	}
}

func TestQueryBuilderLeftJoin(t *testing.T) {
	qb := NewQueryBuilder()
	query, err := qb.Select("*").
		From("users").
		LeftJoin("profiles", "users.id", "profiles.user_id").
		Build()

	if err != nil {
		t.Fatalf("Build failed: %v", err)
	}

	if !containsQueryStr(query, "LEFT JOIN profiles") {
		t.Error("Expected LEFT JOIN")
	}
}

func TestQueryBuilderOrderBy(t *testing.T) {
	qb := NewQueryBuilder()
	query, err := qb.Select("*").
		From("users").
		OrderBy("name", "ASC").
		OrderBy("created_at", "DESC").
		Build()

	if err != nil {
		t.Fatalf("Build failed: %v", err)
	}

	if !containsQueryStr(query, "ORDER BY") {
		t.Error("Expected ORDER BY clause")
	}

	if !containsQueryStr(query, "name ASC") {
		t.Error("Expected 'name ASC'")
	}

	if !containsQueryStr(query, "created_at DESC") {
		t.Error("Expected 'created_at DESC'")
	}
}

func TestQueryBuilderGroupBy(t *testing.T) {
	qb := NewQueryBuilder()
	query, err := qb.Select("status", "COUNT(*)").
		From("orders").
		GroupBy("status").
		Build()

	if err != nil {
		t.Fatalf("Build failed: %v", err)
	}

	if !containsQueryStr(query, "GROUP BY status") {
		t.Error("Expected GROUP BY clause")
	}
}

func TestQueryBuilderHaving(t *testing.T) {
	qb := NewQueryBuilder()
	query, err := qb.Select("status", "COUNT(*)").
		From("orders").
		GroupBy("status").
		Having("COUNT(*)", ">", 10).
		Build()

	if err != nil {
		t.Fatalf("Build failed: %v", err)
	}

	if !containsQueryStr(query, "HAVING") {
		t.Error("Expected HAVING clause")
	}
}

func TestQueryBuilderLimit(t *testing.T) {
	qb := NewQueryBuilder()
	query, err := qb.Select("*").
		From("users").
		Limit(10).
		Build()

	if err != nil {
		t.Fatalf("Build failed: %v", err)
	}

	if !containsQueryStr(query, "LIMIT 10") {
		t.Error("Expected LIMIT 10")
	}
}

func TestQueryBuilderOffset(t *testing.T) {
	qb := NewQueryBuilder()
	query, err := qb.Select("*").
		From("users").
		Limit(10).
		Offset(20).
		Build()

	if err != nil {
		t.Fatalf("Build failed: %v", err)
	}

	if !containsQueryStr(query, "OFFSET 20") {
		t.Error("Expected OFFSET 20")
	}
}

func TestQueryBuilderInsert(t *testing.T) {
	qb := NewQueryBuilder()
	query, err := qb.Insert("users").
		Values(map[string]interface{}{
			"name":  "John Doe",
			"email": "john@example.com",
			"age":   30,
		}).
		Build()

	if err != nil {
		t.Fatalf("Build failed: %v", err)
	}

	if !containsQueryStr(query, "INSERT INTO users") {
		t.Error("Expected INSERT INTO users")
	}

	if !containsQueryStr(query, "VALUES") {
		t.Error("Expected VALUES clause")
	}

	// Check that all values are present
	if !containsQueryStr(query, "'John Doe'") {
		t.Error("Expected name value")
	}

	if !containsQueryStr(query, "'john@example.com'") {
		t.Error("Expected email value")
	}

	if !containsQueryStr(query, "30") {
		t.Error("Expected age value")
	}
}

func TestQueryBuilderUpdate(t *testing.T) {
	qb := NewQueryBuilder()
	query, err := qb.Update("users").
		Set("name", "Jane Doe").
		Set("age", 25).
		Where("id", "=", 1).
		Build()

	if err != nil {
		t.Fatalf("Build failed: %v", err)
	}

	if !containsQueryStr(query, "UPDATE users") {
		t.Error("Expected UPDATE users")
	}

	if !containsQueryStr(query, "SET") {
		t.Error("Expected SET clause")
	}

	if !containsQueryStr(query, "name = 'Jane Doe'") {
		t.Error("Expected name update")
	}

	if !containsQueryStr(query, "age = 25") {
		t.Error("Expected age update")
	}

	if !containsQueryStr(query, "WHERE id = 1") {
		t.Error("Expected WHERE clause")
	}
}

func TestQueryBuilderDelete(t *testing.T) {
	qb := NewQueryBuilder()
	query, err := qb.Delete().
		From("users").
		Where("inactive", "=", true).
		Build()

	if err != nil {
		t.Fatalf("Build failed: %v", err)
	}

	if !containsQueryStr(query, "DELETE FROM users") {
		t.Error("Expected DELETE FROM users")
	}

	if !containsQueryStr(query, "WHERE inactive = TRUE") {
		t.Error("Expected WHERE clause")
	}
}

func TestQueryBuilderComplexQuery(t *testing.T) {
	qb := NewQueryBuilder()
	query, err := qb.Select("u.name", "COUNT(o.id) as order_count").
		From("users u").
		LeftJoin("orders o", "u.id", "o.user_id").
		Where("u.active", "=", true).
		Where("u.created_at", ">", "2024-01-01").
		GroupBy("u.id", "u.name").
		Having("COUNT(o.id)", ">", 5).
		OrderBy("order_count", "DESC").
		Limit(10).
		Build()

	if err != nil {
		t.Fatalf("Build failed: %v", err)
	}

	// Verify all parts are present
	requiredParts := []string{
		"SELECT", "u.name", "COUNT(o.id)", "order_count",
		"FROM", "users u",
		"LEFT JOIN", "orders o", "ON u.id = o.user_id",
		"WHERE", "u.active = TRUE", "AND",
		"GROUP BY", "u.id", "u.name",
		"HAVING", "COUNT(o.id) > 5",
		"ORDER BY", "order_count DESC",
		"LIMIT 10",
	}

	for _, part := range requiredParts {
		if !containsQueryStr(query, part) {
			t.Errorf("Expected query to contain '%s'\nQuery: %s", part, query)
		}
	}
}

func TestFormatValueForSQL(t *testing.T) {
	tests := []struct {
		name     string
		value    interface{}
		expected string
	}{
		{"nil", nil, "NULL"},
		{"string", "test", "'test'"},
		{"string with quotes", "it's", "'it''s'"},
		{"int", 42, "42"},
		{"float", 3.14, "3.14"},
		{"bool true", true, "TRUE"},
		{"bool false", false, "FALSE"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := formatValueForSQL(tt.value)
			if result != tt.expected {
				t.Errorf("Expected '%s', got '%s'", tt.expected, result)
			}
		})
	}
}

func containsQueryStr(s, substr string) bool {
	return strings.Contains(s, substr)
}
