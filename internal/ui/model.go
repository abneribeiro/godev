package ui

import (
	"encoding/json"
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/atotto/clipboard"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	httpclient "github.com/abneribeiro/godev/internal/http"
	"github.com/abneribeiro/godev/internal/storage"
)

type AppState int

const (
	StateRequestBuilder AppState = iota
	StateLoading
	StateViewResponse
	StateRequestList
	StateHeaderEditor
	StateBodyEditor
	StateQueryEditor
	StateHelp
)

type Model struct {
	state   AppState
	width   int
	height  int
	storage *storage.Storage

	method     string
	urlInput   textinput.Model
	headers    map[string]string
	body       string
	focusIndex int

	httpClient *httpclient.Client
	response   *httpclient.Response
	spinner    spinner.Model
	loading    bool

	savedRequests  []storage.SavedRequest
	selectedReqIdx int
	scrollOffset   int

	headerKeyInput   textinput.Model
	headerValueInput textinput.Model
	headerList       []string
	selectedHeader   int
	editingHeader    bool

	bodyEditor  textarea.Model
	editingBody bool
	bodyError   string

	queryParams     map[string]string
	queryKeyInput   textinput.Model
	queryValueInput textinput.Model
	queryList       []string
	selectedQuery   int
	editingQuery    bool

	viewResponseHeaders bool
	responseScrollY     int

	urlError              string
	copySuccess           bool
	copySuccessTimer      int
	saveSuccess           bool
	saveSuccessTimer      int
	confirmingDelete      bool
	requestToDelete       int
	requestSaved          bool
	currentRequestSavedID string

	err error
}

type tickMsg time.Time

type responseMsg httpclient.Response

func NewModel() *Model {
	ti := textinput.New()
	ti.Placeholder = "https://api.example.com/endpoint"
	ti.Focus()
	ti.CharLimit = 2000
	ti.Width = 60

	headerKey := textinput.New()
	headerKey.Placeholder = "Header-Name"
	headerKey.CharLimit = 100
	headerKey.Width = 30

	headerValue := textinput.New()
	headerValue.Placeholder = "Header Value"
	headerValue.CharLimit = 500
	headerValue.Width = 50

	queryKey := textinput.New()
	queryKey.Placeholder = "Param Name"
	queryKey.CharLimit = 100
	queryKey.Width = 30

	queryValue := textinput.New()
	queryValue.Placeholder = "Param Value"
	queryValue.CharLimit = 500
	queryValue.Width = 50

	bodyTextarea := textarea.New()
	bodyTextarea.Placeholder = "{\n  \"key\": \"value\"\n}"
	bodyTextarea.CharLimit = 10000
	bodyTextarea.SetWidth(80)
	bodyTextarea.SetHeight(10)

	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = SpinnerStyle

	store, storageErr := storage.NewStorage()
	if storageErr != nil {
		fmt.Printf("Warning: Failed to initialize storage: %v\n", storageErr)
		fmt.Println("The application will continue but requests cannot be saved.")
		fmt.Println("Press Enter to continue...")
		fmt.Scanln()
	}

	m := &Model{
		state:            StateRequestBuilder,
		method:           "GET",
		urlInput:         ti,
		headers:          make(map[string]string),
		body:             "",
		focusIndex:       1,
		httpClient:       httpclient.NewClient(30 * time.Second),
		spinner:          s,
		storage:          store,
		err:              nil,
		headerKeyInput:   headerKey,
		headerValueInput: headerValue,
		headerList:       []string{},
		selectedHeader:   0,
		editingHeader:    false,
		bodyEditor:       bodyTextarea,
		editingBody:      false,
		queryParams:      make(map[string]string),
		queryKeyInput:    queryKey,
		queryValueInput:  queryValue,
		queryList:        []string{},
		selectedQuery:    0,
		editingQuery:     false,
		viewResponseHeaders: false,
		responseScrollY:  0,
		urlError:         "",
		copySuccess:      false,
		copySuccessTimer: 0,
	}

	if m.storage != nil {
		m.savedRequests = m.storage.GetRequests()
	}

	return m
}

func (m Model) Init() tea.Cmd {
	return tea.Batch(
		textinput.Blink,
		tickCmd(),
	)
}

