package ui

import (
	"fmt"
	"strings"
	"unicode/utf8"

	"github.com/charmbracelet/bubbles/table"
	"github.com/charmbracelet/lipgloss"
)

const (
	// Default page size for table pagination
	DefaultTablePageSize = 20
	// Maximum column width to prevent layout issues
	MaxColumnWidth = 40
	// Minimum column width for readability
	MinColumnWidth = 8
)

// BubblesTableWrapper wraps the Bubbles table component with additional functionality
type BubblesTableWrapper struct {
	table        table.Model
	allRows      []table.Row
	currentPage  int
	pageSize     int
	totalPages   int
	width        int
	height       int
}

// NewBubblesTableWrapper creates a new table wrapper with pagination support
func NewBubblesTableWrapper(columns []string, rows [][]string, width, height int) *BubblesTableWrapper {
	// Convert columns to Bubbles table columns with calculated widths
	tableCols := calculateTableColumns(columns, rows, width)

	// Convert rows to Bubbles table rows
	tableRows := make([]table.Row, len(rows))
	for i, row := range rows {
		// Ensure all rows have the same number of columns
		tableRow := make(table.Row, len(columns))
		for j := 0; j < len(columns); j++ {
			if j < len(row) {
				// Handle empty cells
				if row[j] == "" {
					tableRow[j] = "NULL"
				} else {
					tableRow[j] = row[j]
				}
			} else {
				tableRow[j] = "NULL"
			}
		}
		tableRows[i] = tableRow
	}

	// Calculate optimal pagination size based on height
	pageSize := DefaultTablePageSize
	if height > 10 {
		// Adjust page size based on available height (reserve space for headers, borders, etc.)
		pageSize = height - 6 // More conservative space allocation
		if pageSize < 5 {
			pageSize = 5 // Minimum reasonable page size
		}
		if pageSize > 100 { // Increase maximum for large displays
			pageSize = 100
		}
	}

	// Handle large datasets more efficiently
	totalRows := len(tableRows)
	totalPages := (totalRows + pageSize - 1) / pageSize
	if totalPages == 0 {
		totalPages = 1
	}

	// Get initial page of rows
	displayRows := getPageRows(tableRows, 0, pageSize)

	// Calculate responsive table height
	tableHeight := min(len(displayRows)+2, height-4)
	if tableHeight < 3 {
		tableHeight = 3 // Minimum for header + at least one row
	}

	// Create Bubbles table with custom styles
	t := table.New(
		table.WithColumns(tableCols),
		table.WithRows(displayRows),
		table.WithFocused(false),
		table.WithHeight(tableHeight),
		table.WithStyles(getTableStyles()),
	)

	return &BubblesTableWrapper{
		table:       t,
		allRows:     tableRows,
		currentPage: 0,
		pageSize:    pageSize,
		totalPages:  totalPages,
		width:       width,
		height:      height,
	}
}

// calculateTableColumns creates table columns with appropriate widths
func calculateTableColumns(columns []string, rows [][]string, maxWidth int) []table.Column {
	if len(columns) == 0 {
		return []table.Column{}
	}

	// Calculate initial column widths based on content
	columnWidths := make([]int, len(columns))

	// Start with header lengths
	for i, col := range columns {
		columnWidths[i] = utf8.RuneCountInString(col)
	}

	// Check all row data
	for _, row := range rows {
		for i, cell := range row {
			if i < len(columnWidths) {
				cellLen := utf8.RuneCountInString(cell)
				if cellLen > columnWidths[i] {
					columnWidths[i] = cellLen
				}
			}
		}
	}

	// Apply width constraints
	for i := range columnWidths {
		if columnWidths[i] > MaxColumnWidth {
			columnWidths[i] = MaxColumnWidth
		}
		if columnWidths[i] < MinColumnWidth {
			columnWidths[i] = MinColumnWidth
		}
	}

	// Adjust widths to fit within available space
	totalWidth := 0
	for _, width := range columnWidths {
		totalWidth += width + 3 // +3 for padding and borders
	}

	// If table is too wide, proportionally reduce column widths
	if totalWidth > maxWidth-10 { // -10 for margins
		availableWidth := maxWidth - 10 - (len(columns) * 3)
		if availableWidth > 0 {
			ratio := float64(availableWidth) / float64(totalWidth-(len(columns)*3))
			for i := range columnWidths {
				newWidth := int(float64(columnWidths[i]) * ratio)
				if newWidth < MinColumnWidth {
					newWidth = MinColumnWidth
				}
				columnWidths[i] = newWidth
			}
		}
	}

	// Create Bubbles table columns
	tableCols := make([]table.Column, len(columns))
	for i, col := range columns {
		tableCols[i] = table.Column{
			Title: col,
			Width: columnWidths[i],
		}
	}

	return tableCols
}

