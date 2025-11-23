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

	"github.com/abneribeiro/godev/internal/database"
	httpclient "github.com/abneribeiro/godev/internal/http"
	"github.com/abneribeiro/godev/internal/storage"
)

type AppState int

const (
	StateHome AppState = iota
	StateRequestBuilder
	StateLoading
	StateViewResponse
	StateRequestList
	StateHeaderEditor
	StateBodyEditor
	StateQueryEditor
	StateHelp
	StateHistory
	StateDatabase
	StateDatabaseConnect
	StateDatabaseQueryEditor
	StateDatabaseResult
	StateDatabaseQueryList
	StateDatabaseSchema
	StateDatabaseQueryHistory
	StateDatabaseExport
	StateEnvironments
	StateEnvironmentEditor
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

	savedRequests    []storage.SavedRequest
	filteredRequests []storage.SavedRequest
	selectedReqIdx   int
	scrollOffset     int
	searchInput      textinput.Model
	searchActive     bool

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
	curlCopySuccess       bool
	curlCopySuccessTimer  int
	confirmingDelete      bool
	requestToDelete       int
	requestSaved          bool
	currentRequestSavedID string

	history              []storage.RequestExecution
	selectedHistoryIdx   int
	historyScrollOffset  int
	confirmingClearHistory bool

	dbClient                *database.PostgresClient
	dbStorage               *database.DatabaseStorage
	dbConnectHostInput      textinput.Model
	dbConnectPortInput      textinput.Model
	dbConnectDatabaseInput  textinput.Model
	dbConnectUserInput      textinput.Model
	dbConnectPasswordInput  textinput.Model
	dbConnectFocusIndex     int
	dbQueryEditor           textarea.Model
	dbQueryResult           *database.QueryResult
	dbSavedQueries          []database.SavedQuery
	dbSelectedQueryIdx      int
	dbMode                  string
	dbTables                []string
	dbSelectedTableIdx      int
	dbTableInfo             *database.TableInfo
	dbQuerySaveSuccess           bool
	dbQuerySaveSuccessTimer      int
	dbConnectSuccess             bool
	dbConnectSuccessTimer        int
	dbQueryHistory               []database.QueryExecution
	dbSelectedQueryHistoryIdx    int
	dbConfirmingClearQueryHistory bool
	dbExportFormatIdx            int
	dbExportTableName            textinput.Model
	dbExportSuccess              bool
	dbExportSuccessTimer         int
	dbExportFilePath             string

	envConfig               *storage.EnvironmentConfig
	envList                 []storage.Environment
	selectedEnvIdx          int
	envScrollOffset         int
	envNameInput            textinput.Model
	envVarKeyInput          textinput.Model
	envVarValueInput        textinput.Model
	envVarList              []storage.Variable
	selectedEnvVarIdx       int
	editingEnvVar           bool
	envFocusIndex           int
	envSaveSuccess          bool
	envSaveSuccessTimer     int
	envDeleteSuccess        bool
	envDeleteSuccessTimer   int
	currentEnvName          string
	confirmingDeleteEnv     bool
	confirmingDeleteEnvVar  bool
	// envVarToDelete          int

	err error
}

type tickMsg time.Time

type responseMsg httpclient.Response

type databaseSchemaMsg []string

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

	searchInput := textinput.New()
	searchInput.Placeholder = "Search requests..."
	searchInput.CharLimit = 100
	searchInput.Width = 50

	dbHostInput := textinput.New()
	dbHostInput.Placeholder = "localhost"
	dbHostInput.CharLimit = 100
	dbHostInput.Width = 40
	dbHostInput.SetValue("localhost")

	dbPortInput := textinput.New()
	dbPortInput.Placeholder = "5432"
	dbPortInput.CharLimit = 10
	dbPortInput.Width = 15
	dbPortInput.SetValue("5432")

	dbDatabaseInput := textinput.New()
	dbDatabaseInput.Placeholder = "database name"
	dbDatabaseInput.CharLimit = 100
	dbDatabaseInput.Width = 40

	dbUserInput := textinput.New()
	dbUserInput.Placeholder = "username"
	dbUserInput.CharLimit = 100
	dbUserInput.Width = 40

	dbPasswordInput := textinput.New()
	dbPasswordInput.Placeholder = "password"
	dbPasswordInput.CharLimit = 100
	dbPasswordInput.Width = 40
	dbPasswordInput.EchoMode = textinput.EchoPassword
	dbPasswordInput.EchoCharacter = '•'

	dbQueryTextarea := textarea.New()
	dbQueryTextarea.Placeholder = "SELECT * FROM table_name;"
	dbQueryTextarea.CharLimit = 50000
	dbQueryTextarea.SetWidth(80)
	dbQueryTextarea.SetHeight(10)
	// Disable Ctrl+K built-in behavior (delete line) so we can use it for query execution
	dbQueryTextarea.KeyMap.DeleteAfterCursor.SetEnabled(false)

	dbExportTableName := textinput.New()
	dbExportTableName.Placeholder = "table_name"
	dbExportTableName.CharLimit = 100
	dbExportTableName.Width = 40

	envNameInput := textinput.New()
	envNameInput.Placeholder = "environment name (e.g., dev, staging, prod)"
	envNameInput.CharLimit = 50
	envNameInput.Width = 50

	envVarKey := textinput.New()
	envVarKey.Placeholder = "Variable Name (e.g., API_URL)"
	envVarKey.CharLimit = 100
	envVarKey.Width = 30

	envVarValue := textinput.New()
	envVarValue.Placeholder = "Variable Value"
	envVarValue.CharLimit = 500
	envVarValue.Width = 50

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

	if store != nil {
		_, envErr := store.LoadEnvironments()
		if envErr != nil {
			fmt.Printf("Warning: Failed to initialize environments: %v\n", envErr)
		}
	}

	dbStorage, dbStorageErr := database.NewDatabaseStorage()
	if dbStorageErr != nil {
		fmt.Printf("Warning: Failed to initialize database storage: %v\n", dbStorageErr)
	}

	dbClient := database.NewPostgresClient()

	m := &Model{
		state:               StateHome,
		method:              "GET",
		urlInput:            ti,
		headers:             make(map[string]string),
		body:                "",
		focusIndex:          1,
		httpClient:          httpclient.NewClient(30 * time.Second),
		spinner:             s,
		storage:             store,
		err:                 nil,
		headerKeyInput:      headerKey,
		headerValueInput:    headerValue,
		headerList:          []string{},
		selectedHeader:      0,
		editingHeader:       false,
		bodyEditor:          bodyTextarea,
		editingBody:         false,
		queryParams:         make(map[string]string),
		queryKeyInput:       queryKey,
		queryValueInput:     queryValue,
		queryList:           []string{},
		selectedQuery:       0,
		editingQuery:        false,
		viewResponseHeaders: false,
		responseScrollY:     0,
		urlError:            "",
		copySuccess:         false,
		copySuccessTimer:       0,
		searchInput:            searchInput,
		searchActive:           false,
		dbClient:               dbClient,
		dbStorage:              dbStorage,
		dbConnectHostInput:     dbHostInput,
		dbConnectPortInput:     dbPortInput,
		dbConnectDatabaseInput: dbDatabaseInput,
		dbConnectUserInput:     dbUserInput,
		dbConnectPasswordInput: dbPasswordInput,
		dbConnectFocusIndex:    0,
		dbQueryEditor:          dbQueryTextarea,
		dbQueryResult:          nil,
		dbSavedQueries:         []database.SavedQuery{},
		dbSelectedQueryIdx:     0,
		dbMode:                 "menu",
		dbExportTableName:      dbExportTableName,
		dbExportFormatIdx:      0,
		envNameInput:           envNameInput,
		envVarKeyInput:         envVarKey,
		envVarValueInput:       envVarValue,
		selectedEnvIdx:         0,
		envScrollOffset:        0,
		editingEnvVar:          false,
		envFocusIndex:          0,
		selectedEnvVarIdx:      0,
	}

	if m.storage != nil {
		m.savedRequests = m.storage.GetRequests()
		m.history = m.storage.GetHistory()
		envConfig, _ := m.storage.LoadEnvironments()
		if envConfig != nil {
			m.envConfig = envConfig
			m.envList = envConfig.Environments
		}
	}

	if m.dbStorage != nil {
		m.dbSavedQueries = m.dbStorage.GetQueries()
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

		if m.storage != nil {
			statusCode := 0
			status := ""
			responseBody := ""
			responseTimeMs := int64(0)
			var err error

			if resp.Error != nil {
				err = resp.Error
			} else {
				statusCode = resp.StatusCode
				status = resp.Status
				responseBody = resp.Body
				responseTimeMs = resp.ResponseTime.Milliseconds()
			}

			finalURL := m.buildURLWithQueryParams()
			m.storage.AddToHistory(m.method, finalURL, m.headers, m.body, m.queryParams, statusCode, status, responseBody, responseTimeMs, err)
			m.history = m.storage.GetHistory()
		}

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
		if m.curlCopySuccessTimer > 0 {
			m.curlCopySuccessTimer--
			if m.curlCopySuccessTimer == 0 {
				m.curlCopySuccess = false
			}
		}
		if m.dbQuerySaveSuccessTimer > 0 {
			m.dbQuerySaveSuccessTimer--
			if m.dbQuerySaveSuccessTimer == 0 {
				m.dbQuerySaveSuccess = false
			}
		}
		if m.dbConnectSuccessTimer > 0 {
			m.dbConnectSuccessTimer--
			if m.dbConnectSuccessTimer == 0 {
				m.dbConnectSuccess = false
			}
		}
		if m.dbExportSuccessTimer > 0 {
			m.dbExportSuccessTimer--
			if m.dbExportSuccessTimer == 0 {
				m.dbExportSuccess = false
			}
		}
		if m.envSaveSuccessTimer > 0 {
			m.envSaveSuccessTimer--
			if m.envSaveSuccessTimer == 0 {
				m.envSaveSuccess = false
			}
		}
		if m.envDeleteSuccessTimer > 0 {
			m.envDeleteSuccessTimer--
			if m.envDeleteSuccessTimer == 0 {
				m.envDeleteSuccess = false
			}
		}
		return m, tickCmd()

	case databaseResultMsg:
		m.loading = false
		result := database.QueryResult(msg)
		m.dbQueryResult = &result

		if m.dbStorage != nil {
			query := strings.TrimSpace(m.dbQueryEditor.Value())
			connectionInfo := m.dbClient.GetConnectionString()
			m.dbStorage.AddToQueryHistory(query, connectionInfo, result.RowsAffected, result.ExecutionTime.Milliseconds(), result.Error)
		}

		m.state = StateDatabaseResult
		return m, nil

	case databaseSchemaMsg:
		m.loading = false
		m.dbTables = []string(msg)
		m.dbSelectedTableIdx = 0
		m.dbConnectSuccess = true
		m.dbConnectSuccessTimer = 3
		m.state = StateDatabaseSchema
		return m, nil

	case spinner.TickMsg:
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd
	}

	return m, cmd
}

