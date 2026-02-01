package ui

import (
	"fmt"
	"strings"
	"unicode/utf8"
)

type TableRenderer struct {
	columns      []string
	rows         [][]string
	columnWidths []int
	maxWidth     int
}

func NewTableRenderer(columns []string, rows [][]string, maxWidth int) *TableRenderer {
	return &TableRenderer{
		columns:  columns,
		rows:     rows,
		maxWidth: maxWidth,
	}
}

func (t *TableRenderer) calculateColumnWidths() {
	if len(t.columns) == 0 {
		return
	}

	t.columnWidths = make([]int, len(t.columns))

	for i, col := range t.columns {
		t.columnWidths[i] = utf8.RuneCountInString(col)
	}

	for _, row := range t.rows {
		for i, cell := range row {
			if i < len(t.columnWidths) {
				cellLen := utf8.RuneCountInString(cell)
				if cellLen > t.columnWidths[i] {
					t.columnWidths[i] = cellLen
				}
			}
		}
	}

	const maxColWidth = 40
	const minColWidth = 8

	for i := range t.columnWidths {
		if t.columnWidths[i] > maxColWidth {
			t.columnWidths[i] = maxColWidth
		}
		if t.columnWidths[i] < minColWidth {
			t.columnWidths[i] = minColWidth
		}
	}
}

func (t *TableRenderer) truncate(s string, width int) string {
	runes := []rune(s)
	if len(runes) <= width {
		return s
	}
	if width <= 3 {
		return "..."
	}
	return string(runes[:width-3]) + "..."
}

func (t *TableRenderer) padRight(s string, width int) string {
	current := utf8.RuneCountInString(s)
	if current >= width {
		return s
	}
	return s + strings.Repeat(" ", width-current)
}

func (t *TableRenderer) padLeft(s string, width int) string {
	current := utf8.RuneCountInString(s)
	if current >= width {
		return s
	}
	return strings.Repeat(" ", width-current) + s
}

func (t *TableRenderer) isNumeric(s string) bool {
	if s == "" || s == "NULL" {
		return false
	}
	for _, r := range s {
		if r != '-' && r != '.' && r != ',' && (r < '0' || r > '9') {
			return false
		}
	}
	return true
}

func (t *TableRenderer) Render() string {
	if len(t.columns) == 0 {
		return ""
	}

	t.calculateColumnWidths()

	var result strings.Builder

	result.WriteString("┌")
	for i, width := range t.columnWidths {
		result.WriteString(strings.Repeat("─", width+2))
		if i < len(t.columnWidths)-1 {
			result.WriteString("┬")
		}
	}
	result.WriteString("┐\n")

	result.WriteString("│")
	for i, col := range t.columns {
		width := t.columnWidths[i]
		cell := t.truncate(col, width)
		cell = t.padRight(cell, width)
		result.WriteString(" " + HeaderStyle.Render(cell) + " ")
		if i < len(t.columns)-1 {
			result.WriteString("│")
		}
	}
	result.WriteString("│\n")

	result.WriteString("├")
	for i, width := range t.columnWidths {
		result.WriteString(strings.Repeat("─", width+2))
		if i < len(t.columnWidths)-1 {
			result.WriteString("┼")
		}
	}
	result.WriteString("┤\n")

	for _, row := range t.rows {
		result.WriteString("│")
		for i, cell := range row {
			if i >= len(t.columnWidths) {
				break
			}
			width := t.columnWidths[i]
			cellStr := cell
			if cellStr == "" {
				cellStr = "NULL"
			}

			cellStr = t.truncate(cellStr, width)

			if t.isNumeric(cell) {
				cellStr = t.padLeft(cellStr, width)
			} else {
				cellStr = t.padRight(cellStr, width)
			}

			result.WriteString(" " + cellStr + " ")
			if i < len(t.columnWidths)-1 {
				result.WriteString("│")
			}
		}
		result.WriteString("│\n")
	}

	result.WriteString("└")
	for i, width := range t.columnWidths {
		result.WriteString(strings.Repeat("─", width+2))
		if i < len(t.columnWidths)-1 {
			result.WriteString("┴")
		}
	}
	result.WriteString("┘")

	return result.String()
}

func (t *TableRenderer) RenderSummary(totalRows int, shownRows int) string {
	if totalRows == shownRows {
		return fmt.Sprintf("Showing all %d rows", totalRows)
	}
	return fmt.Sprintf("Showing %d of %d rows", shownRows, totalRows)
}
