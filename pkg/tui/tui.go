package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/google/generative-ai-go/genai"

	"console-ai/pkg/gemini"
	"console-ai/pkg/history"
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
	stream              *conversationStream
	currentResponse     *strings.Builder
	lastRendered        string
}

// conversationStream holds the channel for receiving messages from the Gemini API.
type conversationStream struct {
	ch chan tea.Msg
}

// InitialModel creates the initial state of the TUI.
func InitialModel() Model {
	ti := textinput.New()
	ti.Placeholder = "Ask the AI to do something..."
	ti.Focus()
	ti.Width = 80

	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))

	vp := viewport.New(80, 20)
	vp.Style = lipgloss.NewStyle().
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("62")).
		Padding(0, 1)

	return Model{
		TextInput:       ti,
		Spinner:         s,
		Viewport:        vp,
		currentResponse: &strings.Builder{},
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
	case tea.KeyMsg:
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
		m.stream = newConversationStream(m.Gemini, m.ConversationHistory, msg.input)
		return m, m.stream.waitForNextMsg()

	case ErrMsg:
		m.Loading = false
		m.currentResponse.WriteString(fmt.Sprintf("\nError: %v", msg))
		m.renderView()
		return m, nil

	case SuccessMsg:
		m.ConversationHistory = append(m.ConversationHistory, m.TextInput.Value(), string(msg))
		history.SaveHistory(m.ConversationHistory)
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

// View renders the entire UI.
func (m Model) View() string {
	header := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#FAFAFA")).
		Background(lipgloss.Color("#7D56F4")).
		Padding(0, 1).
		Render("AI Console Agent")

	var status string
	if m.Loading {
		status = m.Spinner.View() + " AI is working..."
	} else {
		status = "Ready. (ctrl+c to quit)"
	}

	return fmt.Sprintf(
		"%s\n%s\n%s\n%s",
		header,
		m.Viewport.View(),
		m.TextInput.View(),
		status,
	)
}

// renderView updates the viewport with the latest content.
func (m *Model) renderView() {
	newContent := m.currentResponse.String()
	if newContent != m.lastRendered {
		m.Viewport.SetContent(newContent)
		m.lastRendered = newContent
		m.Viewport.GotoBottom()
	}
}

// newConversationStream creates a new stream for handling the Gemini conversation.
func newConversationStream(geminiModel *genai.GenerativeModel, history []string, input string) *conversationStream {
	ch := make(chan tea.Msg)
	go func() {
		defer close(ch)
		reply, err := gemini.ContinueConversation(geminiModel, history, input, func(title, content string) {
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
