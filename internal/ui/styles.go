package ui

import "github.com/charmbracelet/lipgloss"

const (
	ColorBg      = "#0D0D0D"
	ColorPanel   = "#1A1A1A"
	ColorBorder  = "#2D2D2D"
	ColorText    = "#E4E4E4"
	ColorMuted   = "#888888"
	ColorDim     = "#555555"
	ColorAccent  = "#FF8C00"
	ColorSuccess = "#00C853"
	ColorError   = "#D32F2F"
	ColorWarning = "#FFA726"
	Color2xx     = "#00C853"
	Color3xx     = "#FFA726"
	Color4xx     = "#FF5722"
	Color5xx     = "#D32F2F"

	// Responsive breakpoints
	BreakpointSmall  = 80   // Small terminal (80x24)
	BreakpointMedium = 120  // Medium terminal
	BreakpointLarge  = 160  // Large terminal

	// Minimum sizes for functionality
	MinTerminalWidth  = 60
	MinTerminalHeight = 10

	// Default sizes for various elements
	DefaultInputWidth = 60
	DefaultPanelWidth = 80
	MaxInputWidth     = 100
)

// LayoutConfig contains responsive layout configuration
type LayoutConfig struct {
	Width          int
	Height         int
	InputWidth     int
	PanelWidth     int
	ContentWidth   int
	ContentHeight  int
	Compact        bool
	StackVertical  bool
}

// NewLayoutConfig creates a responsive layout configuration
func NewLayoutConfig(width, height int) LayoutConfig {
	config := LayoutConfig{
		Width:  width,
		Height: height,
	}

	// Determine if we need compact mode
	config.Compact = width < BreakpointMedium || height < 20

	// Calculate input width responsively
	config.InputWidth = width - 20 // Leave margins
	if config.InputWidth < 40 {
		config.InputWidth = 40
	}
	if config.InputWidth > MaxInputWidth {
		config.InputWidth = MaxInputWidth
	}

	// Calculate panel width
	config.PanelWidth = width - 10 // Leave smaller margins for panels
	if config.PanelWidth < 50 {
		config.PanelWidth = 50
	}

	// Calculate content dimensions
	config.ContentWidth = config.PanelWidth - 6  // Account for panel padding
	config.ContentHeight = height - 8           // Account for title, headers, footers

	// Determine if we should stack elements vertically
	config.StackVertical = width < BreakpointLarge

	return config
}

// GetTableDimensions returns optimal dimensions for tables
func (lc LayoutConfig) GetTableDimensions() (width, height int) {
	tableWidth := lc.ContentWidth
	if tableWidth < 40 {
		tableWidth = 40
	}

	tableHeight := lc.ContentHeight - 4 // Reserve space for pagination info
	if tableHeight < 5 {
		tableHeight = 5
	}
	if tableHeight > 30 {
		tableHeight = 30 // Reasonable maximum
	}

	return tableWidth, tableHeight
}

// GetFormLayout returns form layout configuration
func (lc LayoutConfig) GetFormLayout() (fieldsPerRow int, fieldWidth int) {
	if lc.Compact || lc.Width < BreakpointMedium {
		// Compact mode: single column
		return 1, lc.InputWidth
	}

	if lc.Width >= BreakpointLarge {
		// Large screen: two columns
		return 2, (lc.ContentWidth - 5) / 2 // -5 for spacing
	}

	// Medium screen: single column but wider fields
	return 1, lc.InputWidth
}

// GetPaginationSize returns appropriate page size based on available height
func (lc LayoutConfig) GetPaginationSize() int {
	if lc.Compact {
		return 10
	}

	pageSize := lc.ContentHeight - 5 // Reserve space for headers and controls
	if pageSize < 5 {
		return 5
	}
	if pageSize > 50 {
		return 50
	}
	return pageSize
}

var (
	TitleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color(ColorAccent)).
			MarginBottom(1)

	TextStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color(ColorText))

	MutedStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color(ColorMuted))

	DimStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color(ColorDim))

	PanelStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color(ColorBorder)).
			Padding(1, 2).
			Background(lipgloss.Color(ColorPanel))

	ButtonActive = lipgloss.NewStyle().
			Foreground(lipgloss.Color(ColorBg)).
			Background(lipgloss.Color(ColorAccent)).
			Padding(0, 2).
			Bold(true)

	ButtonInactive = lipgloss.NewStyle().
			Foreground(lipgloss.Color(ColorText)).
			Padding(0, 2)

	InputStyle = lipgloss.NewStyle().
			Border(lipgloss.NormalBorder()).
			BorderForeground(lipgloss.Color(ColorBorder)).
			Padding(0, 1)

	InputFocused = lipgloss.NewStyle().
			Border(lipgloss.NormalBorder()).
			BorderForeground(lipgloss.Color(ColorAccent)).
			Padding(0, 1)

	StatusSuccessStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color(Color2xx)).
				Bold(true)

	StatusRedirectStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color(Color3xx)).
				Bold(true)

	StatusClientErrorStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color(Color4xx)).
				Bold(true)

	StatusServerErrorStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color(Color5xx)).
				Bold(true)

	ErrorStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color(ColorError)).
			Bold(true)

	SuccessStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color(ColorSuccess)).
			Bold(true)

	WarningStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color(ColorWarning)).
			Bold(true)

	FooterStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color(ColorMuted)).
			MarginTop(1)

	HeaderStyle = lipgloss.NewStyle().
			Background(lipgloss.Color(ColorPanel)).
			Foreground(lipgloss.Color(ColorAccent)).
			Padding(0, 1).
			Bold(true)

	ListItemStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color(ColorText)).
			PaddingLeft(2)

	ListItemSelectedStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color(ColorAccent)).
				PaddingLeft(0).
				Bold(true)

	SpinnerStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color(ColorAccent))

	CodeStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color(ColorText)).
			Background(lipgloss.Color(ColorBg))
)

