package database

import (
	"database/sql"
	"fmt"
	"regexp"
	"strings"
	"time"

	_ "github.com/lib/pq"
)

const (
	// MaxRowsInMemory limits the number of rows loaded to prevent OOM
	MaxRowsInMemory = 10000
	// DefaultPageSize for paginated queries
	DefaultPageSize = 1000
)

type ConnectionConfig struct {
	Host     string
	Port     int
	Database string
	User     string
	Password string
	SSLMode  string
}

// Validate validates the connection configuration
func (c *ConnectionConfig) Validate() error {
	if c.Host == "" {
		return fmt.Errorf("host cannot be empty")
	}
	if c.Port < 1 || c.Port > 65535 {
		return fmt.Errorf("invalid port: %d (must be between 1 and 65535)", c.Port)
	}
	if c.Database == "" {
		return fmt.Errorf("database name cannot be empty")
	}
	if c.User == "" {
		return fmt.Errorf("user cannot be empty")
	}
	// SSLMode validation
	validSSLModes := map[string]bool{
		"disable":     true,
		"require":     true,
		"verify-ca":   true,
		"verify-full": true,
	}
	if c.SSLMode != "" && !validSSLModes[c.SSLMode] {
		return fmt.Errorf("invalid sslmode: %s (must be disable, require, verify-ca, or verify-full)", c.SSLMode)
	}
	if c.SSLMode == "" {
		c.SSLMode = "disable"
	}
	return nil
}

type QueryResult struct {
	Columns       []string
	Rows          [][]string
	RowsAffected  int64
	ExecutionTime time.Duration
	Error         error
	Truncated     bool // Indicates if results were truncated due to MaxRowsInMemory
}

type TableInfo struct {
	Name    string
	Columns []ColumnInfo
}

type ColumnInfo struct {
	Name     string
	Type     string
	Nullable bool
}

type PostgresClient struct {
	db     *sql.DB
	config ConnectionConfig
}

func NewPostgresClient() *PostgresClient {
	return &PostgresClient{}
}

func (c *PostgresClient) Connect(config ConnectionConfig) error {
	// Validate configuration before attempting connection
	if err := config.Validate(); err != nil {
		return fmt.Errorf("invalid configuration: %w", err)
	}

	connStr := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		config.Host, config.Port, config.User, config.Password, config.Database, config.SSLMode)

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return fmt.Errorf("failed to open connection: %w", err)
	}

	// Set connection pool limits
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(5 * time.Minute)

	if err := db.Ping(); err != nil {
		db.Close()
		return fmt.Errorf("failed to ping database: %w", err)
	}

	c.db = db
	c.config = config
	return nil
}

func (c *PostgresClient) IsConnected() bool {
	return c.db != nil
}

func (c *PostgresClient) Close() error {
	if c.db != nil {
		return c.db.Close()
	}
	return nil
}

// isReadOnlyQuery checks if a query is a read-only operation
func isReadOnlyQuery(query string) bool {
	// Remove leading whitespace and comments
	query = strings.TrimSpace(query)
	query = removeComments(query)

	queryUpper := strings.ToUpper(query)

	// Check for read-only query types
	readOnlyPrefixes := []string{
		"SELECT", "SHOW", "EXPLAIN", "DESCRIBE", "DESC",
		"WITH", // CTE can be read-only if it ends with SELECT
	}

	for _, prefix := range readOnlyPrefixes {
		if strings.HasPrefix(queryUpper, prefix) {
			return true
		}
	}

	return false
}

// removeComments removes SQL comments from query
// Note: This is a simple implementation that doesn't handle comments inside strings
func removeComments(query string) string {
	// Remove multi-line comments (/* ... */) first
	// Use (?s) flag to make . match newlines
	re := regexp.MustCompile(`(?s)/\*.*?\*/`)
	result := re.ReplaceAllString(query, " ")

	// Remove single-line comments (-- ...)
	// Note: This will incorrectly remove -- inside string literals
	// For production use, a proper SQL parser should be used
	lines := strings.Split(result, "\n")
	var cleaned []string
	for _, line := range lines {
		if idx := strings.Index(line, "--"); idx != -1 {
			line = line[:idx]
		}
		if strings.TrimSpace(line) != "" {
			cleaned = append(cleaned, line)
		}
	}

	return strings.TrimSpace(strings.Join(cleaned, "\n"))
}

