package tui

import (
	"fmt"
	"strings"

	"console-ai/pkg/agent"
	"console-ai/pkg/config"
	"console-ai/pkg/gemini"
	"console-ai/pkg/history"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/google/generative-ai-go/genai"
)

type (
	ErrMsg               error
	SuccessMsg           string
	StreamMsg            struct{ Title, Content string }
	startConversationMsg struct{ input string }
	finalMsg             struct{}
)

// Model represents the state of the TUI application.
type Model struct {
	Viewport            viewport.Model
	TextInput           textinput.Model
	Spinner             spinner.Model
	Loading             bool
	Gemini              *genai.GenerativeModel
	ConversationHistory []string
	ProjectInfo         *agent.ProjectInfo
	stream              *conversationStream
	currentResponse     *strings.Builder
	lastRendered        string
	Config              *config.Config
	Help                help.Model
	Keys                *helpKeyMap
	width               int
	height              int
}

// conversationStream holds the channel for receiving messages from the Gemini API.
type conversationStream struct {
	ch chan tea.Msg
}

// InitialModel creates the initial state of the TUI.
func InitialModel(cfg *config.Config) Model {
	ti := textinput.New()
	ti.Placeholder = "Ask the AI to do something..."
	ti.Focus()
	ti.CharLimit = 0 // No limit

	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))

	// Start with reasonable defaults, will be updated on first resize
	vp := viewport.New(100, 20)
	vp.Style = lipgloss.NewStyle().
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("62")).
		Padding(0, 1)

	keys := newHelpKeyMap()
	h := newHelp(keys)

	return Model{
		TextInput:       ti,
		Spinner:         s,
		Viewport:        vp,
		currentResponse: &strings.Builder{},
		Config:          cfg,
		Help:            h,
		Keys:            keys,
		width:           100,
		height:          24,
	}
}

// Init initializes the TUI.
func (m Model) Init() tea.Cmd {
	return m.Spinner.Tick
}

// Update handles all incoming messages and updates the model accordingly.
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.updateSizes()
		return m, nil
		
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, m.Keys.help):
			m.Help.ShowAll = !m.Help.ShowAll
			return m, nil
		case key.Matches(msg, m.Keys.quit):
			return m, tea.Quit
		}

		switch msg.Type {
		case tea.KeyEnter:
			if m.Loading {
				return m, nil
			}
			m.Loading = true
			m.currentResponse.Reset()
			m.lastRendered = ""
			return m, func() tea.Msg {
				return startConversationMsg{input: m.TextInput.Value()}
			}
		case tea.KeyCtrlC, tea.KeyEsc:
			return m, tea.Quit
		}

	case startConversationMsg:
		m.stream = newConversationStream(m.Gemini, m.ConversationHistory, msg.input, m.Config.HumorLevel, m.Config)
		return m, m.stream.waitForNextMsg()

	case ErrMsg:
		m.Loading = false
		m.currentResponse.WriteString(fmt.Sprintf("\nError: %v", msg))
		m.renderView()
		return m, nil

	case SuccessMsg:
		m.ConversationHistory = append(m.ConversationHistory, m.TextInput.Value(), string(msg))
		// Save session data with project context
		history.SaveSession(m.Config.ConversationHistory, m.ConversationHistory, m.ProjectInfo, m.Config.HumorLevel)
		m.TextInput.Reset()
		return m, m.stream.waitForNextMsg()

	case StreamMsg:
		m.currentResponse.WriteString(msg.Content)
		m.renderView()
		return m, m.stream.waitForNextMsg()

	case finalMsg:
		m.Loading = false
		m.TextInput.Focus()
		return m, textinput.Blink

	case spinner.TickMsg:
		var cmd tea.Cmd
		m.Spinner, cmd = m.Spinner.Update(msg)
		cmds = append(cmds, cmd)
	}

	var cmd tea.Cmd
	m.TextInput, cmd = m.TextInput.Update(msg)
	cmds = append(cmds, cmd)
	m.Viewport, cmd = m.Viewport.Update(msg)
	cmds = append(cmds, cmd)

	return m, tea.Batch(cmds...)
}

