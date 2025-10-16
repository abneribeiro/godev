package ui

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	httpclient "github.com/abneribeiro/devscope/internal/http"
	"github.com/abneribeiro/devscope/internal/storage"
)

type AppState int

const (
	StateRequestBuilder AppState = iota
	StateLoading
	StateViewResponse
	StateRequestList
	StateHelp
)

type Model struct {
	state   AppState
	width   int
	height  int
	storage *storage.Storage

	method      string
	urlInput    textinput.Model
	headers     map[string]string
	body        string
	focusIndex  int
	cursorPos   int

	httpClient *httpclient.Client
	response   *httpclient.Response
	spinner    spinner.Model
	loading    bool

	savedRequests   []storage.SavedRequest
	selectedReqIdx  int
	scrollOffset    int

	err error
}

type responseMsg httpclient.Response

func NewModel() *Model {
	ti := textinput.New()
	ti.Placeholder = "https://api.example.com/endpoint"
	ti.Focus()
	ti.CharLimit = 2000
	ti.Width = 60

	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = SpinnerStyle

	store, err := storage.NewStorage()

	m := &Model{
		state:      StateRequestBuilder,
		method:     "GET",
		urlInput:   ti,
		headers:    make(map[string]string),
		body:       "",
		focusIndex: 1,
		httpClient: httpclient.NewClient(30 * time.Second),
		spinner:    s,
		storage:    store,
		err:        err,
	}

	if m.storage != nil {
		m.savedRequests = m.storage.GetRequests()
	}

	return m
}

func (m Model) Init() tea.Cmd {
	return textinput.Blink
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		if m.state == StateRequestBuilder && m.focusIndex == 1 {
			switch msg.String() {
			case "ctrl+q", "tab", "shift+tab", "enter", "ctrl+l", "ctrl+?":
				return m.handleKeyPress(msg)
			case "ctrl+c":
				if m.urlInput.Value() != "" {
					m.urlInput, cmd = m.urlInput.Update(msg)
					return m, cmd
				}
				return m.handleKeyPress(msg)
			default:
				m.urlInput, cmd = m.urlInput.Update(msg)
				return m, cmd
			}
		}
		return m.handleKeyPress(msg)

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		inputWidth := m.width - 20
		if inputWidth < 40 {
			inputWidth = 40
		}
		if inputWidth > 100 {
			inputWidth = 100
		}
		m.urlInput.Width = inputWidth
		return m, nil

	case responseMsg:
		m.loading = false
		resp := httpclient.Response(msg)
		m.response = &resp
		m.state = StateViewResponse
		return m, nil

	case spinner.TickMsg:
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd
	}

	return m, cmd
}

func (m Model) handleKeyPress(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch m.state {
	case StateRequestBuilder:
		return m.handleRequestBuilderKeys(msg)
	case StateViewResponse:
		return m.handleResponseViewKeys(msg)
	case StateRequestList:
		return m.handleRequestListKeys(msg)
	case StateHelp:
		return m.handleHelpKeys(msg)
	case StateLoading:
		if msg.String() == "ctrl+c" {
			return m, tea.Quit
		}
		return m, nil
	}
	return m, nil
}

func (m Model) handleRequestBuilderKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "ctrl+c", "ctrl+q":
		return m, tea.Quit

	case "ctrl+?":
		m.state = StateHelp
		return m, nil

	case "tab":
		m.focusIndex++
		if m.focusIndex > 4 {
			m.focusIndex = 0
		}

		if m.focusIndex == 1 {
			m.urlInput.Focus()
		} else {
			m.urlInput.Blur()
		}
		return m, nil

	case "shift+tab":
		m.focusIndex--
		if m.focusIndex < 0 {
			m.focusIndex = 4
		}

		if m.focusIndex == 1 {
			m.urlInput.Focus()
		} else {
			m.urlInput.Blur()
		}
		return m, nil

	case "left", "h":
		if m.focusIndex == 0 {
			methods := []string{"GET", "POST", "PUT", "DELETE", "PATCH"}
			for i, method := range methods {
				if m.method == method {
					if i > 0 {
						m.method = methods[i-1]
					}
					break
				}
			}
		}
		return m, nil

	case "right", "l":
		if m.focusIndex == 0 {
			methods := []string{"GET", "POST", "PUT", "DELETE", "PATCH"}
			for i, method := range methods {
				if m.method == method {
					if i < len(methods)-1 {
						m.method = methods[i+1]
					}
					break
				}
			}
		}
		return m, nil

	case "enter":
		switch m.focusIndex {
		case 2:
			m.state = StateRequestList
			return m, nil
		case 3:
			if m.urlInput.Value() != "" {
				return m, m.sendRequest()
			}
		case 4:
			return m, tea.Quit
		}
		return m, nil

	case "ctrl+l":
		m.state = StateRequestList
		return m, nil
	}

	return m, nil
}

