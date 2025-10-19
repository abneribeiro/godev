package database

import (
	"database/sql"
	"fmt"
	"strings"
	"time"

	_ "github.com/lib/pq"
)

type ConnectionConfig struct {
	Host     string
	Port     int
	Database string
	User     string
	Password string
	SSLMode  string
}

type QueryResult struct {
	Columns      []string
	Rows         [][]string
	RowsAffected int64
	ExecutionTime time.Duration
	Error        error
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
	connStr := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		config.Host, config.Port, config.User, config.Password, config.Database, config.SSLMode)

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return fmt.Errorf("failed to open connection: %w", err)
	}

	if err := db.Ping(); err != nil {
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

func (c *PostgresClient) ExecuteQuery(query string) QueryResult {
	if c.db == nil {
		return QueryResult{Error: fmt.Errorf("not connected to database")}
	}

	startTime := time.Now()

	query = strings.TrimSpace(query)
	if query == "" {
		return QueryResult{Error: fmt.Errorf("query cannot be empty")}
	}

	queryUpper := strings.ToUpper(query)
	if strings.HasPrefix(queryUpper, "SELECT") || strings.HasPrefix(queryUpper, "SHOW") {
		return c.executeSelectQuery(query, startTime)
	}

	return c.executeNonSelectQuery(query, startTime)
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
	for rows.Next() {
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
			if val == nil {
				row[i] = "NULL"
			} else {
				row[i] = fmt.Sprintf("%v", val)
			}
		}
		resultRows = append(resultRows, row)
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