func (m Model) handleKeyPress(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch m.state {
	case StateHome:
		return m.handleHomeKeys(msg)
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
	case StateHistory:
		return m.handleHistoryKeys(msg)
	case StateDatabase:
		return m.handleDatabaseKeys(msg)
	case StateDatabaseConnect:
		return m.handleDatabaseConnectKeys(msg)
	case StateDatabaseQueryEditor:
		return m.handleDatabaseQueryEditorKeys(msg)
	case StateDatabaseResult:
		return m.handleDatabaseResultKeys(msg)
	case StateDatabaseQueryList:
		return m.handleDatabaseQueryListKeys(msg)
	case StateDatabaseSchema:
		return m.handleDatabaseSchemaKeys(msg)
	case StateDatabaseQueryHistory:
		return m.handleDatabaseQueryHistoryKeys(msg)
	case StateDatabaseExport:
		return m.handleDatabaseExportKeys(msg)
	case StateEnvironments:
		return m.handleEnvironmentsKeys(msg)
	case StateEnvironmentEditor:
		return m.handleEnvironmentEditorKeys(msg)
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

	case "ctrl+h", "?":
		m.state = StateHelp
		return m, nil

	case "ctrl+enter":
		if m.urlInput.Value() != "" {
			return m, m.sendRequest()
		}
		return m, nil

	case "ctrl+l":
		m.state = StateRequestList
		return m, nil

	case "ctrl+r":
		m.state = StateHistory
		m.selectedHistoryIdx = 0
		m.historyScrollOffset = 0
		return m, nil

	case "ctrl+d":
		m.state = StateDatabase
		return m, nil

	case "ctrl+e":
		m.state = StateEnvironments
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

	case "left":
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

	case "right":
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

	case "h":
		m.state = StateHeaderEditor
		m.buildHeaderList()
		return m, nil

	case "b":
		m.state = StateBodyEditor
		m.bodyEditor.SetValue(m.body)
		m.bodyEditor.Focus()
		return m, nil

	case "q":
		m.state = StateQueryEditor
		m.buildQueryList()
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

	case "s":
		if m.storage != nil && m.urlInput.Value() != "" {
			name := fmt.Sprintf("%s %s", m.method, m.urlInput.Value())
			if !m.storage.RequestExists(name) {
				err := m.storage.SaveRequest(name, m.method, m.urlInput.Value(), m.headers, m.body, m.queryParams)
				if err == nil {
					m.savedRequests = m.storage.GetRequests()
					m.saveSuccess = true
					m.saveSuccessTimer = 3
				}
			}
		}
		return m, nil

	case "x":
		if m.urlInput.Value() != "" {
			finalURL := m.buildURLWithQueryParams()
			req := httpclient.Request{
				Method:  m.method,
				URL:     finalURL,
				Headers: m.headers,
				Body:    m.body,
			}
			curlCmd := httpclient.RequestToCurl(req)
			err := clipboard.WriteAll(curlCmd)
			if err == nil {
				m.curlCopySuccess = true
				m.curlCopySuccessTimer = 3
			}
		}
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

	case "x":
		finalURL := m.buildURLWithQueryParams()
		req := httpclient.Request{
			Method:  m.method,
			URL:     finalURL,
			Headers: m.headers,
			Body:    m.body,
		}
		curlCmd := httpclient.RequestToCurl(req)
		err := clipboard.WriteAll(curlCmd)
		if err == nil {
			m.curlCopySuccess = true
			m.curlCopySuccessTimer = 3
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
	var cmd tea.Cmd

	if m.searchActive {
		switch msg.String() {
		case "ctrl+c", "ctrl+q":
			return m, tea.Quit
		case "esc":
			m.searchActive = false
			m.searchInput.Blur()
			m.searchInput.SetValue("")
			m.filteredRequests = m.savedRequests
			m.selectedReqIdx = 0
			return m, nil
		case "enter":
			m.searchActive = false
			m.searchInput.Blur()
			return m, nil
		default:
			m.searchInput, cmd = m.searchInput.Update(msg)
			if m.storage != nil {
				m.filteredRequests = m.storage.FilterRequests(m.searchInput.Value())
				if m.selectedReqIdx >= len(m.filteredRequests) {
					m.selectedReqIdx = 0
				}
			}
			return m, cmd
		}
	}

	switch msg.String() {
	case "ctrl+c", "ctrl+q":
		return m, tea.Quit

	case "esc":
		if m.confirmingDelete {
			m.confirmingDelete = false
			return m, nil
		}
		m.state = StateRequestBuilder
		m.searchInput.SetValue("")
		m.filteredRequests = nil
		return m, nil

	case "/":
		m.searchActive = true
		m.searchInput.Focus()
		if m.filteredRequests == nil {
			m.filteredRequests = m.savedRequests
		}
		return m, nil

	case "up", "k":
		if m.selectedReqIdx > 0 {
			m.selectedReqIdx--
		}
		return m, nil

	case "down", "j":
		displayList := m.savedRequests
		if m.filteredRequests != nil {
			displayList = m.filteredRequests
		}
		if m.selectedReqIdx < len(displayList)-1 {
			m.selectedReqIdx++
		}
		return m, nil

	case "enter":
		displayList := m.savedRequests
		if m.filteredRequests != nil {
			displayList = m.filteredRequests
		}
		if len(displayList) > 0 && m.selectedReqIdx < len(displayList) {
			req := displayList[m.selectedReqIdx]
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
		displayList := m.savedRequests
		if m.filteredRequests != nil {
			displayList = m.filteredRequests
		}
		if len(displayList) > 0 && m.selectedReqIdx < len(displayList) {
			if !m.confirmingDelete {
				m.confirmingDelete = true
				m.requestToDelete = m.selectedReqIdx
				return m, nil
			}
		}
		return m, nil

	case "y":
		if m.confirmingDelete && m.storage != nil {
			displayList := m.savedRequests
			if m.filteredRequests != nil {
				displayList = m.filteredRequests
			}
			if m.requestToDelete < len(displayList) {
				req := displayList[m.requestToDelete]
				m.storage.DeleteRequest(req.ID)
				m.savedRequests = m.storage.GetRequests()
				if m.searchInput.Value() != "" {
					m.filteredRequests = m.storage.FilterRequests(m.searchInput.Value())
				} else {
					m.filteredRequests = nil
				}
				displayList = m.savedRequests
				if m.filteredRequests != nil {
					displayList = m.filteredRequests
				}
				if m.selectedReqIdx >= len(displayList) && m.selectedReqIdx > 0 {
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
	finalHeaders := make(map[string]string)
	for k, v := range m.headers {
		finalHeaders[k] = v
	}
	finalBody := m.body

	if m.storage != nil {
		vars, err := m.storage.GetActiveEnvironmentVariables()
		if err == nil && len(vars) > 0 {
			finalURL = storage.ReplaceVariables(finalURL, vars)
			for k, v := range finalHeaders {
				finalHeaders[k] = storage.ReplaceVariables(v, vars)
			}
			finalBody = storage.ReplaceVariables(finalBody, vars)
		}
	}

	return tea.Batch(
		m.spinner.Tick,
		func() tea.Msg {
			req := httpclient.Request{
				Method:  m.method,
				URL:     finalURL,
				Headers: finalHeaders,
				Body:    finalBody,
			}
			resp := m.httpClient.Send(req)
			return responseMsg(resp)
		},
	)
}

func (m Model) handleEnvironmentsKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "ctrl+c", "ctrl+q":
		return m, tea.Quit

	case "esc":
		if m.confirmingDeleteEnv {
			m.confirmingDeleteEnv = false
			return m, nil
		}
		m.state = StateRequestBuilder
		return m, nil

	case "up", "k":
		if m.selectedEnvIdx > 0 {
			m.selectedEnvIdx--
		}
		return m, nil

	case "down", "j":
		if m.selectedEnvIdx < len(m.envList)-1 {
			m.selectedEnvIdx++
		}
		return m, nil

	case "n", "a":
		m.envNameInput.SetValue("")
		m.envNameInput.Focus()
		m.currentEnvName = ""
		m.envVarList = []storage.Variable{}
		m.selectedEnvVarIdx = 0
		m.state = StateEnvironmentEditor
		return m, nil

	case "enter":
		if len(m.envList) > 0 && m.selectedEnvIdx < len(m.envList) {
			env := m.envList[m.selectedEnvIdx]
			m.currentEnvName = env.Name
			m.envVarList = env.Variables
			m.selectedEnvVarIdx = 0
			m.envNameInput.SetValue(env.Name)
			m.state = StateEnvironmentEditor
		}
		return m, nil

	case "d":
		if len(m.envList) > 0 && m.selectedEnvIdx < len(m.envList) {
			m.confirmingDeleteEnv = true
		}
		return m, nil

	case "y":
		if m.confirmingDeleteEnv && len(m.envList) > 0 && m.selectedEnvIdx < len(m.envList) {
			env := m.envList[m.selectedEnvIdx]
			if m.storage != nil {
				err := m.storage.DeleteEnvironment(env.Name)
				if err == nil {
					envConfig, _ := m.storage.LoadEnvironments()
					if envConfig != nil {
						m.envConfig = envConfig
						m.envList = envConfig.Environments
					}
					if m.selectedEnvIdx >= len(m.envList) && m.selectedEnvIdx > 0 {
						m.selectedEnvIdx--
					}
					m.envDeleteSuccess = true
					m.envDeleteSuccessTimer = 3
				}
			}
			m.confirmingDeleteEnv = false
		}
		return m, nil

	case "s":
		if len(m.envList) > 0 && m.selectedEnvIdx < len(m.envList) {
			env := m.envList[m.selectedEnvIdx]
			if m.storage != nil {
				m.storage.SetActiveEnvironment(env.Name)
				envConfig, _ := m.storage.LoadEnvironments()
				if envConfig != nil {
					m.envConfig = envConfig
					m.envList = envConfig.Environments
				}
				m.envSaveSuccess = true
				m.envSaveSuccessTimer = 3
			}
		}
		return m, nil
	}

	return m, nil
}

func (m Model) handleEnvironmentEditorKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	if m.editingEnvVar {
		switch msg.String() {
		case "ctrl+c", "ctrl+q":
			return m, tea.Quit
		case "esc":
			m.editingEnvVar = false
			m.envVarKeyInput.Blur()
			m.envVarValueInput.Blur()
			m.envVarKeyInput.SetValue("")
			m.envVarValueInput.SetValue("")
			return m, nil
		case "enter", "tab":
			if m.envFocusIndex == 0 {
				m.envFocusIndex = 1
				m.envVarKeyInput.Blur()
				m.envVarValueInput.Focus()
				return m, nil
			} else {
				key := strings.TrimSpace(m.envVarKeyInput.Value())
				value := m.envVarValueInput.Value()
				if key != "" && m.storage != nil && m.currentEnvName != "" {
					err := m.storage.AddVariable(m.currentEnvName, key, value)
					if err == nil {
						envConfig, _ := m.storage.LoadEnvironments()
						if envConfig != nil {
							m.envConfig = envConfig
							m.envList = envConfig.Environments
							for _, env := range m.envList {
								if env.Name == m.currentEnvName {
									m.envVarList = env.Variables
									break
								}
							}
						}
						m.envSaveSuccess = true
						m.envSaveSuccessTimer = 3
					}
				}
				m.editingEnvVar = false
				m.envFocusIndex = 0
				m.envVarKeyInput.Blur()
				m.envVarValueInput.Blur()
				m.envVarKeyInput.SetValue("")
				m.envVarValueInput.SetValue("")
				return m, nil
			}
		default:
			if m.envFocusIndex == 0 {
				m.envVarKeyInput, cmd = m.envVarKeyInput.Update(msg)
			} else {
				m.envVarValueInput, cmd = m.envVarValueInput.Update(msg)
			}
			return m, cmd
		}
	}

	switch msg.String() {
	case "ctrl+c", "ctrl+q":
		return m, tea.Quit

	case "esc":
		if m.confirmingDeleteEnvVar {
			m.confirmingDeleteEnvVar = false
			return m, nil
		}
		m.state = StateEnvironments
		m.currentEnvName = ""
		return m, nil

	case "ctrl+s":
		name := strings.TrimSpace(m.envNameInput.Value())
		if name != "" && m.storage != nil {
			if m.currentEnvName == "" {
				err := m.storage.AddEnvironment(name)
				if err == nil {
					m.currentEnvName = name
					envConfig, _ := m.storage.LoadEnvironments()
					if envConfig != nil {
						m.envConfig = envConfig
						m.envList = envConfig.Environments
					}
					m.envSaveSuccess = true
					m.envSaveSuccessTimer = 3
				}
			}
		}
		return m, nil

	case "up", "k":
		if m.selectedEnvVarIdx > 0 {
			m.selectedEnvVarIdx--
		}
		return m, nil

	case "down", "j":
		if m.selectedEnvVarIdx < len(m.envVarList)-1 {
			m.selectedEnvVarIdx++
		}
		return m, nil

	case "n", "a":
		m.editingEnvVar = true
		m.envFocusIndex = 0
		m.envVarKeyInput.SetValue("")
		m.envVarValueInput.SetValue("")
		m.envVarKeyInput.Focus()
		return m, nil

	case "e":
		if len(m.envVarList) > 0 && m.selectedEnvVarIdx < len(m.envVarList) {
			variable := m.envVarList[m.selectedEnvVarIdx]
			m.editingEnvVar = true
			m.envFocusIndex = 0
			m.envVarKeyInput.SetValue(variable.Key)
			m.envVarValueInput.SetValue(variable.Value)
			m.envVarKeyInput.Focus()
		}
		return m, nil

	case "d":
		if len(m.envVarList) > 0 && m.selectedEnvVarIdx < len(m.envVarList) {
			m.confirmingDeleteEnvVar = true
		}
		return m, nil

	case "y":
		if m.confirmingDeleteEnvVar && len(m.envVarList) > 0 && m.selectedEnvVarIdx < len(m.envVarList) {
			variable := m.envVarList[m.selectedEnvVarIdx]
			if m.storage != nil && m.currentEnvName != "" {
				err := m.storage.DeleteVariable(m.currentEnvName, variable.Key)
				if err == nil {
					envConfig, _ := m.storage.LoadEnvironments()
					if envConfig != nil {
						m.envConfig = envConfig
						m.envList = envConfig.Environments
						for _, env := range m.envList {
							if env.Name == m.currentEnvName {
								m.envVarList = env.Variables
								break
							}
						}
					}
					if m.selectedEnvVarIdx >= len(m.envVarList) && m.selectedEnvVarIdx > 0 {
						m.selectedEnvVarIdx--
					}
					m.envDeleteSuccess = true
					m.envDeleteSuccessTimer = 3
				}
			}
			m.confirmingDeleteEnvVar = false
		}
		return m, nil
	}

	return m, nil
}

func (m Model) View() string {
	if m.err != nil {
		return ErrorStyle.Render(fmt.Sprintf("Error: %v\nPress Ctrl+Q to quit", m.err))
	}

	switch m.state {
	case StateHome:
		return m.viewHome()
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
	case StateHistory:
		return m.viewHistory()
	case StateDatabase:
		return m.viewDatabase()
	case StateDatabaseConnect:
		return m.viewDatabaseConnect()
	case StateDatabaseQueryEditor:
		return m.viewDatabaseQueryEditor()
	case StateDatabaseResult:
		return m.viewDatabaseResult()
	case StateDatabaseQueryList:
		return m.viewDatabaseQueryList()
	case StateDatabaseSchema:
		return m.viewDatabaseSchema()
	case StateDatabaseQueryHistory:
		return m.viewDatabaseQueryHistory()
	case StateDatabaseExport:
		return m.viewDatabaseExport()
	case StateEnvironments:
		return m.viewEnvironments()
	case StateEnvironmentEditor:
		return m.viewEnvironmentEditor()
	}

	return ""
}

func (m Model) viewRequestBuilder() string {
	var b strings.Builder

	title := "GoDev v0.4.0"
	if m.requestSaved {
		title += " [SAVED]"
	}
	if m.envConfig != nil && m.envConfig.ActiveEnvironment != "" {
		title += fmt.Sprintf(" [ENV: %s]", m.envConfig.ActiveEnvironment)
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

	b.WriteString("\n")

	if m.curlCopySuccess {
		b.WriteString(SuccessStyle.Render("✓ cURL command copied to clipboard!"))
		b.WriteString("\n")
	}

	b.WriteString("\n")
	b.WriteString(RenderFooter("Ctrl+H: help • Ctrl+Enter: send • Ctrl+L: load • Ctrl+R: history • Ctrl+D: database • Ctrl+E: env • h: headers • b: body • q: query • s: save • x: cURL"))

	return Center(m.width, m.height, b.String())
}

func (m Model) viewLoading() string {
	var b strings.Builder

	if m.dbClient != nil && m.dbClient.IsConnected() && m.dbQueryEditor.Value() != "" {
		b.WriteString(TitleStyle.Render("Executing Query"))
		b.WriteString("\n\n")

		query := m.dbQueryEditor.Value()
		queryPreview := query
		if len(queryPreview) > 100 {
			queryPreview = queryPreview[:100] + "..."
		}
		b.WriteString(MutedStyle.Render(queryPreview))
		b.WriteString("\n\n")

		loadingBox := lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color(ColorAccent)).
			Padding(2, 4).
			Render(SpinnerStyle.Render(m.spinner.View()) + "  " + TextStyle.Render("Executing query..."))

		b.WriteString(loadingBox)
		b.WriteString("\n\n")
		b.WriteString(MutedStyle.Render("Please wait while the database processes your query"))
	} else if m.dbClient != nil && m.dbQueryEditor.Value() == "" {
		b.WriteString(TitleStyle.Render("Connecting to Database"))
		b.WriteString("\n\n")

		connectionInfo := fmt.Sprintf("%s:%s/%s",
			m.dbConnectHostInput.Value(),
			m.dbConnectPortInput.Value(),
			m.dbConnectDatabaseInput.Value())
		b.WriteString(TextStyle.Render(connectionInfo))
		b.WriteString("\n\n")

		loadingBox := lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color(ColorAccent)).
			Padding(2, 4).
			Render(SpinnerStyle.Render(m.spinner.View()) + "  " + TextStyle.Render("Loading database schema..."))

		b.WriteString(loadingBox)
		b.WriteString("\n\n")
		b.WriteString(MutedStyle.Render("Fetching tables and database information"))
	} else {
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
	}

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

	if m.curlCopySuccess {
		b.WriteString(SuccessStyle.Render("✓ cURL command copied to clipboard!"))
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
	b.WriteString(RenderFooter("Esc: back • s: save • c: copy response • x: copy as cURL • h: toggle headers • ↑↓: scroll"))

	return Center(m.width, m.height, b.String())
}

func (m Model) viewRequestList() string {
	var b strings.Builder

	title := fmt.Sprintf("Saved Requests (%d)", len(m.savedRequests))
	b.WriteString(TitleStyle.Render(title))
	b.WriteString("\n\n")

	if m.searchActive || m.searchInput.Value() != "" {
		searchLabel := "Search: "
		b.WriteString(TextStyle.Render(searchLabel))
		b.WriteString("\n")

		inputView := m.searchInput.View()
		var styledInput string
		if m.searchActive {
			styledInput = lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(lipgloss.Color(ColorAccent)).
				Padding(0, 1).
				Width(m.searchInput.Width + 2).
				Render(inputView)
		} else {
			styledInput = lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(lipgloss.Color(ColorBorder)).
				Padding(0, 1).
				Width(m.searchInput.Width + 2).
				Render(inputView)
		}
		b.WriteString(styledInput)
		b.WriteString("\n\n")
	}

	displayList := m.savedRequests
	if m.filteredRequests != nil {
		displayList = m.filteredRequests
	}

	if len(displayList) == 0 {
		if m.searchInput.Value() != "" {
			b.WriteString(MutedStyle.Render("No matching requests"))
		} else {
			b.WriteString(MutedStyle.Render("No saved requests"))
		}
	} else {
		for i, req := range displayList {
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

	if m.confirmingDelete && len(displayList) > 0 && m.requestToDelete < len(displayList) {
		confirmMsg := fmt.Sprintf("⚠ Delete '%s'? Press 'y' to confirm, 'Esc' to cancel", displayList[m.requestToDelete].Name)
		b.WriteString(WarningStyle.Render(confirmMsg))
		b.WriteString("\n\n")
	}

	b.WriteString(RenderFooter("↑↓: navigate • /: search • Enter: load • d: delete • n: new • Esc: back"))

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
	b.WriteString(TextStyle.Render("  Ctrl+R        View request history"))
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

func (m Model) handleHistoryKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "ctrl+c", "ctrl+q":
		return m, tea.Quit

	case "esc":
		if m.confirmingClearHistory {
			m.confirmingClearHistory = false
			return m, nil
		}
		m.state = StateRequestBuilder
		return m, nil

	case "up", "k":
		if m.selectedHistoryIdx > 0 {
			m.selectedHistoryIdx--
		}
		return m, nil

	case "down", "j":
		if m.selectedHistoryIdx < len(m.history)-1 {
			m.selectedHistoryIdx++
		}
		return m, nil

	case "enter":
		if len(m.history) > 0 && m.selectedHistoryIdx < len(m.history) {
			exec := m.history[m.selectedHistoryIdx]
			m.method = exec.Method
			m.urlInput.SetValue(exec.URL)
			m.headers = exec.Headers
			m.body = exec.Body
			if exec.QueryParams != nil {
				m.queryParams = exec.QueryParams
			} else {
				m.queryParams = make(map[string]string)
			}
			m.state = StateRequestBuilder
			m.requestSaved = false
		}
		return m, nil

	case "d":
		if len(m.history) > 0 && m.selectedHistoryIdx < len(m.history) {
			exec := m.history[m.selectedHistoryIdx]
			if m.storage != nil {
				m.storage.DeleteHistoryItem(exec.ID)
				m.history = m.storage.GetHistory()
				if m.selectedHistoryIdx >= len(m.history) && m.selectedHistoryIdx > 0 {
					m.selectedHistoryIdx--
				}
			}
		}
		return m, nil

	case "c":
		if len(m.history) > 0 {
			if !m.confirmingClearHistory {
				m.confirmingClearHistory = true
				return m, nil
			}
		}
		return m, nil

	case "y":
		if m.confirmingClearHistory && m.storage != nil {
			m.storage.ClearHistory()
			m.history = m.storage.GetHistory()
			m.selectedHistoryIdx = 0
			m.confirmingClearHistory = false
			return m, nil
		}
		return m, nil
	}

	return m, nil
}

func (m Model) viewHistory() string {
	var b strings.Builder

	b.WriteString(TitleStyle.Render(fmt.Sprintf("Request History (%d)", len(m.history))))
	b.WriteString("\n\n")

	if len(m.history) == 0 {
		b.WriteString(MutedStyle.Render("No request history"))
		b.WriteString("\n\n")
		b.WriteString(TextStyle.Render("Execute some requests to see them here"))
	} else {
		maxLines := m.height - 15
		start := m.selectedHistoryIdx
		if start > len(m.history)-maxLines {
			start = len(m.history) - maxLines
		}
		if start < 0 {
			start = 0
		}
		end := start + maxLines
		if end > len(m.history) {
			end = len(m.history)
		}

		for i := start; i < end; i++ {
			exec := m.history[i]
			statusStyle := TextStyle
			statusText := "ERROR"

			if exec.Error == "" {
				statusStyle = GetStatusStyle(exec.StatusCode)
				statusText = exec.Status
			}

			timestamp := exec.Timestamp.Format("15:04:05")
			line := fmt.Sprintf("%s  %s  %s", timestamp, exec.Method, exec.URL)

			if i == m.selectedHistoryIdx {
				b.WriteString(ListItemSelectedStyle.Render("> " + line))
				b.WriteString("\n")
				b.WriteString(MutedStyle.Render(fmt.Sprintf("    %s • %dms", statusStyle.Render(statusText), exec.ResponseTime)))
			} else {
				b.WriteString(ListItemStyle.Render(line))
				b.WriteString("\n")
				b.WriteString(MutedStyle.Render(fmt.Sprintf("    %s • %dms", statusStyle.Render(statusText), exec.ResponseTime)))
			}
			b.WriteString("\n")
		}
	}

	b.WriteString("\n")

	if m.confirmingClearHistory {
		b.WriteString(WarningStyle.Render("⚠ Clear all history? Press 'y' to confirm, 'Esc' to cancel"))
		b.WriteString("\n\n")
	}

	b.WriteString(RenderFooter("↑↓: navigate • Enter: load • d: delete item • c: clear all • Esc: back"))

	return Center(m.width, m.height, b.String())
}

func (m Model) handleDatabaseKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "ctrl+c", "ctrl+q":
		return m, tea.Quit

	case "esc":
		if m.dbClient != nil && m.dbClient.IsConnected() {
			m.dbClient.Close()
		}
		m.state = StateRequestBuilder
		return m, nil

	case "c":
		m.state = StateDatabaseConnect
		m.dbConnectFocusIndex = 0
		m.updateDatabaseConnectFocus()
		return m, nil

	case "q":
		if m.dbClient != nil && m.dbClient.IsConnected() {
			m.state = StateDatabaseQueryEditor
			m.dbQueryEditor.Focus()
			return m, nil
		}
		return m, nil

	case "l":
		if m.dbClient != nil && m.dbClient.IsConnected() {
			m.state = StateDatabaseQueryList
			m.dbSelectedQueryIdx = 0
			return m, nil
		}
		return m, nil

	case "s", "t":
		if m.dbClient != nil && m.dbClient.IsConnected() {
			m.state = StateDatabaseSchema
			return m, nil
		}
		return m, nil

	case "h":
		if m.dbClient != nil && m.dbClient.IsConnected() {
			if m.dbStorage != nil {
				m.dbQueryHistory = m.dbStorage.GetQueryHistory()
			}
			m.state = StateDatabaseQueryHistory
			m.dbSelectedQueryHistoryIdx = 0
			m.dbConfirmingClearQueryHistory = false
			return m, nil
		}
		return m, nil

	case "d":
		if m.dbClient != nil && m.dbClient.IsConnected() {
			m.dbClient.Close()
			return m, nil
		}
		return m, nil
	}

	return m, nil
}

func (m Model) viewDatabase() string {
	var b strings.Builder

	b.WriteString(TitleStyle.Render("Database Explorer (PostgreSQL)"))
	b.WriteString("\n\n")

	if m.dbClient == nil || !m.dbClient.IsConnected() {
		b.WriteString(TextStyle.Render("Welcome to the Database Explorer!"))
		b.WriteString("\n\n")
		b.WriteString(MutedStyle.Render("Connect to a PostgreSQL database to start"))
		b.WriteString("\n\n")

		menuPanel := lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color(ColorAccent)).
			Padding(1, 2).
			Width(m.width - 10).
			Render(HeaderStyle.Render("Actions") + "\n\n" +
				ButtonActive.Render("[ c ] Connect to Database") + "\n\n" +
				MutedStyle.Render("Press 'c' to open the connection form"))

		b.WriteString(menuPanel)
		b.WriteString("\n\n")

		b.WriteString(MutedStyle.Render("Features: Execute SQL • Save Queries • Browse Tables • Query History"))
	} else {
		connectionInfo := m.dbClient.GetConnectionString()
		b.WriteString(SuccessStyle.Render("✓ Connected to: " + connectionInfo))
		b.WriteString("\n\n")

		menuPanel := lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color(ColorBorder)).
			Padding(1, 2).
			Width(m.width - 10).
			Render(HeaderStyle.Render("Menu") + "\n\n" +
				TextStyle.Render("  [q] Execute Query") + "\n" +
				TextStyle.Render("  [s] Schema Browser") + "\n" +
				TextStyle.Render("  [l] Saved Queries") + "\n" +
				TextStyle.Render("  [h] Query History") + "\n" +
				TextStyle.Render("  [d] Disconnect") + "\n")

		b.WriteString(menuPanel)
	}

	b.WriteString("\n\n")
	b.WriteString(RenderFooter("q: query • s: schema • l: saved queries • h: history • d: disconnect • Esc: back"))

	return Center(m.width, m.height, b.String())
}

func (m Model) handleDatabaseConnectKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg.String() {
	case "ctrl+c", "ctrl+q":
		return m, tea.Quit

	case "esc":
		m.state = StateDatabase
		m.dbConnectFocusIndex = 0
		m.dbConnectHostInput.Blur()
		m.dbConnectPortInput.Blur()
		m.dbConnectDatabaseInput.Blur()
		m.dbConnectUserInput.Blur()
		m.dbConnectPasswordInput.Blur()
		return m, nil

	case "tab":
		m.dbConnectFocusIndex++
		if m.dbConnectFocusIndex > 4 {
			m.dbConnectFocusIndex = 0
		}
		m.updateDatabaseConnectFocus()
		return m, nil

	case "shift+tab":
		m.dbConnectFocusIndex--
		if m.dbConnectFocusIndex < 0 {
			m.dbConnectFocusIndex = 4
		}
		m.updateDatabaseConnectFocus()
		return m, nil

	case "enter":
		host := strings.TrimSpace(m.dbConnectHostInput.Value())
		portStr := strings.TrimSpace(m.dbConnectPortInput.Value())
		dbname := strings.TrimSpace(m.dbConnectDatabaseInput.Value())
		user := strings.TrimSpace(m.dbConnectUserInput.Value())
		password := m.dbConnectPasswordInput.Value()

		if host == "" || portStr == "" || dbname == "" || user == "" {
			return m, nil
		}

		port := 5432
		fmt.Sscanf(portStr, "%d", &port)

		config := database.ConnectionConfig{
			Host:     host,
			Port:     port,
			Database: dbname,
			User:     user,
			Password: password,
			SSLMode:  "disable",
		}

		err := m.dbClient.Connect(config)
		if err != nil {
			m.err = err
			return m, nil
		}

		if m.dbStorage != nil {
			m.dbStorage.SaveConnection(config)
		}

		m.state = StateLoading
		m.loading = true
		m.err = nil
		return m, loadDatabaseSchemaCmd(m.dbClient)

	default:
		switch m.dbConnectFocusIndex {
		case 0:
			m.dbConnectHostInput, cmd = m.dbConnectHostInput.Update(msg)
		case 1:
			m.dbConnectPortInput, cmd = m.dbConnectPortInput.Update(msg)
		case 2:
			m.dbConnectDatabaseInput, cmd = m.dbConnectDatabaseInput.Update(msg)
		case 3:
			m.dbConnectUserInput, cmd = m.dbConnectUserInput.Update(msg)
		case 4:
			m.dbConnectPasswordInput, cmd = m.dbConnectPasswordInput.Update(msg)
		}
		return m, cmd
	}
}