func (m Model) handleResponseViewKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "ctrl+c", "ctrl+q":
		return m, tea.Quit

	case "esc":
		m.state = StateRequestBuilder
		m.response = nil
		return m, nil

	case "s":
		if m.storage != nil && m.response != nil {
			name := fmt.Sprintf("%s %s", m.method, m.urlInput.Value())
			if !m.storage.RequestExists(name) {
				err := m.storage.SaveRequest(name, m.method, m.urlInput.Value(), m.headers, m.body)
				if err == nil {
					m.savedRequests = m.storage.GetRequests()
				}
			}
		}
		return m, nil

	case "up", "k":
		if m.scrollOffset > 0 {
			m.scrollOffset--
		}
		return m, nil

	case "down", "j":
		m.scrollOffset++
		return m, nil
	}

	return m, nil
}

func (m Model) handleRequestListKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "ctrl+c", "ctrl+q":
		return m, tea.Quit

	case "esc":
		m.state = StateRequestBuilder
		return m, nil

	case "up", "k":
		if m.selectedReqIdx > 0 {
			m.selectedReqIdx--
		}
		return m, nil

	case "down", "j":
		if m.selectedReqIdx < len(m.savedRequests)-1 {
			m.selectedReqIdx++
		}
		return m, nil

	case "enter":
		if len(m.savedRequests) > 0 && m.selectedReqIdx < len(m.savedRequests) {
			req := m.savedRequests[m.selectedReqIdx]
			m.method = req.Method
			m.urlInput.SetValue(req.URL)
			m.headers = req.Headers
			m.body = req.Body
			m.state = StateRequestBuilder

			if m.storage != nil {
				m.storage.UpdateLastUsed(req.ID)
			}
		}
		return m, nil

	case "d":
		if len(m.savedRequests) > 0 && m.selectedReqIdx < len(m.savedRequests) && m.storage != nil {
			req := m.savedRequests[m.selectedReqIdx]
			m.storage.DeleteRequest(req.ID)
			m.savedRequests = m.storage.GetRequests()
			if m.selectedReqIdx >= len(m.savedRequests) && m.selectedReqIdx > 0 {
				m.selectedReqIdx--
			}
		}
		return m, nil

	case "n":
		m.method = "GET"
		m.urlInput.SetValue("")
		m.headers = make(map[string]string)
		m.body = ""
		m.state = StateRequestBuilder
		return m, nil
	}

	return m, nil
}

func (m Model) handleHelpKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	m.state = StateRequestBuilder
	return m, nil
}

func (m Model) sendRequest() tea.Cmd {
	m.state = StateLoading
	m.loading = true
	m.scrollOffset = 0

	return tea.Batch(
		m.spinner.Tick,
		func() tea.Msg {
			req := httpclient.Request{
				Method:  m.method,
				URL:     m.urlInput.Value(),
				Headers: m.headers,
				Body:    m.body,
			}
			resp := m.httpClient.Send(req)
			return responseMsg(resp)
		},
	)
}

func (m Model) View() string {
	if m.err != nil {
		return ErrorStyle.Render(fmt.Sprintf("Error: %v\nPress Ctrl+Q to quit", m.err))
	}

	switch m.state {
	case StateRequestBuilder:
		return m.viewRequestBuilder()
	case StateLoading:
		return m.viewLoading()
	case StateViewResponse:
		return m.viewResponse()
	case StateRequestList:
		return m.viewRequestList()
	case StateHelp:
		return m.viewHelp()
	}

	return ""
}