func tickCmd() tea.Cmd {
	return tea.Tick(time.Second, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
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
				m.requestSaved = false
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

	case tickMsg:
		if m.copySuccessTimer > 0 {
			m.copySuccessTimer--
			if m.copySuccessTimer == 0 {
				m.copySuccess = false
			}
		}
		if m.saveSuccessTimer > 0 {
			m.saveSuccessTimer--
			if m.saveSuccessTimer == 0 {
				m.saveSuccess = false
			}
		}
		return m, tickCmd()

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
	case StateHeaderEditor:
		return m.handleHeaderEditorKeys(msg)
	case StateBodyEditor:
		return m.handleBodyEditorKeys(msg)
	case StateQueryEditor:
		return m.handleQueryEditorKeys(msg)
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
		if m.focusIndex > 7 {
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
			m.focusIndex = 7
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
		case 0:
			return m, nil
		case 1:
			if m.urlInput.Value() != "" {
				return m, m.sendRequest()
			}
			return m, nil
		case 2:
			m.state = StateQueryEditor
			m.buildQueryList()
			return m, nil
		case 3:
			m.state = StateHeaderEditor
			m.buildHeaderList()
			return m, nil
		case 4:
			m.state = StateBodyEditor
			m.bodyEditor.SetValue(m.body)
			m.bodyEditor.Focus()
			return m, nil
		case 5:
			if m.urlInput.Value() != "" {
				return m, m.sendRequest()
			}
		case 6:
			m.state = StateRequestList
			return m, nil
		case 7:
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
		m.viewResponseHeaders = false
		return m, nil

	case "s":
		if m.storage != nil && m.response != nil {
			name := fmt.Sprintf("%s %s", m.method, m.urlInput.Value())
			if !m.storage.RequestExists(name) {
				err := m.storage.SaveRequest(name, m.method, m.urlInput.Value(), m.headers, m.body, m.queryParams)
				if err == nil {
					m.savedRequests = m.storage.GetRequests()
					m.saveSuccess = true
					m.saveSuccessTimer = 3
					m.requestSaved = true
					if len(m.savedRequests) > 0 {
						m.currentRequestSavedID = m.savedRequests[len(m.savedRequests)-1].ID
					}
				}
			}
		}
		return m, nil

	case "c":
		if m.response != nil && m.response.Error == nil {
			err := clipboard.WriteAll(m.response.Body)
			if err == nil {
				m.copySuccess = true
				m.copySuccessTimer = 3
			}
		}
		return m, nil

	case "h":
		m.viewResponseHeaders = !m.viewResponseHeaders
		m.scrollOffset = 0
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
		if m.confirmingDelete {
			m.confirmingDelete = false
			return m, nil
		}
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
			if req.QueryParams != nil {
				m.queryParams = req.QueryParams
			} else {
				m.queryParams = make(map[string]string)
			}
			m.state = StateRequestBuilder
			m.requestSaved = true
			m.currentRequestSavedID = req.ID

			if m.storage != nil {
				m.storage.UpdateLastUsed(req.ID)
			}
		}
		return m, nil

	case "d":
		if len(m.savedRequests) > 0 && m.selectedReqIdx < len(m.savedRequests) {
			if !m.confirmingDelete {
				m.confirmingDelete = true
				m.requestToDelete = m.selectedReqIdx
				return m, nil
			}
		}
		return m, nil

	case "y":
		if m.confirmingDelete && m.storage != nil {
			if m.requestToDelete < len(m.savedRequests) {
				req := m.savedRequests[m.requestToDelete]
				m.storage.DeleteRequest(req.ID)
				m.savedRequests = m.storage.GetRequests()
				if m.selectedReqIdx >= len(m.savedRequests) && m.selectedReqIdx > 0 {
					m.selectedReqIdx--
				}
			}
			m.confirmingDelete = false
			return m, nil
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

func (m Model) handleHelpKeys(_ tea.KeyMsg) (tea.Model, tea.Cmd) {
	m.state = StateRequestBuilder
	return m, nil
}

func (m *Model) validateURL(urlStr string) error {
	if urlStr == "" {
		return fmt.Errorf("url cannot be empty")
	}

	parsedURL, err := url.Parse(urlStr)
	if err != nil {
		return fmt.Errorf("invalid url: %v", err)
	}

	if parsedURL.Scheme == "" {
		return fmt.Errorf("url must include protocol (http:// or https://)")
	}

	if parsedURL.Scheme != "http" && parsedURL.Scheme != "https" {
		return fmt.Errorf("protocol must be http or https")
	}

	if parsedURL.Host == "" {
		return fmt.Errorf("url must include a valid host")
	}

	return nil
}

func (m *Model) validateJSON(body string) error {
	if body == "" {
		return nil
	}

	var js interface{}
	if err := json.Unmarshal([]byte(body), &js); err != nil {
		return fmt.Errorf("invalid json: %v", err)
	}
	return nil
}

func (m *Model) buildURLWithQueryParams() string {
	baseURL := m.urlInput.Value()
	if len(m.queryParams) == 0 {
		return baseURL
	}

	parsedURL, err := url.Parse(baseURL)
	if err != nil {
		return baseURL
	}

	q := parsedURL.Query()
	for key, value := range m.queryParams {
		q.Set(key, value)
	}
	parsedURL.RawQuery = q.Encode()

	return parsedURL.String()
}

func (m Model) sendRequest() tea.Cmd {
	urlStr := m.urlInput.Value()

	if err := m.validateURL(urlStr); err != nil {
		return func() tea.Msg {
			resp := httpclient.Response{
				Error: err,
			}
			return responseMsg(resp)
		}
	}

	m.state = StateLoading
	m.loading = true
	m.scrollOffset = 0
	m.urlError = ""

	finalURL := m.buildURLWithQueryParams()

	return tea.Batch(
		m.spinner.Tick,
		func() tea.Msg {
			req := httpclient.Request{
				Method:  m.method,
				URL:     finalURL,
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
	case StateHeaderEditor:
		return m.viewHeaderEditor()
	case StateBodyEditor:
		return m.viewBodyEditor()
	case StateQueryEditor:
		return m.viewQueryEditor()
	case StateHelp:
		return m.viewHelp()
	}

	return ""
}

func (m Model) viewRequestBuilder() string {
	var b strings.Builder

	title := "GoDev v0.2.0"
	if m.requestSaved {
		title += " [SAVED]"
	}
	b.WriteString(TitleStyle.Render(title))
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
	b.WriteString("\n")

	if len(m.queryParams) > 0 {
		finalURL := m.buildURLWithQueryParams()
		b.WriteString(MutedStyle.Render(fmt.Sprintf("    → Final URL: %s", finalURL)))
		b.WriteString("\n")
	}
	b.WriteString("\n")

	queryCount := len(m.queryParams)
	queryText := fmt.Sprintf("Query Params: (%d)", queryCount)
	if m.focusIndex == 2 {
		b.WriteString(ButtonActive.Render("[ " + queryText + " ]"))
	} else {
		b.WriteString(MutedStyle.Render(queryText))
	}
	b.WriteString("\n")

	headersCount := len(m.headers)
	headersText := fmt.Sprintf("Headers: (%d)", headersCount)
	if m.focusIndex == 3 {
		b.WriteString(ButtonActive.Render("[ " + headersText + " ]"))
	} else {
		b.WriteString(MutedStyle.Render(headersText))
	}
	b.WriteString("\n")

	bodyPreview := "empty"
	if m.body != "" {
		bodyStr := strings.ReplaceAll(m.body, "\n", " ")
		bodyStr = strings.TrimSpace(bodyStr)
		if len(bodyStr) > 80 {
			bodyPreview = bodyStr[:80] + "..."
		} else {
			bodyPreview = bodyStr
		}
	}
	bodyText := fmt.Sprintf("Body: (%s)", bodyPreview)
	if m.focusIndex == 4 {
		b.WriteString(ButtonActive.Render("[ " + bodyText + " ]"))
	} else {
		b.WriteString(MutedStyle.Render(bodyText))
	}
	b.WriteString("\n\n")

	buttons := RenderButton("Send Request", m.focusIndex == 5) + "  "
	buttons += RenderButton("Load Saved", m.focusIndex == 6) + "  "
	buttons += RenderButton("Quit", m.focusIndex == 7)
	b.WriteString(buttons)

	b.WriteString("\n\n")
	b.WriteString(RenderFooter("Tab: next • Enter: action/send • Ctrl+Q: quit • Ctrl+L: load saved"))

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

	title := "Response"
	if m.viewResponseHeaders {
		title = "Response Headers"
	}
	b.WriteString(TitleStyle.Render(title))
	b.WriteString("\n\n")

	requestInfo := fmt.Sprintf("%s %s", m.method, m.buildURLWithQueryParams())
	b.WriteString(MutedStyle.Render(requestInfo))
	b.WriteString("\n\n")

	if m.saveSuccess {
		b.WriteString(SuccessStyle.Render("✓ Request saved successfully!"))
		b.WriteString("\n\n")
	}

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

		if m.copySuccess {
			b.WriteString(SuccessStyle.Render("✓ Copied to clipboard!"))
			b.WriteString("\n\n")
		}

		var content string
		if m.viewResponseHeaders {
			var headerLines []string
			for key, values := range m.response.Headers {
				for _, value := range values {
					headerLines = append(headerLines, fmt.Sprintf("%-30s : %s", key, value))
				}
			}
			content = strings.Join(headerLines, "\n")
		} else {
			content = m.response.Body
		}

		maxLines := m.height - 17
		lines := strings.Split(content, "\n")
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

	buttons := RenderButton("Back (Esc)", true) + "  "
	buttons += RenderButton("Save (s)", false) + "  "
	if m.response.Error == nil {
		buttons += RenderButton("Copy (c)", false) + "  "
		if m.viewResponseHeaders {
			buttons += RenderButton("Body (h)", false)
		} else {
			buttons += RenderButton("Headers (h)", false)
		}
	}
	b.WriteString(buttons)

	b.WriteString("\n\n")
	b.WriteString(RenderFooter("Esc: back • s: save • c: copy • h: toggle headers • ↑↓: scroll"))

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

	if m.confirmingDelete && len(m.savedRequests) > 0 && m.requestToDelete < len(m.savedRequests) {
		confirmMsg := fmt.Sprintf("⚠ Delete '%s'? Press 'y' to confirm, 'Esc' to cancel", m.savedRequests[m.requestToDelete].Name)
		b.WriteString(WarningStyle.Render(confirmMsg))
		b.WriteString("\n\n")
	}

	b.WriteString(RenderFooter("↑↓: navigate • Enter: load • d: delete • n: new • Esc: back"))

	return Center(m.width, m.height, b.String())
}

func (m Model) viewHelp() string {
	var b strings.Builder

	b.WriteString(TitleStyle.Render("GoDev - Help"))
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
