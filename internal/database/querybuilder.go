package database

import (
	"fmt"
	"strings"
)

// QueryBuilder provides a fluent API for building SQL queries
type QueryBuilder struct {
	queryType       string // SELECT, INSERT, UPDATE, DELETE
	table           string
	columns         []string
	whereConditions []WhereCondition
	joins           []JoinClause
	orderBy         []OrderByClause
	groupBy         []string
	having          []WhereCondition
	limit           int
	offset          int
	values          map[string]interface{}
	updates         map[string]interface{}
}

// WhereCondition represents a WHERE clause condition
type WhereCondition struct {
	Column    string
	Operator  string // =, !=, <, >, <=, >=, LIKE, IN, IS NULL, IS NOT NULL
	Value     interface{}
	LogicalOp string // AND, OR
}

// JoinClause represents a JOIN operation
type JoinClause struct {
	Type    string // INNER, LEFT, RIGHT, FULL
	Table   string
	OnLeft  string // Left side of ON condition (this_table.column)
	OnRight string // Right side of ON condition (joined_table.column)
}

// OrderByClause represents an ORDER BY clause
type OrderByClause struct {
	Column    string
	Direction string // ASC, DESC
}

// NewQueryBuilder creates a new query builder
func NewQueryBuilder() *QueryBuilder {
	return &QueryBuilder{
		columns:         []string{},
		whereConditions: []WhereCondition{},
		joins:           []JoinClause{},
		orderBy:         []OrderByClause{},
		groupBy:         []string{},
		having:          []WhereCondition{},
		values:          make(map[string]interface{}),
		updates:         make(map[string]interface{}),
	}
}

// Select starts a SELECT query
func (qb *QueryBuilder) Select(columns ...string) *QueryBuilder {
	qb.queryType = "SELECT"
	qb.columns = columns
	return qb
}

// Insert starts an INSERT query
func (qb *QueryBuilder) Insert(table string) *QueryBuilder {
	qb.queryType = "INSERT"
	qb.table = table
	return qb
}

// Update starts an UPDATE query
func (qb *QueryBuilder) Update(table string) *QueryBuilder {
	qb.queryType = "UPDATE"
	qb.table = table
	return qb
}

// Delete starts a DELETE query
func (qb *QueryBuilder) Delete() *QueryBuilder {
	qb.queryType = "DELETE"
	return qb
}

// From specifies the table for SELECT or DELETE
func (qb *QueryBuilder) From(table string) *QueryBuilder {
	qb.table = table
	return qb
}

// Where adds a WHERE condition
func (qb *QueryBuilder) Where(column, operator string, value interface{}) *QueryBuilder {
	qb.whereConditions = append(qb.whereConditions, WhereCondition{
		Column:    column,
		Operator:  operator,
		Value:     value,
		LogicalOp: "AND",
	})
	return qb
}

// OrWhere adds an OR WHERE condition
func (qb *QueryBuilder) OrWhere(column, operator string, value interface{}) *QueryBuilder {
	qb.whereConditions = append(qb.whereConditions, WhereCondition{
		Column:    column,
		Operator:  operator,
		Value:     value,
		LogicalOp: "OR",
	})
	return qb
}

// WhereNull adds a WHERE column IS NULL condition
func (qb *QueryBuilder) WhereNull(column string) *QueryBuilder {
	qb.whereConditions = append(qb.whereConditions, WhereCondition{
		Column:    column,
		Operator:  "IS NULL",
		LogicalOp: "AND",
	})
	return qb
}

// WhereNotNull adds a WHERE column IS NOT NULL condition
func (qb *QueryBuilder) WhereNotNull(column string) *QueryBuilder {
	qb.whereConditions = append(qb.whereConditions, WhereCondition{
		Column:    column,
		Operator:  "IS NOT NULL",
		LogicalOp: "AND",
	})
	return qb
}

// WhereIn adds a WHERE column IN (...) condition
func (qb *QueryBuilder) WhereIn(column string, values []interface{}) *QueryBuilder {
	qb.whereConditions = append(qb.whereConditions, WhereCondition{
		Column:    column,
		Operator:  "IN",
		Value:     values,
		LogicalOp: "AND",
	})
	return qb
}

// Join adds an INNER JOIN
func (qb *QueryBuilder) Join(table, onLeft, onRight string) *QueryBuilder {
	qb.joins = append(qb.joins, JoinClause{
		Type:    "INNER",
		Table:   table,
		OnLeft:  onLeft,
		OnRight: onRight,
	})
	return qb
}

