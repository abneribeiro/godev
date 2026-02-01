package database

import (
	"fmt"
	"sort"
	"strings"
	"time"
)

// Migration represents a database migration
type Migration struct {
	ID          string
	Name        string
	Description string
	UpSQL       string
	DownSQL     string
	AppliedAt   time.Time
	Applied     bool
}

// MigrationHistory represents the migration history table
type MigrationHistory struct {
	ID        int
	Migration string
	AppliedAt time.Time
	Batch     int
}

// MigrationManager manages database migrations
type MigrationManager struct {
	client     *PostgresClient
	migrations []Migration
}

// NewMigrationManager creates a new migration manager
func NewMigrationManager(client *PostgresClient) *MigrationManager {
	return &MigrationManager{
		client:     client,
		migrations: []Migration{},
	}
}

// InitializeMigrationTable creates the migration tracking table
func (mm *MigrationManager) InitializeMigrationTable() error {
	createTableSQL := `
		CREATE TABLE IF NOT EXISTS godev_migrations (
			id SERIAL PRIMARY KEY,
			migration VARCHAR(255) NOT NULL UNIQUE,
			applied_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
			batch INTEGER NOT NULL DEFAULT 0
		);
	`

	result := mm.client.ExecuteQuery(createTableSQL)
	if result.Error != nil {
		return fmt.Errorf("failed to create migrations table: %w", result.Error)
	}

	return nil
}

// AddMigration registers a new migration
func (mm *MigrationManager) AddMigration(id, name, description, upSQL, downSQL string) {
	mm.migrations = append(mm.migrations, Migration{
		ID:          id,
		Name:        name,
		Description: description,
		UpSQL:       upSQL,
		DownSQL:     downSQL,
		Applied:     false,
	})
}

// GetPendingMigrations returns migrations that haven't been applied
func (mm *MigrationManager) GetPendingMigrations() ([]Migration, error) {
	appliedMigrations, err := mm.getAppliedMigrations()
	if err != nil {
		return nil, err
	}

	appliedMap := make(map[string]bool)
	for _, migration := range appliedMigrations {
		appliedMap[migration] = true
	}

	var pending []Migration
	for _, migration := range mm.migrations {
		if !appliedMap[migration.ID] {
			pending = append(pending, migration)
		}
	}

	// Sort by ID to ensure correct order
	sort.Slice(pending, func(i, j int) bool {
		return pending[i].ID < pending[j].ID
	})

	return pending, nil
}

// GetAppliedMigrations returns migrations that have been applied
func (mm *MigrationManager) GetAppliedMigrations() ([]Migration, error) {
	appliedIDs, err := mm.getAppliedMigrations()
	if err != nil {
		return nil, err
	}

	appliedMap := make(map[string]bool)
	for _, id := range appliedIDs {
		appliedMap[id] = true
	}

	var applied []Migration
	for _, migration := range mm.migrations {
		if appliedMap[migration.ID] {
			migration.Applied = true
			applied = append(applied, migration)
		}
	}

	return applied, nil
}

// getAppliedMigrations queries the database for applied migration IDs
func (mm *MigrationManager) getAppliedMigrations() ([]string, error) {
	query := "SELECT migration FROM godev_migrations ORDER BY id"
	result := mm.client.ExecuteQuery(query)

	if result.Error != nil {
		// If table doesn't exist, return empty list
		if strings.Contains(result.Error.Error(), "does not exist") {
			return []string{}, nil
		}
		return nil, result.Error
	}

	var migrations []string
	for _, row := range result.Rows {
		if len(row) > 0 {
			migrations = append(migrations, row[0])
		}
	}

	return migrations, nil
}

// Migrate runs all pending migrations
func (mm *MigrationManager) Migrate() error {
	// Ensure migration table exists
	if err := mm.InitializeMigrationTable(); err != nil {
		return err
	}

	pending, err := mm.GetPendingMigrations()
	if err != nil {
		return fmt.Errorf("failed to get pending migrations: %w", err)
	}

	if len(pending) == 0 {
		return nil
	}

	// Get current batch number
	batch, err := mm.getCurrentBatch()
	if err != nil {
		return err
	}
	batch++

	// Apply each migration
	for _, migration := range pending {
		if err := mm.applyMigration(migration, batch); err != nil {
			return fmt.Errorf("failed to apply migration %s: %w", migration.ID, err)
		}
	}

	return nil
}