// Responsive style functions
func GetResponsivePanelStyle(layout LayoutConfig) lipgloss.Style {
	style := PanelStyle.Width(layout.PanelWidth)

	if layout.Compact {
		// Reduce padding in compact mode
		style = style.Padding(0, 1)
	}

	return style
}

func GetResponsiveInputStyle(layout LayoutConfig, focused bool) lipgloss.Style {
	var style lipgloss.Style
	if focused {
		style = InputFocused
	} else {
		style = InputStyle
	}

	// Adjust width based on layout
	style = style.Width(layout.InputWidth)

	return style
}

func GetResponsiveTitleStyle(layout LayoutConfig) lipgloss.Style {
	style := TitleStyle

	if layout.Compact {
		// Reduce bottom margin in compact mode
		style = style.MarginBottom(0)
	}

	return style
}

func GetStatusStyle(statusCode int) lipgloss.Style {
	switch {
	case statusCode >= 200 && statusCode < 300:
		return StatusSuccessStyle
	case statusCode >= 300 && statusCode < 400:
		return StatusRedirectStyle
	case statusCode >= 400 && statusCode < 500:
		return StatusClientErrorStyle
	case statusCode >= 500:
		return StatusServerErrorStyle
	default:
		return TextStyle
	}
}

func RenderButton(text string, active bool) string {
	if active {
		return ButtonActive.Render("[ " + text + " ]")
	}
	return ButtonInactive.Render(text)
}

func RenderPanel(title, content string) string {
	panel := lipgloss.JoinVertical(
		lipgloss.Left,
		TitleStyle.Render(title),
		lipgloss.NewStyle().
			BorderStyle(lipgloss.NormalBorder()).
			BorderTop(true).
			BorderForeground(lipgloss.Color(ColorBorder)).
			Render(""),
		TextStyle.Render(content),
	)
	return PanelStyle.Render(panel)
}

func RenderResponsivePanel(title, content string, layout LayoutConfig) string {
	titleStyle := GetResponsiveTitleStyle(layout)
	panelStyle := GetResponsivePanelStyle(layout)

	panel := lipgloss.JoinVertical(
		lipgloss.Left,
		titleStyle.Render(title),
		lipgloss.NewStyle().
			BorderStyle(lipgloss.NormalBorder()).
			BorderTop(true).
			BorderForeground(lipgloss.Color(ColorBorder)).
			Render(""),
		TextStyle.Render(content),
	)
	return panelStyle.Render(panel)
}

func RenderFooter(shortcuts string) string {
	return FooterStyle.Render(shortcuts)
}

func RenderResponsiveFooter(shortcuts string, layout LayoutConfig) string {
	footer := FooterStyle.Render(shortcuts)

	// Wrap footer text if it's too long for the terminal
	if lipgloss.Width(footer) > layout.Width {
		// Split shortcuts on " • " and wrap
		parts := splitFooterText(shortcuts, layout.Width)
		footer = FooterStyle.Render(parts)
	}

	return footer
}

// splitFooterText splits footer text to fit within the given width
func splitFooterText(text string, maxWidth int) string {
	if len(text) <= maxWidth {
		return text
	}

	// Simple word wrapping for footer text
	parts := []string{}
	current := ""

	// Split on " • " first
	sections := []string{}
	for _, section := range []string{text} {
		if len(section) > 0 {
			sections = append(sections, section)
		}
	}

	for _, section := range sections {
		if len(current+" • "+section) <= maxWidth-10 { // -10 for margins
			if current != "" {
				current += " • "
			}
			current += section
		} else {
			if current != "" {
				parts = append(parts, current)
			}
			current = section
		}
	}

	if current != "" {
		parts = append(parts, current)
	}

	if len(parts) == 0 {
		return text // Fallback
	}

	// Join with newlines
	result := ""
	for i, part := range parts {
		if i > 0 {
			result += "\n"
		}
		result += part
	}

	return result
}

func CenterHorizontal(width int, content string) string {
	return lipgloss.Place(
		width,
		lipgloss.Height(content),
		lipgloss.Center,
		lipgloss.Center,
		content,
	)
}

func CenterVertical(height int, content string) string {
	return lipgloss.Place(
		lipgloss.Width(content),
		height,
		lipgloss.Center,
		lipgloss.Center,
		content,
	)
}

func Center(width, height int, content string) string {
	return lipgloss.Place(
		width,
		height,
		lipgloss.Center,
		lipgloss.Center,
		content,
	)
}

func CenterResponsive(layout LayoutConfig, content string) string {
	return lipgloss.Place(
		layout.Width,
		layout.Height,
		lipgloss.Center,
		lipgloss.Center,
		content,
	)
}
