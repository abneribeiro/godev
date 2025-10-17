package ui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

func (m *Model) buildHeaderList() {
	m.headerList = []string{}
	for key := range m.headers {
		m.headerList = append(m.headerList, key)
	}
	m.selectedHeader = 0
	m.editingHeader = false
	m.headerKeyInput.SetValue("")
	m.headerValueInput.SetValue("")
}

func (m Model) handleHeaderEditorKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	if m.editingHeader {
		switch msg.String() {
		case "ctrl+c", "ctrl+q":
			return m, tea.Quit
		case "esc":
			m.editingHeader = false
			m.headerKeyInput.Blur()
			m.headerValueInput.Blur()
			m.headerKeyInput.SetValue("")
			m.headerValueInput.SetValue("")
			return m, nil
		case "tab":
			if m.headerKeyInput.Focused() {
				m.headerKeyInput.Blur()
				m.headerValueInput.Focus()
			} else {
				m.headerValueInput.Blur()
				m.headerKeyInput.Focus()
			}
			return m, nil
		case "enter":
			key := strings.TrimSpace(m.headerKeyInput.Value())
			value := strings.TrimSpace(m.headerValueInput.Value())
			if key != "" && value != "" {
				m.headers[key] = value
				m.buildHeaderList()
			}
			m.editingHeader = false
			return m, nil
		default:
			if m.headerKeyInput.Focused() {
				m.headerKeyInput, cmd = m.headerKeyInput.Update(msg)
			} else if m.headerValueInput.Focused() {
				m.headerValueInput, cmd = m.headerValueInput.Update(msg)
			}
			return m, cmd
		}
	}

	switch msg.String() {
	case "ctrl+c", "ctrl+q":
		return m, tea.Quit

	case "esc":
		m.state = StateRequestBuilder
		return m, nil

	case "up", "k":
		if m.selectedHeader > 0 {
			m.selectedHeader--
		}
		return m, nil

	case "down", "j":
		if m.selectedHeader < len(m.headerList)-1 {
			m.selectedHeader++
		}
		return m, nil

	case "n", "a":
		m.editingHeader = true
		m.headerKeyInput.Focus()
		m.headerKeyInput.SetValue("")
		m.headerValueInput.SetValue("")
		return m, nil

	case "d":
		if len(m.headerList) > 0 && m.selectedHeader < len(m.headerList) {
			key := m.headerList[m.selectedHeader]
			delete(m.headers, key)
			m.buildHeaderList()
			if m.selectedHeader >= len(m.headerList) && m.selectedHeader > 0 {
				m.selectedHeader--
			}
		}
		return m, nil

	case "e", "enter":
		if len(m.headerList) > 0 && m.selectedHeader < len(m.headerList) {
			key := m.headerList[m.selectedHeader]
			m.editingHeader = true
			m.headerKeyInput.Focus()
			m.headerKeyInput.SetValue(key)
			m.headerValueInput.SetValue(m.headers[key])
			delete(m.headers, key)
			m.buildHeaderList()
		}
		return m, nil
	}

	return m, nil
}

func (m Model) handleBodyEditorKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg.String() {
	case "ctrl+c", "ctrl+q":
		return m, tea.Quit

	case "esc":
		m.state = StateRequestBuilder
		m.bodyEditor.Blur()
		return m, nil

	case "ctrl+s":
		bodyValue := m.bodyEditor.Value()
		if err := m.validateJSON(bodyValue); err != nil {
			m.bodyError = err.Error()
			return m, nil
		}
		m.body = bodyValue
		m.bodyError = ""
		m.state = StateRequestBuilder
		m.bodyEditor.Blur()
		m.requestSaved = false
		return m, nil

	default:
		m.bodyEditor, cmd = m.bodyEditor.Update(msg)
		return m, cmd
	}
}

func (m *Model) buildQueryList() {
	m.queryList = []string{}
	for key := range m.queryParams {
		m.queryList = append(m.queryList, key)
	}
	m.selectedQuery = 0
	m.editingQuery = false
	m.queryKeyInput.SetValue("")
	m.queryValueInput.SetValue("")
}

