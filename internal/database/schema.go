package database

import (
	"fmt"
	"strings"
)

// SchemaInfo represents complete schema information
type SchemaInfo struct {
	Tables        []TableMetadata
	Relationships []ForeignKeyRelationship
}

// TableMetadata represents detailed table information
type TableMetadata struct {
	Name          string
	Schema        string
	Columns       []ColumnMetadata
	PrimaryKeys   []string
	ForeignKeys   []ForeignKeyMetadata
	Indexes       []IndexMetadata
	Constraints   []ConstraintMetadata
	RowCount      int64
	TableSize     string
}

// ColumnMetadata represents detailed column information
type ColumnMetadata struct {
	Name         string
	Type         string
	Nullable     bool
	DefaultValue string
	IsPrimaryKey bool
	IsForeignKey bool
	IsUnique     bool
	MaxLength    int
	Precision    int
	Scale        int
	Comment      string
}

// ForeignKeyMetadata represents a foreign key constraint
type ForeignKeyMetadata struct {
	Name             string
	ColumnName       string
	ReferencedTable  string
	ReferencedColumn string
	OnDelete         string
	OnUpdate         string
}

// IndexMetadata represents an index
type IndexMetadata struct {
	Name      string
	Columns   []string
	IsUnique  bool
	IsPrimary bool
	Type      string
}

// ConstraintMetadata represents a constraint
type ConstraintMetadata struct {
	Name       string
	Type       string // PRIMARY, FOREIGN, UNIQUE, CHECK
	Definition string
}

// ForeignKeyRelationship represents a relationship between tables
type ForeignKeyRelationship struct {
	FromTable     string
	FromColumn    string
	ToTable       string
	ToColumn      string
	Constraint    string
	OnDelete      string
	OnUpdate      string
}

// GetTableMetadata retrieves detailed metadata for a table
func (c *PostgresClient) GetTableMetadata(tableName string) (*TableMetadata, error) {
	if c.db == nil {
		return nil, fmt.Errorf("not connected to database")
	}

	metadata := &TableMetadata{
		Name:   tableName,
		Schema: "public",
	}

	// Get columns with detailed info
	columns, err := c.getTableColumns(tableName)
	if err != nil {
		return nil, fmt.Errorf("failed to get columns: %w", err)
	}
	metadata.Columns = columns

	// Get primary keys
	primaryKeys, err := c.getTablePrimaryKeys(tableName)
	if err != nil {
		return nil, fmt.Errorf("failed to get primary keys: %w", err)
	}
	metadata.PrimaryKeys = primaryKeys

	// Get foreign keys
	foreignKeys, err := c.getTableForeignKeys(tableName)
	if err != nil {
		return nil, fmt.Errorf("failed to get foreign keys: %w", err)
	}
	metadata.ForeignKeys = foreignKeys

	// Get indexes
	indexes, err := c.getTableIndexes(tableName)
	if err != nil {
		return nil, fmt.Errorf("failed to get indexes: %w", err)
	}
	metadata.Indexes = indexes

	// Get constraints
	constraints, err := c.getTableConstraints(tableName)
	if err != nil {
		return nil, fmt.Errorf("failed to get constraints: %w", err)
	}
	metadata.Constraints = constraints

	// Get row count and size
	metadata.RowCount, metadata.TableSize, _ = c.getTableStats(tableName)

	return metadata, nil
}

// getTableColumns retrieves detailed column information
func (c *PostgresClient) getTableColumns(tableName string) ([]ColumnMetadata, error) {
	query := `
		SELECT
			c.column_name,
			c.data_type,
			c.is_nullable,
			COALESCE(c.column_default, ''),
			COALESCE(c.character_maximum_length, 0),
			COALESCE(c.numeric_precision, 0),
			COALESCE(c.numeric_scale, 0),
			COALESCE(pgd.description, '')
		FROM information_schema.columns c
		LEFT JOIN pg_catalog.pg_statio_all_tables as st
			ON c.table_schema = st.schemaname
			AND c.table_name = st.relname
		LEFT JOIN pg_catalog.pg_description pgd
			ON pgd.objoid = st.relid
			AND pgd.objsubid = c.ordinal_position
		WHERE c.table_name = $1
			AND c.table_schema = 'public'
		ORDER BY c.ordinal_position
	`

	rows, err := c.db.Query(query, tableName)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var columns []ColumnMetadata
	for rows.Next() {
		var col ColumnMetadata
		var nullable string

		err := rows.Scan(
			&col.Name,
			&col.Type,
			&nullable,
			&col.DefaultValue,
			&col.MaxLength,
			&col.Precision,
			&col.Scale,
			&col.Comment,
		)
		if err != nil {
			return nil, err
		}

		col.Nullable = (nullable == "YES")
		columns = append(columns, col)
	}

	return columns, rows.Err()
}