// applyMigration applies a single migration
func (mm *MigrationManager) applyMigration(migration Migration, batch int) error {
	// Execute the UP SQL
	result := mm.client.ExecuteQuery(migration.UpSQL)
	if result.Error != nil {
		return fmt.Errorf("failed to execute UP SQL: %w", result.Error)
	}

	// Record the migration
	recordSQL := fmt.Sprintf(
		"INSERT INTO godev_migrations (migration, batch) VALUES ('%s', %d)",
		escapeSQLString(migration.ID),
		batch,
	)

	result = mm.client.ExecuteQuery(recordSQL)
	if result.Error != nil {
		return fmt.Errorf("failed to record migration: %w", result.Error)
	}

	return nil
}

// Rollback rolls back the last batch of migrations
func (mm *MigrationManager) Rollback() error {
	// Get the last batch number
	batch, err := mm.getCurrentBatch()
	if err != nil {
		return err
	}

	if batch == 0 {
		return fmt.Errorf("no migrations to rollback")
	}

	// Get migrations from the last batch
	query := fmt.Sprintf("SELECT migration FROM godev_migrations WHERE batch = %d ORDER BY id DESC", batch)
	result := mm.client.ExecuteQuery(query)

	if result.Error != nil {
		return fmt.Errorf("failed to get migrations to rollback: %w", result.Error)
	}

	// Rollback each migration
	for _, row := range result.Rows {
		if len(row) > 0 {
			migrationID := row[0]

			// Find the migration
			var migration *Migration
			for i := range mm.migrations {
				if mm.migrations[i].ID == migrationID {
					migration = &mm.migrations[i]
					break
				}
			}

			if migration == nil {
				return fmt.Errorf("migration not found: %s", migrationID)
			}

			// Execute DOWN SQL
			if migration.DownSQL != "" {
				downResult := mm.client.ExecuteQuery(migration.DownSQL)
				if downResult.Error != nil {
					return fmt.Errorf("failed to rollback migration %s: %w", migrationID, downResult.Error)
				}
			}

			// Remove from migration history
			deleteSQL := fmt.Sprintf(
				"DELETE FROM godev_migrations WHERE migration = '%s'",
				escapeSQLString(migrationID),
			)

			deleteResult := mm.client.ExecuteQuery(deleteSQL)
			if deleteResult.Error != nil {
				return fmt.Errorf("failed to delete migration record: %w", deleteResult.Error)
			}
		}
	}

	return nil
}

// getCurrentBatch gets the current batch number
func (mm *MigrationManager) getCurrentBatch() (int, error) {
	query := "SELECT COALESCE(MAX(batch), 0) FROM godev_migrations"
	result := mm.client.ExecuteQuery(query)

	if result.Error != nil {
		// If table doesn't exist, return 0
		if strings.Contains(result.Error.Error(), "does not exist") {
			return 0, nil
		}
		return 0, result.Error
	}

	if len(result.Rows) > 0 && len(result.Rows[0]) > 0 {
		var batch int
		fmt.Sscanf(result.Rows[0][0], "%d", &batch)
		return batch, nil
	}

	return 0, nil
}

// GetMigrationStatus returns the status of all migrations
func (mm *MigrationManager) GetMigrationStatus() (string, error) {
	appliedIDs, err := mm.getAppliedMigrations()
	if err != nil {
		return "", err
	}

	appliedMap := make(map[string]bool)
	for _, id := range appliedIDs {
		appliedMap[id] = true
	}

	var output strings.Builder
	output.WriteString("Migration Status\n")
	output.WriteString("================\n\n")

	for _, migration := range mm.migrations {
		status := "[ ]"
		if appliedMap[migration.ID] {
			status = "[✓]"
		}

		output.WriteString(fmt.Sprintf("%s %s - %s\n", status, migration.ID, migration.Name))
		if migration.Description != "" {
			output.WriteString(fmt.Sprintf("    %s\n", migration.Description))
		}
	}

	return output.String(), nil
}