func (m *Model) updateDatabaseConnectFocus() {
	m.dbConnectHostInput.Blur()
	m.dbConnectPortInput.Blur()
	m.dbConnectDatabaseInput.Blur()
	m.dbConnectUserInput.Blur()
	m.dbConnectPasswordInput.Blur()

	switch m.dbConnectFocusIndex {
	case 0:
		m.dbConnectHostInput.Focus()
	case 1:
		m.dbConnectPortInput.Focus()
	case 2:
		m.dbConnectDatabaseInput.Focus()
	case 3:
		m.dbConnectUserInput.Focus()
	case 4:
		m.dbConnectPasswordInput.Focus()
	}
}

func (m Model) viewDatabaseConnect() string {
	var b strings.Builder

	b.WriteString(TitleStyle.Render("Connect to PostgreSQL Database"))
	b.WriteString("\n\n")

	if m.err != nil {
		b.WriteString(ErrorStyle.Render(fmt.Sprintf("✗ Connection failed: %v", m.err)))
		b.WriteString("\n\n")
	}

	renderInput := func(label string, input textinput.Model, focused bool) string {
		var result strings.Builder
		result.WriteString(TextStyle.Render(label))
		result.WriteString("\n")

		inputView := input.View()
		var styledInput string
		if focused {
			styledInput = lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(lipgloss.Color(ColorAccent)).
				Padding(0, 1).
				Width(input.Width + 2).
				Render(inputView)
		} else {
			styledInput = lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(lipgloss.Color(ColorBorder)).
				Padding(0, 1).
				Width(input.Width + 2).
				Render(inputView)
		}
		result.WriteString(styledInput)
		result.WriteString("\n\n")
		return result.String()
	}

	b.WriteString(renderInput("Host:", m.dbConnectHostInput, m.dbConnectFocusIndex == 0))
	b.WriteString(renderInput("Port:", m.dbConnectPortInput, m.dbConnectFocusIndex == 1))
	b.WriteString(renderInput("Database:", m.dbConnectDatabaseInput, m.dbConnectFocusIndex == 2))
	b.WriteString(renderInput("User:", m.dbConnectUserInput, m.dbConnectFocusIndex == 3))
	b.WriteString(renderInput("Password:", m.dbConnectPasswordInput, m.dbConnectFocusIndex == 4))

	buttons := RenderButton("Connect (Enter)", true) + "  "
	buttons += RenderButton("Cancel (Esc)", false)
	b.WriteString(buttons)

	b.WriteString("\n\n")
	b.WriteString(RenderFooter("Tab: next field • Enter: connect • Esc: cancel"))

	return Center(m.width, m.height, b.String())
}