func (m Model) viewRequestBuilder() string {
	var b strings.Builder

	b.WriteString(TitleStyle.Render("DevScope v0.1.0"))
	b.WriteString("\n\n")

	methodLabel := "Method: "
	methodSection := methodLabel
	if m.focusIndex == 0 {
		methodSection = TextStyle.Render(methodLabel) + ButtonActive.Render("[ " + m.method + " ▾ ]")
	} else {
		methodSection = MutedStyle.Render(methodLabel) + TextStyle.Render(m.method + " ▾")
	}
	b.WriteString(methodSection)
	b.WriteString("\n\n")

	urlLabel := "URL: "
	b.WriteString(TextStyle.Render(urlLabel))
	b.WriteString("\n")

	if m.focusIndex == 1 {
		inputView := m.urlInput.View()
		styledInput := lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color(ColorAccent)).
			Padding(0, 1).
			Width(m.urlInput.Width + 2).
			Render(inputView)
		b.WriteString(styledInput)
	} else {
		inputView := m.urlInput.View()
		styledInput := lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color(ColorBorder)).
			Padding(0, 1).
			Width(m.urlInput.Width + 2).
			Render(inputView)
		b.WriteString(styledInput)
	}
	b.WriteString("\n\n")

	headersCount := len(m.headers)
	headersText := fmt.Sprintf("Headers: (%d)", headersCount)
	if m.focusIndex == 2 {
		b.WriteString(MutedStyle.Render(headersText) + " ")
		b.WriteString(TextStyle.Render("[Not implemented yet]"))
	} else {
		b.WriteString(DimStyle.Render(headersText + " [Not implemented yet]"))
	}
	b.WriteString("\n\n\n")

	buttons := RenderButton("Send Request", m.focusIndex == 3) + "  "
	buttons += RenderButton("Load Saved", m.focusIndex == 2) + "  "
	buttons += RenderButton("Quit", m.focusIndex == 4)
	b.WriteString(buttons)

	b.WriteString("\n\n")
	b.WriteString(RenderFooter("Tab: next • Enter: action • Ctrl+Q: quit • Ctrl+?: help"))

	return Center(m.width, m.height, b.String())
}

func (m Model) viewLoading() string {
	var b strings.Builder

	b.WriteString(TitleStyle.Render("Sending Request"))
	b.WriteString("\n\n")

	requestInfo := fmt.Sprintf("%s %s", m.method, m.urlInput.Value())
	b.WriteString(TextStyle.Render(requestInfo))
	b.WriteString("\n\n")

	loadingBox := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color(ColorAccent)).
		Padding(2, 4).
		Render(SpinnerStyle.Render(m.spinner.View()) + "  " + TextStyle.Render("Loading..."))

	b.WriteString(loadingBox)
	b.WriteString("\n\n")
	b.WriteString(MutedStyle.Render("Please wait while we fetch the response"))

	return Center(m.width, m.height, b.String())
}

func (m Model) viewResponse() string {
	if m.response == nil {
		return Center(m.width, m.height, ErrorStyle.Render("No response"))
	}

	var b strings.Builder

	b.WriteString(TitleStyle.Render("Response"))
	b.WriteString("\n\n")

	requestInfo := fmt.Sprintf("%s %s", m.method, m.urlInput.Value())
	b.WriteString(MutedStyle.Render(requestInfo))
	b.WriteString("\n\n")

	if m.response.Error != nil {
		errorPanel := lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color(ColorError)).
			Padding(1, 2).
			Width(m.width - 10).
			Render(ErrorStyle.Render(fmt.Sprintf("Error: %v", m.response.Error)))
		b.WriteString(errorPanel)
	} else {
		statusStyle := GetStatusStyle(m.response.StatusCode)
		statusLine := fmt.Sprintf("Status: %s • %s • %s",
			m.response.Status,
			httpclient.FormatDuration(m.response.ResponseTime),
			httpclient.FormatSize(m.response.Size))
		b.WriteString(statusStyle.Render(statusLine))
		b.WriteString("\n\n")

		maxLines := m.height - 15
		lines := strings.Split(m.response.Body, "\n")
		totalLines := len(lines)

		start := m.scrollOffset
		end := start + maxLines
		if end > totalLines {
			end = totalLines
		}
		if start >= totalLines {
			start = totalLines - maxLines
			if start < 0 {
				start = 0
			}
			m.scrollOffset = start
		}

		responsePanel := ""
		if start < totalLines {
			visibleLines := lines[start:end]
			responseContent := strings.Join(visibleLines, "\n")

			scrollInfo := ""
			if totalLines > maxLines {
				scrollInfo = fmt.Sprintf("\n\n%s Lines %d-%d of %d",
					MutedStyle.Render("│"),
					start+1,
					end,
					totalLines)
			}

			responsePanel = lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(lipgloss.Color(ColorBorder)).
				Padding(1, 2).
				Width(m.width - 10).
				Render(CodeStyle.Render(responseContent) + scrollInfo)
		}
		b.WriteString(responsePanel)
	}

	b.WriteString("\n\n")

	buttons := RenderButton("Back", true) + "  "
	buttons += RenderButton("Save (s)", false) + "  "
	if m.response.Error == nil {
		buttons += RenderButton("Copy", false)
	}
	b.WriteString(buttons)

	b.WriteString("\n\n")
	b.WriteString(RenderFooter("Esc: back • s: save • ↑↓: scroll"))

	return Center(m.width, m.height, b.String())
}

