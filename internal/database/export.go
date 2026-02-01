package database

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type ExportFormat string

const (
	ExportFormatCSV  ExportFormat = "csv"
	ExportFormatJSON ExportFormat = "json"
	ExportFormatSQL  ExportFormat = "sql"
)

type ExportResult struct {
	FilePath string
	Format   ExportFormat
	RowCount int
	Error    error
}

func ExportQueryResult(result *QueryResult, format ExportFormat, tableName string) ExportResult {
	if result == nil || len(result.Columns) == 0 {
		return ExportResult{Error: fmt.Errorf("no data to export")}
	}

	homeDir, err := os.UserHomeDir()
	if err != nil {
		return ExportResult{Error: fmt.Errorf("failed to get home directory: %w", err)}
	}

	exportDir := filepath.Join(homeDir, ".godev", "exports")
	// Use secure directory permissions (0700 - only owner can access)
	if err := os.MkdirAll(exportDir, 0o700); err != nil {
		return ExportResult{Error: fmt.Errorf("failed to create export directory: %w", err)}
	}

	timestamp := time.Now().Format("20060102_150405")
	var fileName string
	var exportFunc func(string, *QueryResult, string) error

	switch format {
	case ExportFormatCSV:
		fileName = fmt.Sprintf("export_%s.csv", timestamp)
		exportFunc = exportToCSV
	case ExportFormatJSON:
		fileName = fmt.Sprintf("export_%s.json", timestamp)
		exportFunc = exportToJSON
	case ExportFormatSQL:
		fileName = fmt.Sprintf("export_%s.sql", timestamp)
		exportFunc = exportToSQL
	default:
		return ExportResult{Error: fmt.Errorf("unsupported export format: %s", format)}
	}

	filePath := filepath.Join(exportDir, fileName)

	if err := exportFunc(filePath, result, tableName); err != nil {
		return ExportResult{Error: err}
	}

	return ExportResult{
		FilePath: filePath,
		Format:   format,
		RowCount: len(result.Rows),
	}
}

func exportToCSV(filePath string, result *QueryResult, tableName string) error {
	// Create file with secure permissions (0600 - only owner can read/write)
	file, err := os.OpenFile(filePath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o600)
	if err != nil {
		return fmt.Errorf("failed to create CSV file: %w", err)
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	if err := writer.Write(result.Columns); err != nil {
		return fmt.Errorf("failed to write CSV header: %w", err)
	}

	for _, row := range result.Rows {
		if err := writer.Write(row); err != nil {
			return fmt.Errorf("failed to write CSV row: %w", err)
		}
	}

	return nil
}

func exportToJSON(filePath string, result *QueryResult, tableName string) error {
	records := make([]map[string]string, 0, len(result.Rows))

	for _, row := range result.Rows {
		record := make(map[string]string)
		for i, col := range result.Columns {
			if i < len(row) {
				record[col] = row[i]
			}
		}
		records = append(records, record)
	}

	data, err := json.MarshalIndent(records, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal JSON: %w", err)
	}

	// Use secure file permissions (0600 - only owner can read/write)
	if err := os.WriteFile(filePath, data, 0o600); err != nil {
		return fmt.Errorf("failed to write JSON file: %w", err)
	}

	return nil
}

// quoteIdentifier quotes a PostgreSQL identifier (table or column name)
func quoteIdentifier(name string) string {
	// Replace double quotes with double-double quotes and wrap in quotes
	escaped := strings.ReplaceAll(name, `"`, `""`)
	return fmt.Sprintf(`"%s"`, escaped)
}

// escapeSQLString escapes a string value for SQL
func escapeSQLString(value string) string {
	// Escape single quotes and backslashes
	value = strings.ReplaceAll(value, `\`, `\\`)
	value = strings.ReplaceAll(value, `'`, `''`)
	return value
}

func exportToSQL(filePath string, result *QueryResult, tableName string) error {
	if tableName == "" {
		tableName = "exported_table"
	}

	var sql strings.Builder

	sql.WriteString(fmt.Sprintf("-- SQL Export generated at %s\n", time.Now().Format("2006-01-02 15:04:05")))
	sql.WriteString(fmt.Sprintf("-- Total rows: %d\n\n", len(result.Rows)))

	// Quote table name to prevent SQL injection
	quotedTableName := quoteIdentifier(tableName)

	for _, row := range result.Rows {
		sql.WriteString(fmt.Sprintf("INSERT INTO %s (", quotedTableName))

		// Quote all column names
		for i, col := range result.Columns {
			if i > 0 {
				sql.WriteString(", ")
			}
			sql.WriteString(quoteIdentifier(col))
		}

		sql.WriteString(") VALUES (")

		for i, value := range row {
			if i > 0 {
				sql.WriteString(", ")
			}

			if value == "" || strings.ToUpper(value) == "NULL" {
				sql.WriteString("NULL")
			} else if isNumeric(value) {
				sql.WriteString(value)
			} else {
				escapedValue := escapeSQLString(value)
				sql.WriteString(fmt.Sprintf("'%s'", escapedValue))
			}
		}

		sql.WriteString(");\n")
	}

	// Use secure file permissions (0600 - only owner can read/write)
	if err := os.WriteFile(filePath, []byte(sql.String()), 0o600); err != nil {
		return fmt.Errorf("failed to write SQL file: %w", err)
	}

	return nil
}

// isNumeric checks if a string represents a valid numeric value
func isNumeric(s string) bool {
	if s == "" || s == "NULL" {
		return false
	}

	// Track state
	hasDigit := false
	hasDot := false
	hasE := false
	i := 0

	// Check for sign at the beginning
	if s[0] == '-' || s[0] == '+' {
		i++
		if i >= len(s) {
			return false
		}
	}

	for ; i < len(s); i++ {
		r := rune(s[i])

		switch {
		case r >= '0' && r <= '9':
			hasDigit = true
		case r == '.':
			// Only one dot allowed, and not after 'e'
			if hasDot || hasE {
				return false
			}
			hasDot = true
		case r == 'e' || r == 'E':
			// Only one 'e' allowed, and must have digit before it
			if hasE || !hasDigit {
				return false
			}
			hasE = true
			hasDigit = false // Reset for exponent part
			// Check for optional sign after 'e'
			if i+1 < len(s) && (s[i+1] == '+' || s[i+1] == '-') {
				i++
			}
		default:
			return false
		}
	}

	return hasDigit
}
