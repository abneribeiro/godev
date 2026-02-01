package ui

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// SyntaxHighlighter provides syntax highlighting for various languages
type SyntaxHighlighter struct {
	Theme HighlightTheme
}

// HighlightTheme defines colors for syntax highlighting
type HighlightTheme struct {
	Keyword  lipgloss.Style
	String   lipgloss.Style
	Number   lipgloss.Style
	Comment  lipgloss.Style
	Function lipgloss.Style
	Operator lipgloss.Style
	Type     lipgloss.Style
	Variable lipgloss.Style
	Property lipgloss.Style
	Error    lipgloss.Style
}

// DefaultDarkTheme returns a default dark theme for syntax highlighting
func DefaultDarkTheme() HighlightTheme {
	return HighlightTheme{
		Keyword:  lipgloss.NewStyle().Foreground(lipgloss.Color("205")), // Pink
		String:   lipgloss.NewStyle().Foreground(lipgloss.Color("114")), // Green
		Number:   lipgloss.NewStyle().Foreground(lipgloss.Color("221")), // Yellow
		Comment:  lipgloss.NewStyle().Foreground(lipgloss.Color("243")), // Gray
		Function: lipgloss.NewStyle().Foreground(lipgloss.Color("111")), // Blue
		Operator: lipgloss.NewStyle().Foreground(lipgloss.Color("117")), // Cyan
		Type:     lipgloss.NewStyle().Foreground(lipgloss.Color("180")), // Orange
		Variable: lipgloss.NewStyle().Foreground(lipgloss.Color("222")), // Light yellow
		Property: lipgloss.NewStyle().Foreground(lipgloss.Color("146")), // Light green
		Error:    lipgloss.NewStyle().Foreground(lipgloss.Color("196")), // Red
	}
}

// DefaultLightTheme returns a default light theme for syntax highlighting
func DefaultLightTheme() HighlightTheme {
	return HighlightTheme{
		Keyword:  lipgloss.NewStyle().Foreground(lipgloss.Color("127")), // Purple
		String:   lipgloss.NewStyle().Foreground(lipgloss.Color("28")),  // Dark green
		Number:   lipgloss.NewStyle().Foreground(lipgloss.Color("166")), // Orange
		Comment:  lipgloss.NewStyle().Foreground(lipgloss.Color("245")), // Gray
		Function: lipgloss.NewStyle().Foreground(lipgloss.Color("21")),  // Blue
		Operator: lipgloss.NewStyle().Foreground(lipgloss.Color("31")),  // Cyan
		Type:     lipgloss.NewStyle().Foreground(lipgloss.Color("130")), // Brown
		Variable: lipgloss.NewStyle().Foreground(lipgloss.Color("130")), // Brown
		Property: lipgloss.NewStyle().Foreground(lipgloss.Color("24")),  // Dark blue
		Error:    lipgloss.NewStyle().Foreground(lipgloss.Color("160")), // Red
	}
}

// NewSyntaxHighlighter creates a new syntax highlighter with default dark theme
func NewSyntaxHighlighter() *SyntaxHighlighter {
	return &SyntaxHighlighter{
		Theme: DefaultDarkTheme(),
	}
}