type databaseResultMsg database.QueryResult

func executeDatabaseQueryCmd(client *database.PostgresClient, query string) tea.Cmd {
	return func() tea.Msg {
		result := client.ExecuteQuery(query)
		return databaseResultMsg(result)
	}
}

func loadDatabaseSchemaCmd(client *database.PostgresClient) tea.Cmd {
	return func() tea.Msg {
		tables, err := client.GetTables()
		if err != nil {
			return databaseSchemaMsg([]string{})
		}
		return databaseSchemaMsg(tables)
	}
}

func (m Model) handleDatabaseQueryEditorKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg.String() {
	case "ctrl+c", "ctrl+q":
		return m, tea.Quit

	case "esc":
		m.state = StateDatabase
		m.dbQueryEditor.Blur()
		return m, nil

	case "ctrl+k":
		query := strings.TrimSpace(m.dbQueryEditor.Value())
		if query == "" {
			return m, nil
		}

		m.state = StateLoading
		m.loading = true

		return m, executeDatabaseQueryCmd(m.dbClient, query)

	case "ctrl+s":
		query := strings.TrimSpace(m.dbQueryEditor.Value())
		if query == "" || m.dbStorage == nil {
			return m, nil
		}

		name := fmt.Sprintf("Query %s", time.Now().Format("15:04:05"))
		if !m.dbStorage.QueryExists(name) {
			m.dbStorage.SaveQuery(name, query)
			m.dbSavedQueries = m.dbStorage.GetQueries()
			m.dbQuerySaveSuccess = true
			m.dbQuerySaveSuccessTimer = 3
		}
		return m, nil

	default:
		m.dbQueryEditor, cmd = m.dbQueryEditor.Update(msg)
		return m, cmd
	}
}