func (c *PostgresClient) ExecuteQuery(query string) QueryResult {
	if c.db == nil {
		return QueryResult{Error: fmt.Errorf("not connected to database")}
	}

	startTime := time.Now()

	query = strings.TrimSpace(query)
	if query == "" {
		return QueryResult{Error: fmt.Errorf("query cannot be empty")}
	}

	// Detect if query returns rows (SELECT-like) or just affects rows (INSERT/UPDATE/DELETE)
	if isReadOnlyQuery(query) {
		return c.executeSelectQuery(query, startTime)
	}

	return c.executeNonSelectQuery(query, startTime)
}

// formatValue converts a database value to a string representation
func formatValue(val interface{}) string {
	if val == nil {
		return "NULL"
	}

	switch v := val.(type) {
	case []byte:
		return string(v)
	case time.Time:
		return v.Format("2006-01-02 15:04:05.999999")
	case bool:
		if v {
			return "true"
		}
		return "false"
	case int64, int32, int16, int8, int:
		return fmt.Sprintf("%d", v)
	case uint64, uint32, uint16, uint8, uint:
		return fmt.Sprintf("%d", v)
	case float64, float32:
		return fmt.Sprintf("%g", v)
	case string:
		return v
	default:
		return fmt.Sprintf("%v", v)
	}
}

func (c *PostgresClient) executeSelectQuery(query string, startTime time.Time) QueryResult {
	rows, err := c.db.Query(query)
	if err != nil {
		return QueryResult{
			Error:         err,
			ExecutionTime: time.Since(startTime),
		}
	}
	defer rows.Close()

	columns, err := rows.Columns()
	if err != nil {
		return QueryResult{
			Error:         err,
			ExecutionTime: time.Since(startTime),
		}
	}

	var resultRows [][]string
	rowCount := 0
	truncated := false

	for rows.Next() {
		// Limit rows to prevent OOM
		if rowCount >= MaxRowsInMemory {
			truncated = true
			break
		}

		values := make([]interface{}, len(columns))
		valuePtrs := make([]interface{}, len(columns))
		for i := range values {
			valuePtrs[i] = &values[i]
		}

		if err := rows.Scan(valuePtrs...); err != nil {
			return QueryResult{
				Error:         err,
				ExecutionTime: time.Since(startTime),
			}
		}

		row := make([]string, len(columns))
		for i, val := range values {
			row[i] = formatValue(val)
		}
		resultRows = append(resultRows, row)
		rowCount++
	}

	if err := rows.Err(); err != nil {
		return QueryResult{
			Error:         err,
			ExecutionTime: time.Since(startTime),
		}
	}

	return QueryResult{
		Columns:       columns,
		Rows:          resultRows,
		RowsAffected:  int64(len(resultRows)),
		ExecutionTime: time.Since(startTime),
		Truncated:     truncated,
	}
}

func (c *PostgresClient) executeNonSelectQuery(query string, startTime time.Time) QueryResult {
	result, err := c.db.Exec(query)
	if err != nil {
		return QueryResult{
			Error:         err,
			ExecutionTime: time.Since(startTime),
		}
	}

	rowsAffected, _ := result.RowsAffected()

	return QueryResult{
		RowsAffected:  rowsAffected,
		ExecutionTime: time.Since(startTime),
	}
}

func (c *PostgresClient) GetTables() ([]string, error) {
	if c.db == nil {
		return nil, fmt.Errorf("not connected to database")
	}

	query := `
		SELECT table_name
		FROM information_schema.tables
		WHERE table_schema = 'public'
		ORDER BY table_name
	`

	rows, err := c.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tables []string
	for rows.Next() {
		var tableName string
		if err := rows.Scan(&tableName); err != nil {
			return nil, err
		}
		tables = append(tables, tableName)
	}

	return tables, nil
}

func (c *PostgresClient) GetTableInfo(tableName string) (*TableInfo, error) {
	if c.db == nil {
		return nil, fmt.Errorf("not connected to database")
	}

	query := `
		SELECT column_name, data_type, is_nullable
		FROM information_schema.columns
		WHERE table_schema = 'public' AND table_name = $1
		ORDER BY ordinal_position
	`

	rows, err := c.db.Query(query, tableName)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	tableInfo := &TableInfo{
		Name:    tableName,
		Columns: []ColumnInfo{},
	}

	for rows.Next() {
		var col ColumnInfo
		var nullable string
		if err := rows.Scan(&col.Name, &col.Type, &nullable); err != nil {
			return nil, err
		}
		col.Nullable = nullable == "YES"
		tableInfo.Columns = append(tableInfo.Columns, col)
	}

	return tableInfo, nil
}

func (c *PostgresClient) GetConnectionString() string {
	if c.db == nil {
		return "Not connected"
	}
	return fmt.Sprintf("%s@%s:%d/%s", c.config.User, c.config.Host, c.config.Port, c.config.Database)
}