// getPageRows returns a page of rows for pagination
func getPageRows(allRows []table.Row, page, pageSize int) []table.Row {
	start := page * pageSize
	if start >= len(allRows) {
		return []table.Row{}
	}

	end := start + pageSize
	if end > len(allRows) {
		end = len(allRows)
	}

	return allRows[start:end]
}

// getTableStyles returns custom styles for the table
func getTableStyles() table.Styles {
	s := table.DefaultStyles()
	s.Header = s.Header.
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color(ColorBorder)).
		BorderBottom(true).
		Bold(true).
		Foreground(lipgloss.Color(ColorAccent)).
		Background(lipgloss.Color(ColorPanel))

	s.Selected = s.Selected.
		Foreground(lipgloss.Color(ColorBg)).
		Background(lipgloss.Color(ColorAccent)).
		Bold(true)

	s.Cell = s.Cell.
		Foreground(lipgloss.Color(ColorText))

	return s
}

// NextPage moves to the next page
func (btw *BubblesTableWrapper) NextPage() {
	if btw.currentPage < btw.totalPages-1 {
		btw.currentPage++
		btw.updateDisplayRows()
	}
}

// PrevPage moves to the previous page
func (btw *BubblesTableWrapper) PrevPage() {
	if btw.currentPage > 0 {
		btw.currentPage--
		btw.updateDisplayRows()
	}
}

// updateDisplayRows updates the table with rows for the current page
func (btw *BubblesTableWrapper) updateDisplayRows() {
	displayRows := getPageRows(btw.allRows, btw.currentPage, btw.pageSize)
	btw.table.SetRows(displayRows)

	// Update table height based on number of rows
	newHeight := min(len(displayRows)+2, btw.height-4)
	btw.table.SetHeight(newHeight)
}

// Render returns the rendered table
func (btw *BubblesTableWrapper) Render() string {
	return btw.table.View()
}

// RenderSummary returns pagination and summary information
func (btw *BubblesTableWrapper) RenderSummary() string {
	totalRows := len(btw.allRows)
	if totalRows == 0 {
		return "No rows"
	}

	if btw.totalPages <= 1 {
		return fmt.Sprintf("Showing all %d rows", totalRows)
	}

	startRow := btw.currentPage*btw.pageSize + 1
	endRow := min((btw.currentPage+1)*btw.pageSize, totalRows)

	return fmt.Sprintf("Showing rows %d-%d of %d (Page %d of %d)",
		startRow, endRow, totalRows, btw.currentPage+1, btw.totalPages)
}

// RenderPaginationFooter returns pagination controls
func (btw *BubblesTableWrapper) RenderPaginationFooter() string {
	if btw.totalPages <= 1 {
		return ""
	}

	var controls []string

	if btw.currentPage > 0 {
		controls = append(controls, "← prev page")
	}

	if btw.currentPage < btw.totalPages-1 {
		controls = append(controls, "→ next page")
	}

	if len(controls) == 0 {
		return ""
	}

	return strings.Join(controls, " • ")
}

// CanPageUp returns true if previous page is available
func (btw *BubblesTableWrapper) CanPageUp() bool {
	return btw.currentPage > 0
}

// CanPageDown returns true if next page is available
func (btw *BubblesTableWrapper) CanPageDown() bool {
	return btw.currentPage < btw.totalPages-1
}

// GetCurrentPage returns the current page number (0-based)
func (btw *BubblesTableWrapper) GetCurrentPage() int {
	return btw.currentPage
}

// GetTotalPages returns the total number of pages
func (btw *BubblesTableWrapper) GetTotalPages() int {
	return btw.totalPages
}