func (m Model) viewDatabaseQueryEditor() string {
	var b strings.Builder

	b.WriteString(TitleStyle.Render("SQL Query Editor"))
	b.WriteString("\n\n")

	connectionInfo := m.dbClient.GetConnectionString()
	b.WriteString(MutedStyle.Render("Connected to: " + connectionInfo))
	b.WriteString("\n\n")

	editorPanel := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color(ColorAccent)).
		Padding(1, 2).
		Width(m.width - 10).
		Render(m.dbQueryEditor.View())

	b.WriteString(editorPanel)
	b.WriteString("\n\n")

	buttons := RenderButton("Execute (Ctrl+K)", true) + "  "
	buttons += RenderButton("Save (Ctrl+S)", false) + "  "
	buttons += RenderButton("Back (Esc)", false)
	b.WriteString(buttons)

	if m.dbQuerySaveSuccess {
		b.WriteString("\n\n")
		b.WriteString(SuccessStyle.Render("✓ Query saved successfully"))
	}

	b.WriteString("\n\n")
	b.WriteString(RenderFooter("Ctrl+K: execute • Ctrl+S: save query • Esc: back"))

	return Center(m.width, m.height, b.String())
}

func (m Model) handleDatabaseResultKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "ctrl+c", "ctrl+q":
		return m, tea.Quit

	case "esc":
		m.state = StateDatabaseQueryEditor
		m.dbQueryEditor.Focus()
		return m, nil

	case "s":
		query := strings.TrimSpace(m.dbQueryEditor.Value())
		if query == "" || m.dbStorage == nil {
			return m, nil
		}

		name := fmt.Sprintf("Query %s", time.Now().Format("15:04:05"))
		if !m.dbStorage.QueryExists(name) {
			m.dbStorage.SaveQuery(name, query)
			m.dbSavedQueries = m.dbStorage.GetQueries()
			m.dbQuerySaveSuccess = true
			m.dbQuerySaveSuccessTimer = 3
		}
		return m, nil

	case "e":
		if m.dbQueryResult != nil && len(m.dbQueryResult.Columns) > 0 {
			m.state = StateDatabaseExport
			m.dbExportFormatIdx = 0
			m.dbExportTableName.SetValue("")
			m.dbExportTableName.Focus()
			return m, nil
		}
		return m, nil
	}

	return m, nil
}