func (m Model) viewRequestList() string {
	var b strings.Builder

	b.WriteString(TitleStyle.Render("Saved Requests"))
	b.WriteString("\n\n")

	if len(m.savedRequests) == 0 {
		b.WriteString(MutedStyle.Render("No saved requests"))
	} else {
		for i, req := range m.savedRequests {
			if i == m.selectedReqIdx {
				b.WriteString(ListItemSelectedStyle.Render("> " + req.Name))
				b.WriteString("  ")
				b.WriteString(ButtonActive.Render(req.Method))
			} else {
				b.WriteString(ListItemStyle.Render(req.Name))
				b.WriteString("  ")
				b.WriteString(MutedStyle.Render(req.Method))
			}
			b.WriteString("\n")
		}
	}

	b.WriteString("\n\n")
	b.WriteString(RenderFooter("↑↓: navigate • Enter: load • d: delete • n: new • Esc: back"))

	return Center(m.width, m.height, b.String())
}

func (m Model) viewHelp() string {
	var b strings.Builder

	b.WriteString(TitleStyle.Render("DevScope - Help"))
	b.WriteString("\n\n")

	b.WriteString(HeaderStyle.Render("Global Shortcuts:"))
	b.WriteString("\n")
	b.WriteString(TextStyle.Render("  Ctrl+Q        Quit application"))
	b.WriteString("\n")
	b.WriteString(TextStyle.Render("  Ctrl+?        Show this help"))
	b.WriteString("\n")
	b.WriteString(TextStyle.Render("  Esc           Back/Cancel"))
	b.WriteString("\n")
	b.WriteString(TextStyle.Render("  Tab           Next field"))
	b.WriteString("\n\n")

	b.WriteString(HeaderStyle.Render("Request Builder:"))
	b.WriteString("\n")
	b.WriteString(TextStyle.Render("  Enter         Send request"))
	b.WriteString("\n")
	b.WriteString(TextStyle.Render("  Ctrl+L        Load saved requests"))
	b.WriteString("\n")
	b.WriteString(TextStyle.Render("  ←/→           Change method"))
	b.WriteString("\n\n")

	b.WriteString(HeaderStyle.Render("Response View:"))
	b.WriteString("\n")
	b.WriteString(TextStyle.Render("  s             Save request"))
	b.WriteString("\n")
	b.WriteString(TextStyle.Render("  ↑/↓           Scroll"))
	b.WriteString("\n\n")

	b.WriteString(HeaderStyle.Render("Request List:"))
	b.WriteString("\n")
	b.WriteString(TextStyle.Render("  Enter         Load request"))
	b.WriteString("\n")
	b.WriteString(TextStyle.Render("  d             Delete request"))
	b.WriteString("\n")
	b.WriteString(TextStyle.Render("  n             New request"))
	b.WriteString("\n\n")

	b.WriteString(RenderFooter("Press any key to close"))

	return Center(m.width, m.height, b.String())
}