// HighlightSQL highlights SQL syntax
func (sh *SyntaxHighlighter) HighlightSQL(sql string) string {
	result := sql

	// SQL Keywords
	sqlKeywords := []string{
		"SELECT", "FROM", "WHERE", "INSERT", "INTO", "UPDATE", "DELETE",
		"CREATE", "ALTER", "DROP", "TABLE", "INDEX", "VIEW", "DATABASE",
		"JOIN", "LEFT", "RIGHT", "INNER", "OUTER", "FULL", "CROSS",
		"ON", "AS", "AND", "OR", "NOT", "IN", "EXISTS", "BETWEEN",
		"LIKE", "IS", "NULL", "TRUE", "FALSE",
		"ORDER", "BY", "GROUP", "HAVING", "LIMIT", "OFFSET",
		"UNION", "INTERSECT", "EXCEPT",
		"PRIMARY", "KEY", "FOREIGN", "REFERENCES", "CONSTRAINT",
		"UNIQUE", "CHECK", "DEFAULT",
		"BEGIN", "COMMIT", "ROLLBACK", "TRANSACTION",
		"GRANT", "REVOKE", "WITH", "CASCADE",
	}

	// Highlight keywords (case-insensitive)
	for _, keyword := range sqlKeywords {
		pattern := regexp.MustCompile(fmt.Sprintf(`(?i)\b%s\b`, keyword))
		result = pattern.ReplaceAllStringFunc(result, func(match string) string {
			return sh.Theme.Keyword.Render(strings.ToUpper(match))
		})
	}

	// Highlight strings (single quotes)
	stringPattern := regexp.MustCompile(`'([^'\\]|\\.)*'`)
	result = stringPattern.ReplaceAllStringFunc(result, func(match string) string {
		return sh.Theme.String.Render(match)
	})

	// Highlight numbers
	numberPattern := regexp.MustCompile(`\b\d+(\.\d+)?\b`)
	result = numberPattern.ReplaceAllStringFunc(result, func(match string) string {
		return sh.Theme.Number.Render(match)
	})

	// Highlight comments
	commentPattern := regexp.MustCompile(`--[^\n]*`)
	result = commentPattern.ReplaceAllStringFunc(result, func(match string) string {
		return sh.Theme.Comment.Render(match)
	})

	// Highlight multi-line comments
	multiCommentPattern := regexp.MustCompile(`(?s)/\*.*?\*/`)
	result = multiCommentPattern.ReplaceAllStringFunc(result, func(match string) string {
		return sh.Theme.Comment.Render(match)
	})

	// Highlight data types
	sqlTypes := []string{
		"INTEGER", "INT", "BIGINT", "SMALLINT", "DECIMAL", "NUMERIC",
		"REAL", "DOUBLE", "FLOAT", "SERIAL", "BIGSERIAL",
		"VARCHAR", "CHAR", "TEXT", "BOOLEAN", "BOOL",
		"DATE", "TIME", "TIMESTAMP", "INTERVAL",
		"JSON", "JSONB", "UUID", "BYTEA", "ARRAY",
	}

	for _, sqlType := range sqlTypes {
		pattern := regexp.MustCompile(fmt.Sprintf(`(?i)\b%s\b`, sqlType))
		result = pattern.ReplaceAllStringFunc(result, func(match string) string {
			return sh.Theme.Type.Render(strings.ToUpper(match))
		})
	}

	// Highlight functions
	functionPattern := regexp.MustCompile(`\b([A-Za-z_][A-Za-z0-9_]*)\s*\(`)
	result = functionPattern.ReplaceAllStringFunc(result, func(match string) string {
		funcName := match[:len(match)-1] // Remove the (
		return sh.Theme.Function.Render(funcName) + "("
	})

	return result
}

// HighlightJSON highlights JSON syntax
func (sh *SyntaxHighlighter) HighlightJSON(json string) string {
	result := json

	// Highlight property names (keys)
	propertyPattern := regexp.MustCompile(`"([^"\\]|\\.)*"\s*:`)
	result = propertyPattern.ReplaceAllStringFunc(result, func(match string) string {
		colonPos := strings.LastIndex(match, ":")
		property := match[:colonPos]
		return sh.Theme.Property.Render(property) + ":"
	})

	// Highlight string values (but not keys which are already highlighted)
	stringPattern := regexp.MustCompile(`:\s*"([^"\\]|\\.)*"`)
	result = stringPattern.ReplaceAllStringFunc(result, func(match string) string {
		colonPos := strings.Index(match, ":")
		prefix := match[:colonPos+1]
		value := strings.TrimSpace(match[colonPos+1:])
		return prefix + " " + sh.Theme.String.Render(value)
	})

	// Highlight numbers
	numberPattern := regexp.MustCompile(`:\s*-?\d+(\.\d+)?([eE][+-]?\d+)?`)
	result = numberPattern.ReplaceAllStringFunc(result, func(match string) string {
		colonPos := strings.Index(match, ":")
		prefix := match[:colonPos+1]
		value := strings.TrimSpace(match[colonPos+1:])
		return prefix + " " + sh.Theme.Number.Render(value)
	})

	// Highlight booleans
	boolPattern := regexp.MustCompile(`:\s*(true|false)`)
	result = boolPattern.ReplaceAllStringFunc(result, func(match string) string {
		colonPos := strings.Index(match, ":")
		prefix := match[:colonPos+1]
		value := strings.TrimSpace(match[colonPos+1:])
		return prefix + " " + sh.Theme.Keyword.Render(value)
	})

	// Highlight null
	nullPattern := regexp.MustCompile(`:\s*null`)
	result = nullPattern.ReplaceAllStringFunc(result, func(match string) string {
		colonPos := strings.Index(match, ":")
		prefix := match[:colonPos+1]
		return prefix + " " + sh.Theme.Keyword.Render("null")
	})

	return result
}