// LeftJoin adds a LEFT JOIN
func (qb *QueryBuilder) LeftJoin(table, onLeft, onRight string) *QueryBuilder {
	qb.joins = append(qb.joins, JoinClause{
		Type:    "LEFT",
		Table:   table,
		OnLeft:  onLeft,
		OnRight: onRight,
	})
	return qb
}

// RightJoin adds a RIGHT JOIN
func (qb *QueryBuilder) RightJoin(table, onLeft, onRight string) *QueryBuilder {
	qb.joins = append(qb.joins, JoinClause{
		Type:    "RIGHT",
		Table:   table,
		OnLeft:  onLeft,
		OnRight: onRight,
	})
	return qb
}

// OrderBy adds an ORDER BY clause
func (qb *QueryBuilder) OrderBy(column, direction string) *QueryBuilder {
	qb.orderBy = append(qb.orderBy, OrderByClause{
		Column:    column,
		Direction: strings.ToUpper(direction),
	})
	return qb
}

// GroupBy adds a GROUP BY clause
func (qb *QueryBuilder) GroupBy(columns ...string) *QueryBuilder {
	qb.groupBy = columns
	return qb
}

// Having adds a HAVING condition (used with GROUP BY)
func (qb *QueryBuilder) Having(column, operator string, value interface{}) *QueryBuilder {
	qb.having = append(qb.having, WhereCondition{
		Column:    column,
		Operator:  operator,
		Value:     value,
		LogicalOp: "AND",
	})
	return qb
}

// Limit adds a LIMIT clause
func (qb *QueryBuilder) Limit(limit int) *QueryBuilder {
	qb.limit = limit
	return qb
}

// Offset adds an OFFSET clause
func (qb *QueryBuilder) Offset(offset int) *QueryBuilder {
	qb.offset = offset
	return qb
}

// Values sets values for INSERT
func (qb *QueryBuilder) Values(values map[string]interface{}) *QueryBuilder {
	qb.values = values
	return qb
}

// Set sets values for UPDATE
func (qb *QueryBuilder) Set(column string, value interface{}) *QueryBuilder {
	qb.updates[column] = value
	return qb
}

// Build generates the SQL query string
func (qb *QueryBuilder) Build() (string, error) {
	switch qb.queryType {
	case "SELECT":
		return qb.buildSelect(), nil
	case "INSERT":
		return qb.buildInsert(), nil
	case "UPDATE":
		return qb.buildUpdate(), nil
	case "DELETE":
		return qb.buildDelete(), nil
	default:
		return "", fmt.Errorf("unknown query type: %s", qb.queryType)
	}
}

// buildSelect builds a SELECT query
func (qb *QueryBuilder) buildSelect() string {
	var query strings.Builder

	// SELECT clause
	query.WriteString("SELECT ")
	if len(qb.columns) == 0 {
		query.WriteString("*")
	} else {
		for i, col := range qb.columns {
			if i > 0 {
				query.WriteString(", ")
			}
			query.WriteString(quoteIdentifierIfNeeded(col))
		}
	}

	// FROM clause
	query.WriteString(fmt.Sprintf("\nFROM %s", quoteIdentifierIfNeeded(qb.table)))

	// JOIN clauses
	for _, join := range qb.joins {
		query.WriteString(fmt.Sprintf("\n%s JOIN %s ON %s = %s",
			join.Type,
			quoteIdentifierIfNeeded(join.Table),
			quoteIdentifierIfNeeded(join.OnLeft),
			quoteIdentifierIfNeeded(join.OnRight),
		))
	}

	// WHERE clause
	if len(qb.whereConditions) > 0 {
		query.WriteString("\nWHERE ")
		query.WriteString(buildWhereClause(qb.whereConditions))
	}

	// GROUP BY clause
	if len(qb.groupBy) > 0 {
		query.WriteString("\nGROUP BY ")
		for i, col := range qb.groupBy {
			if i > 0 {
				query.WriteString(", ")
			}
			query.WriteString(quoteIdentifierIfNeeded(col))
		}
	}

	// HAVING clause
	if len(qb.having) > 0 {
		query.WriteString("\nHAVING ")
		query.WriteString(buildWhereClause(qb.having))
	}

	// ORDER BY clause
	if len(qb.orderBy) > 0 {
		query.WriteString("\nORDER BY ")
		for i, order := range qb.orderBy {
			if i > 0 {
				query.WriteString(", ")
			}
			query.WriteString(fmt.Sprintf("%s %s",
				quoteIdentifierIfNeeded(order.Column),
				order.Direction,
			))
		}
	}

	// LIMIT clause
	if qb.limit > 0 {
		query.WriteString(fmt.Sprintf("\nLIMIT %d", qb.limit))
	}

	// OFFSET clause
	if qb.offset > 0 {
		query.WriteString(fmt.Sprintf("\nOFFSET %d", qb.offset))
	}

	query.WriteString(";")
	return query.String()
}