func (m Model) handleQueryEditorKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	if m.editingQuery {
		switch msg.String() {
		case "ctrl+c", "ctrl+q":
			return m, tea.Quit
		case "esc":
			m.editingQuery = false
			m.queryKeyInput.Blur()
			m.queryValueInput.Blur()
			m.queryKeyInput.SetValue("")
			m.queryValueInput.SetValue("")
			return m, nil
		case "tab":
			if m.queryKeyInput.Focused() {
				m.queryKeyInput.Blur()
				m.queryValueInput.Focus()
			} else {
				m.queryValueInput.Blur()
				m.queryKeyInput.Focus()
			}
			return m, nil
		case "enter":
			key := strings.TrimSpace(m.queryKeyInput.Value())
			value := strings.TrimSpace(m.queryValueInput.Value())
			if key != "" && value != "" {
				m.queryParams[key] = value
				m.buildQueryList()
			}
			m.editingQuery = false
			return m, nil
		default:
			if m.queryKeyInput.Focused() {
				m.queryKeyInput, cmd = m.queryKeyInput.Update(msg)
			} else if m.queryValueInput.Focused() {
				m.queryValueInput, cmd = m.queryValueInput.Update(msg)
			}
			return m, cmd
		}
	}

	switch msg.String() {
	case "ctrl+c", "ctrl+q":
		return m, tea.Quit

	case "esc":
		m.state = StateRequestBuilder
		return m, nil

	case "up", "k":
		if m.selectedQuery > 0 {
			m.selectedQuery--
		}
		return m, nil

	case "down", "j":
		if m.selectedQuery < len(m.queryList)-1 {
			m.selectedQuery++
		}
		return m, nil

	case "n", "a":
		m.editingQuery = true
		m.queryKeyInput.Focus()
		m.queryKeyInput.SetValue("")
		m.queryValueInput.SetValue("")
		return m, nil

	case "d":
		if len(m.queryList) > 0 && m.selectedQuery < len(m.queryList) {
			key := m.queryList[m.selectedQuery]
			delete(m.queryParams, key)
			m.buildQueryList()
			if m.selectedQuery >= len(m.queryList) && m.selectedQuery > 0 {
				m.selectedQuery--
			}
		}
		return m, nil

	case "e", "enter":
		if len(m.queryList) > 0 && m.selectedQuery < len(m.queryList) {
			key := m.queryList[m.selectedQuery]
			m.editingQuery = true
			m.queryKeyInput.Focus()
			m.queryKeyInput.SetValue(key)
			m.queryValueInput.SetValue(m.queryParams[key])
			delete(m.queryParams, key)
			m.buildQueryList()
		}
		return m, nil
	}

	return m, nil
}

func (m Model) viewHeaderEditor() string {
	var b strings.Builder

	b.WriteString(TitleStyle.Render("Header Editor"))
	b.WriteString("\n\n")

	if m.editingHeader {
		b.WriteString(TextStyle.Render("Add/Edit Header"))
		b.WriteString("\n\n")

		keyLabel := "Key: "
		b.WriteString(TextStyle.Render(keyLabel))
		b.WriteString("\n")
		keyInput := m.headerKeyInput.View()
		if m.headerKeyInput.Focused() {
			styledInput := lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(lipgloss.Color(ColorAccent)).
				Padding(0, 1).
				Width(m.headerKeyInput.Width + 2).
				Render(keyInput)
			b.WriteString(styledInput)
		} else {
			styledInput := lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(lipgloss.Color(ColorBorder)).
				Padding(0, 1).
				Width(m.headerKeyInput.Width + 2).
				Render(keyInput)
			b.WriteString(styledInput)
		}
		b.WriteString("\n\n")

		valueLabel := "Value: "
		b.WriteString(TextStyle.Render(valueLabel))
		b.WriteString("\n")
		valueInput := m.headerValueInput.View()
		if m.headerValueInput.Focused() {
			styledInput := lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(lipgloss.Color(ColorAccent)).
				Padding(0, 1).
				Width(m.headerValueInput.Width + 2).
				Render(valueInput)
			b.WriteString(styledInput)
		} else {
			styledInput := lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(lipgloss.Color(ColorBorder)).
				Padding(0, 1).
				Width(m.headerValueInput.Width + 2).
				Render(valueInput)
			b.WriteString(styledInput)
		}
		b.WriteString("\n\n")

		buttons := RenderButton("Save (Enter)", true) + "  "
		buttons += RenderButton("Cancel (Esc)", false)
		b.WriteString(buttons)

		b.WriteString("\n\n")
		b.WriteString(RenderFooter("Tab: switch field • Enter: save • Esc: cancel"))
	} else {
		if len(m.headerList) == 0 {
			b.WriteString(MutedStyle.Render("No headers"))
			b.WriteString("\n\n")
			b.WriteString(TextStyle.Render("Press 'n' to add a new header"))
		} else {
			headerPanel := lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(lipgloss.Color(ColorBorder)).
				Padding(1, 2).
				Width(m.width - 10)

			var headerContent strings.Builder
			for i, key := range m.headerList {
				if i == m.selectedHeader {
					headerContent.WriteString(ListItemSelectedStyle.Render(fmt.Sprintf("> %-20s : %s", key, m.headers[key])))
				} else {
					headerContent.WriteString(ListItemStyle.Render(fmt.Sprintf("  %-20s : %s", key, m.headers[key])))
				}
				headerContent.WriteString("\n")
			}

			b.WriteString(headerPanel.Render(headerContent.String()))
		}

		b.WriteString("\n\n")

		buttons := RenderButton("Add (n)", false) + "  "
		buttons += RenderButton("Edit (e)", len(m.headerList) > 0) + "  "
		buttons += RenderButton("Delete (d)", len(m.headerList) > 0) + "  "
		buttons += RenderButton("Done (Esc)", false)
		b.WriteString(buttons)

		b.WriteString("\n\n")
		b.WriteString(RenderFooter("↑↓: navigate • n: add • e: edit • d: delete • Esc: back"))
	}

	return Center(m.width, m.height, b.String())
}