func (m Model) viewDatabaseResult() string {
	var b strings.Builder

	b.WriteString(TitleStyle.Render("Query Result"))
	b.WriteString("\n\n")

	if m.dbQueryResult == nil {
		b.WriteString(MutedStyle.Render("No result"))
		return Center(m.width, m.height, b.String())
	}

	if m.dbQueryResult.Error != nil {
		errorPanel := lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color(ColorError)).
			Padding(1, 2).
			Width(m.width - 10).
			Render(ErrorStyle.Render(fmt.Sprintf("Error: %v", m.dbQueryResult.Error)))

		b.WriteString(errorPanel)
	} else {
		timeInfo := fmt.Sprintf("Execution time: %dms", m.dbQueryResult.ExecutionTime.Milliseconds())
		b.WriteString(MutedStyle.Render(timeInfo))
		b.WriteString("\n\n")

		if len(m.dbQueryResult.Columns) > 0 {
			maxRows := 20
			rowsToShow := m.dbQueryResult.Rows
			if len(rowsToShow) > maxRows {
				rowsToShow = rowsToShow[:maxRows]
			}

			tableRenderer := NewTableRenderer(m.dbQueryResult.Columns, rowsToShow, m.width-20)
			tableContent := tableRenderer.Render()

			resultPanel := lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(lipgloss.Color(ColorBorder)).
				Padding(1, 2).
				Width(m.width - 10).
				Render(tableContent)

			b.WriteString(resultPanel)
			b.WriteString("\n\n")

			summary := tableRenderer.RenderSummary(len(m.dbQueryResult.Rows), len(rowsToShow))
			b.WriteString(SuccessStyle.Render("✓ " + summary))
		} else {
			b.WriteString(SuccessStyle.Render("✓ Query executed successfully"))
			b.WriteString("\n\n")
			b.WriteString(TextStyle.Render(fmt.Sprintf("Rows affected: %d", m.dbQueryResult.RowsAffected)))
		}
	}

	if m.dbQuerySaveSuccess {
		b.WriteString("\n\n")
		b.WriteString(SuccessStyle.Render("✓ Query saved successfully"))
	}

	if m.dbExportSuccess {
		b.WriteString("\n\n")
		b.WriteString(SuccessStyle.Render(fmt.Sprintf("✓ Results exported to: %s", m.dbExportFilePath)))
	}

	b.WriteString("\n\n")
	b.WriteString(RenderFooter("s: save query • e: export results • Esc: back to editor"))

	return Center(m.width, m.height, b.String())
}

func (m Model) handleDatabaseQueryListKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "ctrl+c", "ctrl+q":
		return m, tea.Quit

	case "esc":
		m.state = StateDatabase
		return m, nil

	case "up", "k":
		if m.dbSelectedQueryIdx > 0 {
			m.dbSelectedQueryIdx--
		}
		return m, nil

	case "down", "j":
		if m.dbSelectedQueryIdx < len(m.dbSavedQueries)-1 {
			m.dbSelectedQueryIdx++
		}
		return m, nil

	case "enter":
		if len(m.dbSavedQueries) > 0 && m.dbSelectedQueryIdx < len(m.dbSavedQueries) {
			query := m.dbSavedQueries[m.dbSelectedQueryIdx]
			m.dbQueryEditor.SetValue(query.Query)
			m.state = StateDatabaseQueryEditor
			m.dbQueryEditor.Focus()
		}
		return m, nil

	case "d":
		if len(m.dbSavedQueries) > 0 && m.dbSelectedQueryIdx < len(m.dbSavedQueries) && m.dbStorage != nil {
			query := m.dbSavedQueries[m.dbSelectedQueryIdx]
			m.dbStorage.DeleteQuery(query.ID)
			m.dbSavedQueries = m.dbStorage.GetQueries()
			if m.dbSelectedQueryIdx >= len(m.dbSavedQueries) && m.dbSelectedQueryIdx > 0 {
				m.dbSelectedQueryIdx--
			}
		}
		return m, nil
	}

	return m, nil
}

func (m Model) viewDatabaseQueryList() string {
	var b strings.Builder

	b.WriteString(TitleStyle.Render(fmt.Sprintf("Saved Queries (%d)", len(m.dbSavedQueries))))
	b.WriteString("\n\n")

	if len(m.dbSavedQueries) == 0 {
		b.WriteString(MutedStyle.Render("No saved queries"))
		b.WriteString("\n\n")
		b.WriteString(TextStyle.Render("Save queries from the editor with Ctrl+S"))
	} else {
		for i, query := range m.dbSavedQueries {
			if i == m.dbSelectedQueryIdx {
				b.WriteString(ListItemSelectedStyle.Render("> " + query.Name))
				b.WriteString("\n")
				preview := query.Query
				if len(preview) > 80 {
					preview = preview[:80] + "..."
				}
				b.WriteString(MutedStyle.Render("    " + preview))
			} else {
				b.WriteString(ListItemStyle.Render(query.Name))
			}
			b.WriteString("\n")
		}
	}

	b.WriteString("\n\n")
	b.WriteString(RenderFooter("↑↓: navigate • Enter: load • d: delete • Esc: back"))

	return Center(m.width, m.height, b.String())
}

func (m Model) handleDatabaseSchemaKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "ctrl+c", "ctrl+q":
		return m, tea.Quit

	case "esc":
		m.state = StateDatabase
		return m, nil

	case "up", "k":
		if m.dbSelectedTableIdx > 0 {
			m.dbSelectedTableIdx--
			m.dbTableInfo = nil
		}
		return m, nil

	case "down", "j":
		if m.dbSelectedTableIdx < len(m.dbTables)-1 {
			m.dbSelectedTableIdx++
			m.dbTableInfo = nil
		}
		return m, nil

	case "enter":
		if len(m.dbTables) > 0 && m.dbSelectedTableIdx < len(m.dbTables) {
			tableName := m.dbTables[m.dbSelectedTableIdx]
			tableInfo, err := m.dbClient.GetTableInfo(tableName)
			if err == nil {
				m.dbTableInfo = tableInfo
			}
		}
		return m, nil

	case "q":
		m.state = StateDatabaseQueryEditor
		m.dbQueryEditor.Focus()
		return m, nil

	case "l":
		m.state = StateDatabaseQueryList
		m.dbSelectedQueryIdx = 0
		return m, nil
	}

	return m, nil
}

