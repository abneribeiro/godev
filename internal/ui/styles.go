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
)

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

func RenderFooter(shortcuts string) string {
	return FooterStyle.Render(shortcuts)
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