// getTablePrimaryKeys retrieves primary key columns
func (c *PostgresClient) getTablePrimaryKeys(tableName string) ([]string, error) {
	query := `
		SELECT a.attname
		FROM pg_index i
		JOIN pg_attribute a ON a.attrelid = i.indrelid AND a.attnum = ANY(i.indkey)
		WHERE i.indrelid = $1::regclass
			AND i.indisprimary
		ORDER BY array_position(i.indkey, a.attnum)
	`

	rows, err := c.db.Query(query, tableName)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var primaryKeys []string
	for rows.Next() {
		var colName string
		if err := rows.Scan(&colName); err != nil {
			return nil, err
		}
		primaryKeys = append(primaryKeys, colName)
	}

	return primaryKeys, rows.Err()
}

// getTableForeignKeys retrieves foreign key constraints
func (c *PostgresClient) getTableForeignKeys(tableName string) ([]ForeignKeyMetadata, error) {
	query := `
		SELECT
			tc.constraint_name,
			kcu.column_name,
			ccu.table_name AS referenced_table,
			ccu.column_name AS referenced_column,
			rc.delete_rule,
			rc.update_rule
		FROM information_schema.table_constraints AS tc
		JOIN information_schema.key_column_usage AS kcu
			ON tc.constraint_name = kcu.constraint_name
			AND tc.table_schema = kcu.table_schema
		JOIN information_schema.constraint_column_usage AS ccu
			ON ccu.constraint_name = tc.constraint_name
			AND ccu.table_schema = tc.table_schema
		JOIN information_schema.referential_constraints AS rc
			ON tc.constraint_name = rc.constraint_name
		WHERE tc.constraint_type = 'FOREIGN KEY'
			AND tc.table_name = $1
			AND tc.table_schema = 'public'
	`

	rows, err := c.db.Query(query, tableName)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var foreignKeys []ForeignKeyMetadata
	for rows.Next() {
		var fk ForeignKeyMetadata
		err := rows.Scan(
			&fk.Name,
			&fk.ColumnName,
			&fk.ReferencedTable,
			&fk.ReferencedColumn,
			&fk.OnDelete,
			&fk.OnUpdate,
		)
		if err != nil {
			return nil, err
		}
		foreignKeys = append(foreignKeys, fk)
	}

	return foreignKeys, rows.Err()
}

// getTableIndexes retrieves table indexes
func (c *PostgresClient) getTableIndexes(tableName string) ([]IndexMetadata, error) {
	query := `
		SELECT
			i.relname AS index_name,
			array_agg(a.attname ORDER BY a.attnum) AS column_names,
			ix.indisunique AS is_unique,
			ix.indisprimary AS is_primary,
			am.amname AS index_type
		FROM pg_class t
		JOIN pg_index ix ON t.oid = ix.indrelid
		JOIN pg_class i ON i.oid = ix.indexrelid
		JOIN pg_attribute a ON a.attrelid = t.oid AND a.attnum = ANY(ix.indkey)
		JOIN pg_am am ON i.relam = am.oid
		WHERE t.relname = $1
			AND t.relnamespace = (SELECT oid FROM pg_namespace WHERE nspname = 'public')
		GROUP BY i.relname, ix.indisunique, ix.indisprimary, am.amname
		ORDER BY i.relname
	`

	rows, err := c.db.Query(query, tableName)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var indexes []IndexMetadata
	for rows.Next() {
		var idx IndexMetadata
		var columns string

		err := rows.Scan(
			&idx.Name,
			&columns,
			&idx.IsUnique,
			&idx.IsPrimary,
			&idx.Type,
		)
		if err != nil {
			return nil, err
		}

		// Parse columns array (format: {col1,col2})
		columns = strings.Trim(columns, "{}")
		idx.Columns = strings.Split(columns, ",")

		indexes = append(indexes, idx)
	}

	return indexes, rows.Err()
}