func (m Model) viewDatabaseSchema() string {
	var b strings.Builder

	connectionInfo := m.dbClient.GetConnectionString()
	b.WriteString(TitleStyle.Render("Database Schema"))
	b.WriteString("\n")
	b.WriteString(MutedStyle.Render(connectionInfo))
	b.WriteString("\n")

	if m.dbConnectSuccess {
		b.WriteString("\n")
		b.WriteString(SuccessStyle.Render("✓ Connected successfully to database"))
		b.WriteString("\n")
	}

	b.WriteString("\n")

	if len(m.dbTables) == 0 {
		b.WriteString(MutedStyle.Render("No tables found in this database"))
		b.WriteString("\n\n")
		b.WriteString(TextStyle.Render("Press 'q' to open query editor"))
	} else {
		b.WriteString(HeaderStyle.Render(fmt.Sprintf("Tables (%d)", len(m.dbTables))))
		b.WriteString("\n\n")

		maxTablesToShow := 15
		start := m.dbSelectedTableIdx
		if start > len(m.dbTables)-maxTablesToShow {
			start = len(m.dbTables) - maxTablesToShow
		}
		if start < 0 {
			start = 0
		}
		end := start + maxTablesToShow
		if end > len(m.dbTables) {
			end = len(m.dbTables)
		}

		for i := start; i < end; i++ {
			tableName := m.dbTables[i]
			if i == m.dbSelectedTableIdx {
				b.WriteString(ListItemSelectedStyle.Render("> " + tableName))
			} else {
				b.WriteString(ListItemStyle.Render(tableName))
			}
			b.WriteString("\n")
		}

		if m.dbTableInfo != nil {
			b.WriteString("\n")
			b.WriteString(HeaderStyle.Render(fmt.Sprintf("Table: %s", m.dbTableInfo.Name)))
			b.WriteString("\n\n")

			if len(m.dbTableInfo.Columns) > 0 {
				columnData := [][]string{}
				for _, col := range m.dbTableInfo.Columns {
					nullable := "NO"
					if col.Nullable {
						nullable = "YES"
					}
					columnData = append(columnData, []string{col.Name, col.Type, nullable})
				}

				tableRenderer := NewTableRenderer(
					[]string{"Column", "Type", "Nullable"},
					columnData,
					m.width-20,
				)
				b.WriteString(tableRenderer.Render())
			}
		}
	}

	b.WriteString("\n\n")
	b.WriteString(RenderFooter("↑↓: navigate • Enter: view columns • q: query editor • l: saved queries • Esc: back"))

	return Center(m.width, m.height, b.String())
}

func (m Model) handleDatabaseQueryHistoryKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "ctrl+c", "ctrl+q":
		return m, tea.Quit

	case "esc":
		m.dbConfirmingClearQueryHistory = false
		m.state = StateDatabase
		return m, nil

	case "up", "k":
		if m.dbSelectedQueryHistoryIdx > 0 {
			m.dbSelectedQueryHistoryIdx--
		}
		return m, nil

	case "down", "j":
		if m.dbSelectedQueryHistoryIdx < len(m.dbQueryHistory)-1 {
			m.dbSelectedQueryHistoryIdx++
		}
		return m, nil

	case "enter":
		if len(m.dbQueryHistory) > 0 && m.dbSelectedQueryHistoryIdx < len(m.dbQueryHistory) {
			execution := m.dbQueryHistory[m.dbSelectedQueryHistoryIdx]
			m.dbQueryEditor.SetValue(execution.Query)
			m.state = StateDatabaseQueryEditor
			m.dbQueryEditor.Focus()
			return m, nil
		}
		return m, nil

	case "d":
		if len(m.dbQueryHistory) > 0 && m.dbSelectedQueryHistoryIdx < len(m.dbQueryHistory) {
			execution := m.dbQueryHistory[m.dbSelectedQueryHistoryIdx]
			if m.dbStorage != nil {
				m.dbStorage.DeleteQueryHistoryItem(execution.ID)
				m.dbQueryHistory = m.dbStorage.GetQueryHistory()
				if m.dbSelectedQueryHistoryIdx >= len(m.dbQueryHistory) && len(m.dbQueryHistory) > 0 {
					m.dbSelectedQueryHistoryIdx = len(m.dbQueryHistory) - 1
				}
			}
		}
		return m, nil

	case "c":
		if !m.dbConfirmingClearQueryHistory {
			m.dbConfirmingClearQueryHistory = true
		}
		return m, nil

	case "y":
		if m.dbConfirmingClearQueryHistory && m.dbStorage != nil {
			m.dbStorage.ClearQueryHistory()
			m.dbQueryHistory = []database.QueryExecution{}
			m.dbSelectedQueryHistoryIdx = 0
			m.dbConfirmingClearQueryHistory = false
		}
		return m, nil
	}

	return m, nil
}

func (m Model) viewDatabaseQueryHistory() string {
	var b strings.Builder

	b.WriteString(TitleStyle.Render(fmt.Sprintf("Query History (%d)", len(m.dbQueryHistory))))
	b.WriteString("\n\n")

	if len(m.dbQueryHistory) == 0 {
		b.WriteString(MutedStyle.Render("No query history"))
		b.WriteString("\n\n")
		b.WriteString(TextStyle.Render("Execute some queries to see them here"))
	} else {
		maxLines := m.height - 15
		start := m.dbSelectedQueryHistoryIdx
		if start > len(m.dbQueryHistory)-maxLines {
			start = len(m.dbQueryHistory) - maxLines
		}
		if start < 0 {
			start = 0
		}
		end := start + maxLines
		if end > len(m.dbQueryHistory) {
			end = len(m.dbQueryHistory)
		}

		for i := start; i < end; i++ {
			exec := m.dbQueryHistory[i]

			statusStyle := SuccessStyle
			statusText := "SUCCESS"
			if exec.Error != "" {
				statusStyle = ErrorStyle
				statusText = "ERROR"
			}

			timestamp := exec.Timestamp.Format("15:04:05")
			queryPreview := exec.Query
			if len(queryPreview) > 60 {
				queryPreview = queryPreview[:60] + "..."
			}
			queryPreview = strings.ReplaceAll(queryPreview, "\n", " ")

			line := fmt.Sprintf("%s  %s", timestamp, queryPreview)

			if i == m.dbSelectedQueryHistoryIdx {
				b.WriteString(ListItemSelectedStyle.Render("> " + line))
				b.WriteString("\n")

				info := fmt.Sprintf("    %s", statusStyle.Render(statusText))
				if exec.Error == "" {
					info += fmt.Sprintf(" • %dms • %d rows", exec.ExecutionTime, exec.RowsAffected)
				} else {
					info += fmt.Sprintf(" • %s", exec.Error)
				}
				b.WriteString(MutedStyle.Render(info))
			} else {
				b.WriteString(ListItemStyle.Render(line))
				b.WriteString("\n")
				info := fmt.Sprintf("    %s • %dms", statusStyle.Render(statusText), exec.ExecutionTime)
				b.WriteString(MutedStyle.Render(info))
			}
			b.WriteString("\n")
		}
	}

	b.WriteString("\n")

	if m.dbConfirmingClearQueryHistory {
		b.WriteString(WarningStyle.Render("⚠ Clear all history? Press 'y' to confirm, 'Esc' to cancel"))
		b.WriteString("\n\n")
	}

	b.WriteString(RenderFooter("↑↓: navigate • Enter: load • d: delete item • c: clear all • Esc: back"))

	return Center(m.width, m.height, b.String())
}

func (m Model) handleDatabaseExportKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg.String() {
	case "ctrl+c", "ctrl+q":
		return m, tea.Quit

	case "esc":
		m.state = StateDatabaseResult
		m.dbExportTableName.Blur()
		return m, nil

	case "up", "k":
		if m.dbExportFormatIdx > 0 {
			m.dbExportFormatIdx--
		}
		return m, nil

	case "down", "j":
		if m.dbExportFormatIdx < 2 {
			m.dbExportFormatIdx++
		}
		return m, nil

	case "tab", "shift+tab":
		m.dbExportTableName.Focus()
		return m, nil

	case "enter":
		formats := []database.ExportFormat{
			database.ExportFormatCSV,
			database.ExportFormatJSON,
			database.ExportFormatSQL,
		}

		format := formats[m.dbExportFormatIdx]
		tableName := strings.TrimSpace(m.dbExportTableName.Value())

		if format == database.ExportFormatSQL && tableName == "" {
			tableName = "exported_table"
		}

		result := database.ExportQueryResult(m.dbQueryResult, format, tableName)

		if result.Error != nil {
			m.err = result.Error
			return m, nil
		}

		m.dbExportFilePath = result.FilePath
		m.dbExportSuccess = true
		m.dbExportSuccessTimer = 5
		m.state = StateDatabaseResult
		m.dbExportTableName.Blur()

		return m, nil

	default:
		if m.dbExportTableName.Focused() {
			m.dbExportTableName, cmd = m.dbExportTableName.Update(msg)
			return m, cmd
		}
		return m, nil
	}
}

