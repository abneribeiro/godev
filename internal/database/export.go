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
	if err := os.MkdirAll(exportDir, 0755); err != nil {
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
	file, err := os.Create(filePath)
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

	if err := os.WriteFile(filePath, data, 0644); err != nil {
		return fmt.Errorf("failed to write JSON file: %w", err)
	}

	return nil
}

func exportToSQL(filePath string, result *QueryResult, tableName string) error {
	if tableName == "" {
		tableName = "exported_table"
	}

	var sql strings.Builder

	sql.WriteString(fmt.Sprintf("-- SQL Export generated at %s\n", time.Now().Format("2006-01-02 15:04:05")))
	sql.WriteString(fmt.Sprintf("-- Total rows: %d\n\n", len(result.Rows)))

	for _, row := range result.Rows {
		sql.WriteString(fmt.Sprintf("INSERT INTO %s (", tableName))

		for i, col := range result.Columns {
			if i > 0 {
				sql.WriteString(", ")
			}
			sql.WriteString(col)
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
				escapedValue := strings.ReplaceAll(value, "'", "''")
				sql.WriteString(fmt.Sprintf("'%s'", escapedValue))
			}
		}

		sql.WriteString(");\n")
	}

	if err := os.WriteFile(filePath, []byte(sql.String()), 0644); err != nil {
		return fmt.Errorf("failed to write SQL file: %w", err)
	}

	return nil
}

func isNumeric(s string) bool {
	if s == "" {
		return false
	}

	for i, r := range s {
		if i == 0 && (r == '-' || r == '+') {
			continue
		}
		if r == '.' {
			continue
		}
		if r < '0' || r > '9' {
			return false
		}
	}

	return true
}