// buildInsert builds an INSERT query
func (qb *QueryBuilder) buildInsert() string {
	if len(qb.values) == 0 {
		return ""
	}

	var columns []string
	var values []string

	for col, val := range qb.values {
		columns = append(columns, quoteIdentifierIfNeeded(col))
		values = append(values, formatValueForSQL(val))
	}

	return fmt.Sprintf("INSERT INTO %s (%s)\nVALUES (%s);",
		quoteIdentifierIfNeeded(qb.table),
		strings.Join(columns, ", "),
		strings.Join(values, ", "),
	)
}

// buildUpdate builds an UPDATE query
func (qb *QueryBuilder) buildUpdate() string {
	if len(qb.updates) == 0 {
		return ""
	}

	var sets []string
	for col, val := range qb.updates {
		sets = append(sets, fmt.Sprintf("%s = %s",
			quoteIdentifierIfNeeded(col),
			formatValueForSQL(val),
		))
	}

	query := fmt.Sprintf("UPDATE %s\nSET %s",
		quoteIdentifierIfNeeded(qb.table),
		strings.Join(sets, ", "),
	)

	if len(qb.whereConditions) > 0 {
		query += "\nWHERE " + buildWhereClause(qb.whereConditions)
	}

	query += ";"
	return query
}

// buildDelete builds a DELETE query
func (qb *QueryBuilder) buildDelete() string {
	query := fmt.Sprintf("DELETE FROM %s", quoteIdentifierIfNeeded(qb.table))

	if len(qb.whereConditions) > 0 {
		query += "\nWHERE " + buildWhereClause(qb.whereConditions)
	}

	query += ";"
	return query
}

// buildWhereClause builds a WHERE/HAVING clause from conditions
func buildWhereClause(conditions []WhereCondition) string {
	var parts []string

	for i, cond := range conditions {
		var condStr string

		if cond.Operator == "IS NULL" || cond.Operator == "IS NOT NULL" {
			condStr = fmt.Sprintf("%s %s",
				quoteIdentifierIfNeeded(cond.Column),
				cond.Operator,
			)
		} else if cond.Operator == "IN" {
			values, ok := cond.Value.([]interface{})
			if ok {
				var valStrs []string
				for _, v := range values {
					valStrs = append(valStrs, formatValueForSQL(v))
				}
				condStr = fmt.Sprintf("%s IN (%s)",
					quoteIdentifierIfNeeded(cond.Column),
					strings.Join(valStrs, ", "),
				)
			}
		} else {
			condStr = fmt.Sprintf("%s %s %s",
				quoteIdentifierIfNeeded(cond.Column),
				cond.Operator,
				formatValueForSQL(cond.Value),
			)
		}

		if i > 0 {
			parts = append(parts, cond.LogicalOp+" "+condStr)
		} else {
			parts = append(parts, condStr)
		}
	}

	return strings.Join(parts, " ")
}

// quoteIdentifierIfNeeded quotes an identifier if it contains special characters or is a keyword
func quoteIdentifierIfNeeded(name string) string {
	// If already quoted or contains *, return as-is
	if strings.HasPrefix(name, "\"") || strings.Contains(name, "*") || strings.Contains(name, ".") {
		return name
	}

	// Check if it's a SQL keyword (simplified list)
	keywords := map[string]bool{
		"select": true, "from": true, "where": true, "join": true,
		"order": true, "group": true, "having": true, "limit": true,
		"user": true, "table": true, "column": true, "index": true,
	}

	lowerName := strings.ToLower(name)
	if keywords[lowerName] {
		return fmt.Sprintf("\"%s\"", name)
	}

	// Quote if contains spaces or special characters
	if strings.ContainsAny(name, " -+") {
		return fmt.Sprintf("\"%s\"", strings.ReplaceAll(name, "\"", "\"\""))
	}

	return name
}

// formatValueForSQL formats a value for SQL
func formatValueForSQL(value interface{}) string {
	if value == nil {
		return "NULL"
	}

	switch v := value.(type) {
	case string:
		// Escape single quotes
		escaped := strings.ReplaceAll(v, "'", "''")
		return fmt.Sprintf("'%s'", escaped)
	case int, int32, int64, float32, float64:
		return fmt.Sprintf("%v", v)
	case bool:
		if v {
			return "TRUE"
		}
		return "FALSE"
	default:
		return fmt.Sprintf("'%v'", v)
	}
}

// ToSQL is an alias for Build() for convenience
func (qb *QueryBuilder) ToSQL() (string, error) {
	return qb.Build()
}