// GetTotalRows returns the total number of rows
func (btw *BubblesTableWrapper) GetTotalRows() int {
	return len(btw.allRows)
}

// GetPageSize returns the current page size
func (btw *BubblesTableWrapper) GetPageSize() int {
	return btw.pageSize
}

// SetPageSize changes the page size and recalculates pagination
func (btw *BubblesTableWrapper) SetPageSize(newSize int) {
	if newSize < 1 {
		newSize = 1
	}
	if newSize > 200 { // Reasonable maximum
		newSize = 200
	}

	btw.pageSize = newSize
	btw.totalPages = (len(btw.allRows) + newSize - 1) / newSize
	if btw.totalPages == 0 {
		btw.totalPages = 1
	}

	// Adjust current page if needed
	if btw.currentPage >= btw.totalPages {
		btw.currentPage = btw.totalPages - 1
		if btw.currentPage < 0 {
			btw.currentPage = 0
		}
	}

	btw.updateDisplayRows()
}

// JumpToPage jumps to a specific page
func (btw *BubblesTableWrapper) JumpToPage(pageNum int) {
	if pageNum < 0 {
		pageNum = 0
	}
	if pageNum >= btw.totalPages {
		pageNum = btw.totalPages - 1
	}

	if pageNum != btw.currentPage {
		btw.currentPage = pageNum
		btw.updateDisplayRows()
	}
}

// FirstPage jumps to the first page
func (btw *BubblesTableWrapper) FirstPage() {
	btw.JumpToPage(0)
}

// LastPage jumps to the last page
func (btw *BubblesTableWrapper) LastPage() {
	btw.JumpToPage(btw.totalPages - 1)
}

// GetCurrentPageInfo returns detailed information about the current page
func (btw *BubblesTableWrapper) GetCurrentPageInfo() (currentPage, totalPages, startRow, endRow, totalRows int) {
	totalRows = len(btw.allRows)
	if totalRows == 0 {
		return 0, 1, 0, 0, 0
	}

	currentPage = btw.currentPage + 1 // 1-based for display
	totalPages = btw.totalPages
	startRow = btw.currentPage*btw.pageSize + 1
	endRow = min((btw.currentPage+1)*btw.pageSize, totalRows)

	return currentPage, totalPages, startRow, endRow, totalRows
}

// Update passes through updates to the underlying table
func (btw *BubblesTableWrapper) Update(msg interface{}) {
	// For now, we don't need to handle updates since we're controlling pagination
	// This could be expanded later for features like row selection
}

// GetPerformanceStats returns performance-related statistics
func (btw *BubblesTableWrapper) GetPerformanceStats() string {
	currentPage, totalPages, startRow, endRow, totalRows := btw.GetCurrentPageInfo()

	if totalRows == 0 {
		return "No data"
	}

	if totalPages == 1 {
		return fmt.Sprintf("%d rows", totalRows)
	}

	return fmt.Sprintf("Page %d/%d • Rows %d-%d of %d • Page size: %d",
		currentPage, totalPages, startRow, endRow, totalRows, btw.pageSize)
}

// IsLargeDataset returns true if the dataset is considered large
func (btw *BubblesTableWrapper) IsLargeDataset() bool {
	return len(btw.allRows) > 1000
}

// GetMemoryEstimate returns a rough estimate of memory usage in KB
func (btw *BubblesTableWrapper) GetMemoryEstimate() int {
	if len(btw.allRows) == 0 {
		return 0
	}

	// Rough estimate: each cell averages 20 characters, 1 byte per char
	avgCellSize := 20
	totalCells := len(btw.allRows) * len(btw.allRows[0])
	estimatedBytes := totalCells * avgCellSize

	return estimatedBytes / 1024 // Convert to KB
}

// Helper function for min
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// Legacy TableRenderer for backward compatibility during transition
// This will be removed once all uses are migrated to BubblesTableWrapper
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

	for i := range t.columnWidths {
		if t.columnWidths[i] > MaxColumnWidth {
			t.columnWidths[i] = MaxColumnWidth
		}
		if t.columnWidths[i] < MinColumnWidth {
			t.columnWidths[i] = MinColumnWidth
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
	// For backward compatibility, use the old rendering logic for tests
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