// GenerateCreateTableMigration generates a CREATE TABLE migration
func GenerateCreateTableMigration(tableName string, columns []ColumnDefinition) Migration {
	timestamp := time.Now().Format("20060102_150405")
	id := fmt.Sprintf("%s_create_%s", timestamp, tableName)

	var columnDefs []string
	for _, col := range columns {
		def := fmt.Sprintf("%s %s", quoteIdentifierIfNeeded(col.Name), col.Type)

		if !col.Nullable {
			def += " NOT NULL"
		}

		if col.DefaultValue != "" {
			def += fmt.Sprintf(" DEFAULT %s", col.DefaultValue)
		}

		if col.IsPrimaryKey {
			def += " PRIMARY KEY"
		}

		if col.IsUnique {
			def += " UNIQUE"
		}

		columnDefs = append(columnDefs, def)
	}

	upSQL := fmt.Sprintf("CREATE TABLE %s (\n  %s\n);",
		quoteIdentifierIfNeeded(tableName),
		strings.Join(columnDefs, ",\n  "),
	)

	downSQL := fmt.Sprintf("DROP TABLE IF EXISTS %s;", quoteIdentifierIfNeeded(tableName))

	return Migration{
		ID:          id,
		Name:        fmt.Sprintf("Create %s table", tableName),
		Description: fmt.Sprintf("Creates the %s table", tableName),
		UpSQL:       upSQL,
		DownSQL:     downSQL,
	}
}

// ColumnDefinition defines a column for migration generation
type ColumnDefinition struct {
	Name         string
	Type         string
	Nullable     bool
	DefaultValue string
	IsPrimaryKey bool
	IsUnique     bool
}

// GenerateAddColumnMigration generates an ADD COLUMN migration
func GenerateAddColumnMigration(tableName string, column ColumnDefinition) Migration {
	timestamp := time.Now().Format("20060102_150405")
	id := fmt.Sprintf("%s_add_%s_to_%s", timestamp, column.Name, tableName)

	def := fmt.Sprintf("%s %s", quoteIdentifierIfNeeded(column.Name), column.Type)

	if !column.Nullable {
		def += " NOT NULL"
	}

	if column.DefaultValue != "" {
		def += fmt.Sprintf(" DEFAULT %s", column.DefaultValue)
	}

	upSQL := fmt.Sprintf("ALTER TABLE %s ADD COLUMN %s;",
		quoteIdentifierIfNeeded(tableName),
		def,
	)

	downSQL := fmt.Sprintf("ALTER TABLE %s DROP COLUMN %s;",
		quoteIdentifierIfNeeded(tableName),
		quoteIdentifierIfNeeded(column.Name),
	)

	return Migration{
		ID:          id,
		Name:        fmt.Sprintf("Add %s to %s", column.Name, tableName),
		Description: fmt.Sprintf("Adds %s column to %s table", column.Name, tableName),
		UpSQL:       upSQL,
		DownSQL:     downSQL,
	}
}

// GenerateAddIndexMigration generates a CREATE INDEX migration
func GenerateAddIndexMigration(tableName string, indexName string, columns []string, unique bool) Migration {
	timestamp := time.Now().Format("20060102_150405")
	id := fmt.Sprintf("%s_add_index_%s", timestamp, indexName)

	uniqueStr := ""
	if unique {
		uniqueStr = "UNIQUE "
	}

	var quotedColumns []string
	for _, col := range columns {
		quotedColumns = append(quotedColumns, quoteIdentifierIfNeeded(col))
	}

	upSQL := fmt.Sprintf("CREATE %sINDEX %s ON %s (%s);",
		uniqueStr,
		quoteIdentifierIfNeeded(indexName),
		quoteIdentifierIfNeeded(tableName),
		strings.Join(quotedColumns, ", "),
	)

	downSQL := fmt.Sprintf("DROP INDEX IF EXISTS %s;", quoteIdentifierIfNeeded(indexName))

	return Migration{
		ID:          id,
		Name:        fmt.Sprintf("Add index %s", indexName),
		Description: fmt.Sprintf("Creates index %s on %s", indexName, tableName),
		UpSQL:       upSQL,
		DownSQL:     downSQL,
	}
}