// getTableConstraints retrieves table constraints
func (c *PostgresClient) getTableConstraints(tableName string) ([]ConstraintMetadata, error) {
	query := `
		SELECT
			tc.constraint_name,
			tc.constraint_type,
			COALESCE(cc.check_clause, '')
		FROM information_schema.table_constraints tc
		LEFT JOIN information_schema.check_constraints cc
			ON tc.constraint_name = cc.constraint_name
		WHERE tc.table_name = $1
			AND tc.table_schema = 'public'
		ORDER BY tc.constraint_type, tc.constraint_name
	`

	rows, err := c.db.Query(query, tableName)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var constraints []ConstraintMetadata
	for rows.Next() {
		var con ConstraintMetadata
		err := rows.Scan(&con.Name, &con.Type, &con.Definition)
		if err != nil {
			return nil, err
		}
		constraints = append(constraints, con)
	}

	return constraints, rows.Err()
}

// getTableStats retrieves table statistics
func (c *PostgresClient) getTableStats(tableName string) (int64, string, error) {
	var rowCount int64
	var tableSize string

	// Get row count
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM %s", quoteIdentifier(tableName))
	err := c.db.QueryRow(countQuery).Scan(&rowCount)
	if err != nil {
		rowCount = -1
	}

	// Get table size
	sizeQuery := `
		SELECT pg_size_pretty(pg_total_relation_size($1::regclass))
	`
	err = c.db.QueryRow(sizeQuery, tableName).Scan(&tableSize)
	if err != nil {
		tableSize = "unknown"
	}

	return rowCount, tableSize, nil
}

// GetDatabaseSchema retrieves complete schema information including relationships
func (c *PostgresClient) GetDatabaseSchema() (*SchemaInfo, error) {
	if c.db == nil {
		return nil, fmt.Errorf("not connected to database")
	}

	schema := &SchemaInfo{}

	// Get all tables
	tables, err := c.GetTables()
	if err != nil {
		return nil, fmt.Errorf("failed to get tables: %w", err)
	}

	// Get metadata for each table
	for _, tableName := range tables {
		metadata, err := c.GetTableMetadata(tableName)
		if err != nil {
			// Log error but continue with other tables
			continue
		}
		schema.Tables = append(schema.Tables, *metadata)
	}

	// Get all foreign key relationships
	relationships, err := c.getAllForeignKeyRelationships()
	if err != nil {
		return nil, fmt.Errorf("failed to get relationships: %w", err)
	}
	schema.Relationships = relationships

	return schema, nil
}

// getAllForeignKeyRelationships retrieves all FK relationships in the database
func (c *PostgresClient) getAllForeignKeyRelationships() ([]ForeignKeyRelationship, error) {
	query := `
		SELECT
			tc.table_name AS from_table,
			kcu.column_name AS from_column,
			ccu.table_name AS to_table,
			ccu.column_name AS to_column,
			tc.constraint_name,
			rc.delete_rule,
			rc.update_rule
		FROM information_schema.table_constraints AS tc
		JOIN information_schema.key_column_usage AS kcu
			ON tc.constraint_name = kcu.constraint_name
			AND tc.table_schema = kcu.table_schema
		JOIN information_schema.constraint_column_usage AS ccu
			ON ccu.constraint_name = tc.constraint_name
			AND ccu.table_schema = tc.table_schema
		JOIN information_schema.referential_constraints AS rc
			ON tc.constraint_name = rc.constraint_name
		WHERE tc.constraint_type = 'FOREIGN KEY'
			AND tc.table_schema = 'public'
		ORDER BY tc.table_name, tc.constraint_name
	`

	rows, err := c.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var relationships []ForeignKeyRelationship
	for rows.Next() {
		var rel ForeignKeyRelationship
		err := rows.Scan(
			&rel.FromTable,
			&rel.FromColumn,
			&rel.ToTable,
			&rel.ToColumn,
			&rel.Constraint,
			&rel.OnDelete,
			&rel.OnUpdate,
		)
		if err != nil {
			return nil, err
		}
		relationships = append(relationships, rel)
	}

	return relationships, rows.Err()
}