// HighlightGraphQL highlights GraphQL syntax
func (sh *SyntaxHighlighter) HighlightGraphQL(gql string) string {
	result := gql

	// GraphQL Keywords
	gqlKeywords := []string{
		"query", "mutation", "subscription", "fragment", "on",
		"type", "interface", "enum", "union", "scalar", "input",
		"extend", "implements", "directive",
	}

	for _, keyword := range gqlKeywords {
		pattern := regexp.MustCompile(fmt.Sprintf(`\b%s\b`, keyword))
		result = pattern.ReplaceAllStringFunc(result, func(match string) string {
			return sh.Theme.Keyword.Render(match)
		})
	}

	// Highlight strings
	stringPattern := regexp.MustCompile(`"([^"\\]|\\.)*"`)
	result = stringPattern.ReplaceAllStringFunc(result, func(match string) string {
		return sh.Theme.String.Render(match)
	})

	// Highlight numbers
	numberPattern := regexp.MustCompile(`\b\d+(\.\d+)?\b`)
	result = numberPattern.ReplaceAllStringFunc(result, func(match string) string {
		return sh.Theme.Number.Render(match)
	})

	// Highlight comments
	commentPattern := regexp.MustCompile(`#[^\n]*`)
	result = commentPattern.ReplaceAllStringFunc(result, func(match string) string {
		return sh.Theme.Comment.Render(match)
	})

	// Highlight variables
	variablePattern := regexp.MustCompile(`\$[a-zA-Z_][a-zA-Z0-9_]*`)
	result = variablePattern.ReplaceAllStringFunc(result, func(match string) string {
		return sh.Theme.Variable.Render(match)
	})

	// Highlight types
	typePattern := regexp.MustCompile(`:\s*([A-Z][a-zA-Z0-9_]*)(!|\[|\s|$)`)
	result = typePattern.ReplaceAllStringFunc(result, func(match string) string {
		parts := strings.SplitN(match, ":", 2)
		if len(parts) == 2 {
			typePart := strings.TrimSpace(parts[1])
			// Find where the type name ends
			endIdx := 0
			for endIdx < len(typePart) && (typePart[endIdx] >= 'A' && typePart[endIdx] <= 'Z' ||
				typePart[endIdx] >= 'a' && typePart[endIdx] <= 'z' ||
				typePart[endIdx] >= '0' && typePart[endIdx] <= '9' ||
				typePart[endIdx] == '_') {
				endIdx++
			}

			if endIdx > 0 {
				typeName := typePart[:endIdx]
				rest := typePart[endIdx:]
				return parts[0] + ": " + sh.Theme.Type.Render(typeName) + rest
			}
		}
		return match
	})

	return result
}

// StripANSI removes ANSI color codes from a string
func StripANSI(s string) string {
	ansiPattern := regexp.MustCompile(`\x1b\[[0-9;]*m`)
	return ansiPattern.ReplaceAllString(s, "")
}

// HighlightError highlights error messages
func (sh *SyntaxHighlighter) HighlightError(errMsg string) string {
	// Highlight file paths
	pathPattern := regexp.MustCompile(`[a-zA-Z0-9_/\.-]+\.(go|sql|json|js|ts|py)`)
	result := pathPattern.ReplaceAllStringFunc(errMsg, func(match string) string {
		return sh.Theme.Property.Render(match)
	})

	// Highlight line numbers
	linePattern := regexp.MustCompile(`line\s+(\d+)`)
	result = linePattern.ReplaceAllStringFunc(result, func(match string) string {
		parts := strings.Split(match, " ")
		if len(parts) == 2 {
			return "line " + sh.Theme.Number.Render(parts[1])
		}
		return match
	})

	// Highlight error keywords
	errorKeywords := []string{"error", "ERROR", "Error", "failed", "FAILED", "Failed"}
	for _, keyword := range errorKeywords {
		result = strings.ReplaceAll(result, keyword, sh.Theme.Error.Render(keyword))
	}

	return result
}

// LineNumberedCode adds line numbers to code
func LineNumberedCode(code string, startLine int) string {
	lines := strings.Split(code, "\n")
	var result strings.Builder

	lineNumStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("243")).
		Width(4).
		Align(lipgloss.Right)

	for i, line := range lines {
		lineNum := startLine + i
		result.WriteString(lineNumStyle.Render(fmt.Sprintf("%d", lineNum)))
		result.WriteString(" │ ")
		result.WriteString(line)
		if i < len(lines)-1 {
			result.WriteString("\n")
		}
	}

	return result.String()
}

// HighlightDiff highlights diff output
func HighlightDiff(diff string) string {
	lines := strings.Split(diff, "\n")
	var result []string

	addedStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("114"))    // Green
	removedStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("196"))  // Red
	modifiedStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("221")) // Yellow
	headerStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("111"))   // Blue

	for _, line := range lines {
		if strings.HasPrefix(line, "+") {
			result = append(result, addedStyle.Render(line))
		} else if strings.HasPrefix(line, "-") {
			result = append(result, removedStyle.Render(line))
		} else if strings.HasPrefix(line, "~") {
			result = append(result, modifiedStyle.Render(line))
		} else if strings.HasPrefix(line, "===") || strings.HasPrefix(line, "---") {
			result = append(result, headerStyle.Render(line))
		} else {
			result = append(result, line)
		}
	}

	return strings.Join(result, "\n")
}