func (m Model) viewBodyEditor() string {
	var b strings.Builder

	b.WriteString(TitleStyle.Render("Body Editor (JSON)"))
	b.WriteString("\n\n")

	if m.bodyError != "" {
		b.WriteString(ErrorStyle.Render("✗ " + m.bodyError))
		b.WriteString("\n\n")
	}

	editorView := m.bodyEditor.View()
	borderColor := ColorAccent
	if m.bodyError != "" {
		borderColor = ColorError
	}
	styledEditor := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color(borderColor)).
		Padding(1, 2).
		Width(m.width - 10).
		Render(editorView)

	b.WriteString(styledEditor)
	b.WriteString("\n\n")

	buttons := RenderButton("Save (Ctrl+S)", true) + "  "
	buttons += RenderButton("Cancel (Esc)", false)
	b.WriteString(buttons)

	b.WriteString("\n\n")
	b.WriteString(RenderFooter("Ctrl+S: save & validate JSON • Esc: cancel"))

	return Center(m.width, m.height, b.String())
}

func (m Model) viewQueryEditor() string {
	var b strings.Builder

	b.WriteString(TitleStyle.Render("Query Parameters Editor"))
	b.WriteString("\n\n")

	if m.editingQuery {
		b.WriteString(TextStyle.Render("Add/Edit Query Parameter"))
		b.WriteString("\n\n")

		keyLabel := "Parameter Name: "
		b.WriteString(TextStyle.Render(keyLabel))
		b.WriteString("\n")
		keyInput := m.queryKeyInput.View()
		if m.queryKeyInput.Focused() {
			styledInput := lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(lipgloss.Color(ColorAccent)).
				Padding(0, 1).
				Width(m.queryKeyInput.Width + 2).
				Render(keyInput)
			b.WriteString(styledInput)
		} else {
			styledInput := lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(lipgloss.Color(ColorBorder)).
				Padding(0, 1).
				Width(m.queryKeyInput.Width + 2).
				Render(keyInput)
			b.WriteString(styledInput)
		}
		b.WriteString("\n\n")

		valueLabel := "Parameter Value: "
		b.WriteString(TextStyle.Render(valueLabel))
		b.WriteString("\n")
		valueInput := m.queryValueInput.View()
		if m.queryValueInput.Focused() {
			styledInput := lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(lipgloss.Color(ColorAccent)).
				Padding(0, 1).
				Width(m.queryValueInput.Width + 2).
				Render(valueInput)
			b.WriteString(styledInput)
		} else {
			styledInput := lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(lipgloss.Color(ColorBorder)).
				Padding(0, 1).
				Width(m.queryValueInput.Width + 2).
				Render(valueInput)
			b.WriteString(styledInput)
		}
		b.WriteString("\n\n")

		buttons := RenderButton("Save (Enter)", true) + "  "
		buttons += RenderButton("Cancel (Esc)", false)
		b.WriteString(buttons)

		b.WriteString("\n\n")
		b.WriteString(RenderFooter("Tab: switch field • Enter: save • Esc: cancel"))
	} else {
		if len(m.queryList) == 0 {
			b.WriteString(MutedStyle.Render("No query parameters"))
			b.WriteString("\n\n")
			b.WriteString(TextStyle.Render("Press 'n' to add a new query parameter"))
		} else {
			queryPanel := lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(lipgloss.Color(ColorBorder)).
				Padding(1, 2).
				Width(m.width - 10)

			var queryContent strings.Builder
			for i, key := range m.queryList {
				if i == m.selectedQuery {
					queryContent.WriteString(ListItemSelectedStyle.Render(fmt.Sprintf("> %-20s = %s", key, m.queryParams[key])))
				} else {
					queryContent.WriteString(ListItemStyle.Render(fmt.Sprintf("  %-20s = %s", key, m.queryParams[key])))
				}
				queryContent.WriteString("\n")
			}

			b.WriteString(queryPanel.Render(queryContent.String()))
		}

		b.WriteString("\n\n")

		buttons := RenderButton("Add (n)", false) + "  "
		buttons += RenderButton("Edit (e)", len(m.queryList) > 0) + "  "
		buttons += RenderButton("Delete (d)", len(m.queryList) > 0) + "  "
		buttons += RenderButton("Done (Esc)", false)
		b.WriteString(buttons)

		b.WriteString("\n\n")
		b.WriteString(RenderFooter("↑↓: navigate • n: add • e: edit • d: delete • Esc: back"))
	}

	return Center(m.width, m.height, b.String())
}