// GenerateERDiagram generates a text-based ER diagram
func GenerateERDiagram(schema *SchemaInfo) string {
	var sb strings.Builder

	sb.WriteString("Entity Relationship Diagram\n")
	sb.WriteString("===========================\n\n")

	// List all tables with their columns
	for _, table := range schema.Tables {
		sb.WriteString(fmt.Sprintf("┌─ %s ", table.Name))
		if len(table.PrimaryKeys) > 0 {
			sb.WriteString(fmt.Sprintf("(PK: %s)", strings.Join(table.PrimaryKeys, ", ")))
		}
		sb.WriteString("\n")

		for _, col := range table.Columns {
			marker := "  "
			if col.IsPrimaryKey {
				marker = "🔑"
			} else if col.IsForeignKey {
				marker = "🔗"
			}

			nullable := ""
			if !col.Nullable {
				nullable = " NOT NULL"
			}

			sb.WriteString(fmt.Sprintf("│  %s %s: %s%s\n", marker, col.Name, col.Type, nullable))
		}

		sb.WriteString("└─\n\n")
	}

	// Show relationships
	if len(schema.Relationships) > 0 {
		sb.WriteString("\nRelationships:\n")
		sb.WriteString("==============\n\n")

		for _, rel := range schema.Relationships {
			sb.WriteString(fmt.Sprintf("%s.%s ──> %s.%s (%s)\n",
				rel.FromTable,
				rel.FromColumn,
				rel.ToTable,
				rel.ToColumn,
				rel.Constraint,
			))

			if rel.OnDelete != "NO ACTION" || rel.OnUpdate != "NO ACTION" {
				sb.WriteString(fmt.Sprintf("    ON DELETE: %s, ON UPDATE: %s\n", rel.OnDelete, rel.OnUpdate))
			}
		}
	}

	return sb.String()
}

// FormatTableMetadata returns a human-readable table metadata summary
func FormatTableMetadata(metadata *TableMetadata) string {
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("Table: %s\n", metadata.Name))
	sb.WriteString(fmt.Sprintf("Schema: %s\n", metadata.Schema))
	sb.WriteString(fmt.Sprintf("Rows: %d\n", metadata.RowCount))
	sb.WriteString(fmt.Sprintf("Size: %s\n\n", metadata.TableSize))

	// Columns
	sb.WriteString("Columns:\n")
	sb.WriteString("--------\n")
	for _, col := range metadata.Columns {
		nullable := "NOT NULL"
		if col.Nullable {
			nullable = "NULL"
		}

		markers := []string{}
		if col.IsPrimaryKey {
			markers = append(markers, "PK")
		}
		if col.IsForeignKey {
			markers = append(markers, "FK")
		}
		if col.IsUnique {
			markers = append(markers, "UNIQUE")
		}

		markerStr := ""
		if len(markers) > 0 {
			markerStr = fmt.Sprintf(" [%s]", strings.Join(markers, ", "))
		}

		sb.WriteString(fmt.Sprintf("  %s: %s %s%s\n", col.Name, col.Type, nullable, markerStr))

		if col.DefaultValue != "" {
			sb.WriteString(fmt.Sprintf("    Default: %s\n", col.DefaultValue))
		}
	}

	// Primary Keys
	if len(metadata.PrimaryKeys) > 0 {
		sb.WriteString(fmt.Sprintf("\nPrimary Key: %s\n", strings.Join(metadata.PrimaryKeys, ", ")))
	}

	// Foreign Keys
	if len(metadata.ForeignKeys) > 0 {
		sb.WriteString("\nForeign Keys:\n")
		for _, fk := range metadata.ForeignKeys {
			sb.WriteString(fmt.Sprintf("  %s: %s -> %s.%s\n",
				fk.Name, fk.ColumnName, fk.ReferencedTable, fk.ReferencedColumn))
		}
	}

	// Indexes
	if len(metadata.Indexes) > 0 {
		sb.WriteString("\nIndexes:\n")
		for _, idx := range metadata.Indexes {
			idxType := idx.Type
			if idx.IsUnique {
				idxType += " UNIQUE"
			}
			if idx.IsPrimary {
				idxType += " PRIMARY"
			}
			sb.WriteString(fmt.Sprintf("  %s (%s): %s\n",
				idx.Name, idxType, strings.Join(idx.Columns, ", ")))
		}
	}

	return sb.String()
}