func (m Model) viewDatabaseExport() string {
	var b strings.Builder

	b.WriteString(TitleStyle.Render("Export Query Results"))
	b.WriteString("\n\n")

	b.WriteString(HeaderStyle.Render("Select Export Format"))
	b.WriteString("\n\n")

	formats := []string{
		"CSV (Comma-Separated Values)",
		"JSON (JavaScript Object Notation)",
		"SQL (INSERT Statements)",
	}

	for i, format := range formats {
		if i == m.dbExportFormatIdx {
			b.WriteString(ListItemSelectedStyle.Render("> " + format))
		} else {
			b.WriteString(ListItemStyle.Render(format))
		}
		b.WriteString("\n")
	}

	b.WriteString("\n")
	b.WriteString(HeaderStyle.Render("Table Name (for SQL export)"))
	b.WriteString("\n\n")

	tableNameBox := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color(ColorAccent)).
		Padding(0, 1).
		Width(m.width - 10).
		Render(m.dbExportTableName.View())

	b.WriteString(tableNameBox)
	b.WriteString("\n\n")

	info := fmt.Sprintf("Exporting %d rows", len(m.dbQueryResult.Rows))
	b.WriteString(MutedStyle.Render(info))

	b.WriteString("\n\n")
	b.WriteString(RenderFooter("↑↓: select format • Tab: edit table name • Enter: export • Esc: cancel"))

	return Center(m.width, m.height, b.String())
}

func (m Model) handleHomeKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "ctrl+c", "ctrl+q", "q":
		return m, tea.Quit

	case "1", "a":
		m.state = StateRequestBuilder
		m.urlInput.Focus()
		return m, nil

	case "2", "d":
		m.state = StateDatabase
		return m, nil

	case "?", "f1":
		m.state = StateHelp
		return m, nil
	}

	return m, nil
}

func (m Model) viewEnvironments() string {
	var b strings.Builder

	activeEnv := ""
	if m.envConfig != nil && m.envConfig.ActiveEnvironment != "" {
		activeEnv = m.envConfig.ActiveEnvironment
	}

	title := fmt.Sprintf("Environment Variables (%d)", len(m.envList))
	if activeEnv != "" {
		title += fmt.Sprintf(" | Active: %s", activeEnv)
	}
	b.WriteString(TitleStyle.Render(title))
	b.WriteString("\n\n")

	if m.envSaveSuccess {
		b.WriteString(SuccessStyle.Render("✓ Environment saved/activated successfully!"))
		b.WriteString("\n\n")
	}

	if m.envDeleteSuccess {
		b.WriteString(SuccessStyle.Render("✓ Environment deleted successfully!"))
		b.WriteString("\n\n")
	}

	if len(m.envList) == 0 {
		b.WriteString(MutedStyle.Render("No environments found"))
		b.WriteString("\n\n")
		b.WriteString(TextStyle.Render("Press 'n' to create your first environment (e.g., dev, staging, prod)"))
	} else {
		for i, env := range m.envList {
			prefix := "  "
			if i == m.selectedEnvIdx {
				prefix = "> "
			}

			envName := env.Name
			if activeEnv == env.Name {
				envName += " ★"
			}

			varCount := fmt.Sprintf("(%d vars)", len(env.Variables))

			if i == m.selectedEnvIdx {
				b.WriteString(ListItemSelectedStyle.Render(prefix + envName))
				b.WriteString("  ")
				b.WriteString(MutedStyle.Render(varCount))
			} else {
				b.WriteString(ListItemStyle.Render(prefix + envName))
				b.WriteString("  ")
				b.WriteString(MutedStyle.Render(varCount))
			}
			b.WriteString("\n")
		}
	}

	b.WriteString("\n\n")

	if m.confirmingDeleteEnv && len(m.envList) > 0 && m.selectedEnvIdx < len(m.envList) {
		confirmMsg := fmt.Sprintf("⚠ Delete environment '%s'? Press 'y' to confirm, 'Esc' to cancel", m.envList[m.selectedEnvIdx].Name)
		b.WriteString(WarningStyle.Render(confirmMsg))
		b.WriteString("\n\n")
	}

	b.WriteString(RenderFooter("↑↓: navigate • Enter: edit • n: new • s: set active • d: delete • Esc: back"))

	return Center(m.width, m.height, b.String())
}

func (m Model) viewEnvironmentEditor() string {
	var b strings.Builder

	if m.currentEnvName == "" {
		b.WriteString(TitleStyle.Render("New Environment"))
	} else {
		b.WriteString(TitleStyle.Render(fmt.Sprintf("Environment: %s", m.currentEnvName)))
	}
	b.WriteString("\n\n")

	if m.envSaveSuccess {
		b.WriteString(SuccessStyle.Render("✓ Saved successfully!"))
		b.WriteString("\n\n")
	}

	if m.envDeleteSuccess {
		b.WriteString(SuccessStyle.Render("✓ Variable deleted successfully!"))
		b.WriteString("\n\n")
	}

	if m.currentEnvName == "" {
		b.WriteString(HeaderStyle.Render("Environment Name:"))
		b.WriteString("\n")
		inputView := m.envNameInput.View()
		styledInput := lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color(ColorAccent)).
			Padding(0, 1).
			Width(m.envNameInput.Width + 2).
			Render(inputView)
		b.WriteString(styledInput)
		b.WriteString("\n\n")
		b.WriteString(MutedStyle.Render("Press Ctrl+S to save environment"))
		b.WriteString("\n\n")
	} else {
		b.WriteString(HeaderStyle.Render(fmt.Sprintf("Variables (%d):", len(m.envVarList))))
		b.WriteString("\n\n")

		if len(m.envVarList) == 0 {
			b.WriteString(MutedStyle.Render("No variables yet"))
			b.WriteString("\n\n")
			b.WriteString(TextStyle.Render("Press 'n' to add a variable (e.g., API_URL, API_KEY)"))
		} else {
			for i, variable := range m.envVarList {
				prefix := "  "
				if i == m.selectedEnvVarIdx {
					prefix = "> "
				}

				varText := fmt.Sprintf("%s = %s", variable.Key, variable.Value)

				if i == m.selectedEnvVarIdx {
					b.WriteString(ListItemSelectedStyle.Render(prefix + varText))
				} else {
					b.WriteString(ListItemStyle.Render(prefix + varText))
				}
				b.WriteString("\n")
			}
		}
	}

	b.WriteString("\n\n")

	if m.editingEnvVar {
		b.WriteString(HeaderStyle.Render("Add/Edit Variable:"))
		b.WriteString("\n\n")

		b.WriteString(TextStyle.Render("Key: "))
		b.WriteString("\n")
		keyInput := m.envVarKeyInput.View()
		keyStyle := lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			Padding(0, 1).
			Width(m.envVarKeyInput.Width + 2)
		if m.envFocusIndex == 0 {
			keyStyle = keyStyle.BorderForeground(lipgloss.Color(ColorAccent))
		} else {
			keyStyle = keyStyle.BorderForeground(lipgloss.Color(ColorBorder))
		}
		b.WriteString(keyStyle.Render(keyInput))
		b.WriteString("\n\n")

		b.WriteString(TextStyle.Render("Value: "))
		b.WriteString("\n")
		valueInput := m.envVarValueInput.View()
		valueStyle := lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			Padding(0, 1).
			Width(m.envVarValueInput.Width + 2)
		if m.envFocusIndex == 1 {
			valueStyle = valueStyle.BorderForeground(lipgloss.Color(ColorAccent))
		} else {
			valueStyle = valueStyle.BorderForeground(lipgloss.Color(ColorBorder))
		}
		b.WriteString(valueStyle.Render(valueInput))
		b.WriteString("\n\n")

		b.WriteString(RenderFooter("Tab: next field • Enter: save • Esc: cancel"))
		return Center(m.width, m.height, b.String())
	}

	if m.confirmingDeleteEnvVar && len(m.envVarList) > 0 && m.selectedEnvVarIdx < len(m.envVarList) {
		confirmMsg := fmt.Sprintf("⚠ Delete variable '%s'? Press 'y' to confirm, 'Esc' to cancel", m.envVarList[m.selectedEnvVarIdx].Key)
		b.WriteString(WarningStyle.Render(confirmMsg))
		b.WriteString("\n\n")
	}

	if m.currentEnvName == "" {
		b.WriteString(RenderFooter("Ctrl+S: save environment • Esc: back"))
	} else {
		b.WriteString(RenderFooter("↑↓: navigate • n: add variable • e: edit • d: delete • Esc: back"))
	}

	return Center(m.width, m.height, b.String())
}

func (m Model) viewHome() string {
	var b strings.Builder

	b.WriteString(TitleStyle.Render("GODEV v0.4.0"))
	b.WriteString("\n")
	b.WriteString(MutedStyle.Render("Professional API Testing & Database Tool"))
	b.WriteString("\n\n\n")

	menuPanel := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color(ColorAccent)).
		Padding(2, 4).
		Width(m.width - 20).
		Render(
			HeaderStyle.Render("SELECT MODE") + "\n\n" +
				ButtonActive.Render("[ 1 ] API Testing (HTTP)") + "\n" +
				MutedStyle.Render("      Test REST APIs, GraphQL & WebSocket") + "\n\n" +
				ButtonActive.Render("[ 2 ] Database Explorer (SQL)") + "\n" +
				MutedStyle.Render("      PostgreSQL queries, schema browser & more") + "\n",
		)

	b.WriteString(menuPanel)
	b.WriteString("\n\n")

	featuresInfo := lipgloss.NewStyle().
		Foreground(lipgloss.Color(ColorMuted)).
		Render("Features: Environment Variables • cURL Import • Request Collections • Query History")

	b.WriteString(featuresInfo)
	b.WriteString("\n\n")
	b.WriteString(RenderFooter("1: API Mode • 2: Database Mode • ?: Help • Q: Quit"))

	return Center(m.width, m.height, b.String())
}