// updateSizes updates component sizes based on terminal dimensions
func (m *Model) updateSizes() {
	// Calculate available space
	headerHeight := 1
	statusHeight := 1
	helpHeight := 2
	inputHeight := 1
	padding := 2
	
	// Update text input width
	inputWidth := m.width - 4 // Account for borders and padding
	if inputWidth < 20 {
		inputWidth = 20
	}
	m.TextInput.Width = inputWidth
	
	// Update viewport dimensions
	viewportHeight := m.height - headerHeight - statusHeight - helpHeight - inputHeight - padding
	if viewportHeight < 5 {
		viewportHeight = 5
	}
	
	viewportWidth := m.width - 4 // Account for borders and padding
	if viewportWidth < 20 {
		viewportWidth = 20
	}
	
	m.Viewport.Width = viewportWidth
	m.Viewport.Height = viewportHeight
}

// View renders the entire UI.
func (m Model) View() string {
	header := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#FAFAFA")).
		Background(lipgloss.Color("#7D56F4")).
		Padding(0, 1).
		Width(m.width-2).
		Align(lipgloss.Center).
		Render("Console Buddy")

	statusText := "Ready. (? for help)"
	if m.Loading {
		statusText = m.Spinner.View() + " AI is working..."
	}

	projectStatus := ""
	if m.ProjectInfo != nil {
		projectStatus = fmt.Sprintf(" | %s", m.ProjectInfo.Language)
		if m.ProjectInfo.Framework != "" && len(m.ProjectInfo.Framework) < 20 {
			projectStatus += fmt.Sprintf(" (%s)", m.ProjectInfo.Framework)
		}
	}
	
	// Create status text and truncate if too long
	statusFullText := fmt.Sprintf("%s | Model: %s%s", statusText, m.Config.ModelName, projectStatus)
	if len(statusFullText) > m.width-4 {
		// Truncate to fit
		statusFullText = statusFullText[:m.width-7] + "..."
	}
	
	statusBar := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#FFF")).
		Background(lipgloss.Color("#5C5C5C")).
		Padding(0, 1).
		Width(m.width-2).
		Render(statusFullText)

	helpView := m.Help.View(m.Keys)
	// Ensure help doesn't overflow
	if len(helpView) > m.width {
		helpLines := strings.Split(helpView, "\n")
		var truncatedLines []string
		for _, line := range helpLines {
			if len(line) > m.width-2 {
				line = line[:m.width-5] + "..."
			}
			truncatedLines = append(truncatedLines, line)
		}
		helpView = strings.Join(truncatedLines, "\n")
	}

	return fmt.Sprintf(
		"%s\n%s\n%s\n%s\n%s",
		header,
		m.Viewport.View(),
		m.TextInput.View(),
		statusBar,
		helpView,
	)
}

// renderView updates the viewport with the latest content.
func (m *Model) renderView() {
	newContent := m.currentResponse.String()
	if newContent != m.lastRendered {
		// Wrap long lines to prevent overflow
		wrappedContent := m.wrapText(newContent, m.Viewport.Width-4)
		m.Viewport.SetContent(wrappedContent)
		m.lastRendered = newContent
		m.Viewport.GotoBottom()
	}
}

// wrapText wraps text to fit within the specified width
func (m *Model) wrapText(text string, width int) string {
	if width <= 0 {
		width = 80 // fallback width
	}
	
	lines := strings.Split(text, "\n")
	var wrappedLines []string
	
	for _, line := range lines {
		if len(line) <= width {
			wrappedLines = append(wrappedLines, line)
		} else {
			// Break long lines into multiple lines
			for len(line) > width {
				// Try to break at word boundaries
				breakPoint := width
				if spaceIdx := strings.LastIndex(line[:width], " "); spaceIdx > width/2 {
					breakPoint = spaceIdx
				}
				
				wrappedLines = append(wrappedLines, line[:breakPoint])
				line = strings.TrimSpace(line[breakPoint:])
			}
			if len(line) > 0 {
				wrappedLines = append(wrappedLines, line)
			}
		}
	}
	
	return strings.Join(wrappedLines, "\n")
}

// newConversationStream creates a new stream for handling the Gemini conversation.
func newConversationStream(geminiModel *genai.GenerativeModel, history []string, input string, humorLevel int, cfg *config.Config) *conversationStream {
	ch := make(chan tea.Msg)
	go func() {
		defer close(ch)
		reply, err := gemini.ContinueConversation(geminiModel, history, input, humorLevel, cfg, func(title, content string) {
			ch <- StreamMsg{Title: title, Content: content}
		})

		if err != nil {
			ch <- ErrMsg(err)
			return
		}

		ch <- SuccessMsg(reply)
		ch <- finalMsg{}
	}()
	return &conversationStream{ch: ch}
}

// waitForNextMsg waits for the next message from the conversation stream.
func (s *conversationStream) waitForNextMsg() tea.Cmd {
	return func() tea.Msg {
		return <-s.ch
	}
}
