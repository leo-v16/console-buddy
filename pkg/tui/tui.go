package tui

import (
	"strings"

	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/google/generative-ai-go/genai"

	"console-ai/pkg/commander"
	"console-ai/pkg/gemini"
	"console-ai/pkg/history"
)

type (
	// ErrMsg is a custom error message to be sent as a tea.Msg.
	ErrMsg error

	// SuccessMsg is a custom success message to be sent as a tea.Msg.
	SuccessMsg string

	// StreamMsg is a message that contains a single step in the model's thinking process.
	StreamMsg struct {
		Step Step
	}

	// StepMsg is a message that contains all the steps in the model's thinking process.
	StepMsg struct {
		Steps []Step
	}

	// finalMsg is a message that signals the end of the streaming.
	finalMsg struct{}
)

// Step represents a single step in the model's thinking process.
type Step struct {
	Title   string
	Content string
}

type Model struct {
	Viewport            viewport.Model
	TextInput           textinput.Model
	Spinner             spinner.Model
	Loading             bool
	Gemini              *genai.GenerativeModel
	ConversationHistory []string
	Output              string
	Steps               []Step
	stream              stream
}

func InitialModel() Model {
	ti := textinput.New()
	ti.Placeholder = "Ask me anything..."
	ti.Focus()
	ti.CharLimit = 156
	ti.Width = 80

	s := spinner.New()
	s.Spinner = spinner.Globe
	s.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))

	vp := viewport.New(80, 20)
	vp.Style = lipgloss.NewStyle().
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("63")).
		Padding(0, 1)

	return Model{
		TextInput: ti,
		Spinner:   s,
		Viewport:  vp,
	}
}

func (m Model) Init() tea.Cmd {
	return m.Spinner.Tick
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var (
		cmd  tea.Cmd
		cmds []tea.Cmd
	)

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyEnter:
			m.Loading = true
			m.Steps = []Step{}
			m.Output = ""
			return m, m.UnderstandAndExecute()
		case tea.KeyCtrlC, tea.KeyEsc:
			return m, tea.Quit
		}
	case ErrMsg:
		m.Loading = false
		m.Output = msg.Error()
		m.Viewport.SetContent(m.Output)
		return m, nil
	case SuccessMsg:
		m.Loading = false
		m.Output = string(msg)
		m.Viewport.SetContent(m.Output)
		return m, nil
	case StreamMsg:
		// append to the last step content
		if len(m.Steps) > 0 {
			m.Steps[len(m.Steps)-1].Content += msg.Step.Content
		} else {
			m.Steps = append(m.Steps, msg.Step)
		}
		m.Viewport.SetContent(m.renderSteps())
		return m, waitForStream(m.stream)
	case StepMsg:
		m.Steps = msg.Steps
		m.Viewport.SetContent(m.renderSteps())
		return m, nil
	case finalMsg:
		m.Loading = false
		return m, nil

	case spinner.TickMsg:
		var cmd tea.Cmd
		m.Spinner, cmd = m.Spinner.Update(msg)
		return m, cmd
	}

	m.TextInput, cmd = m.TextInput.Update(msg)
	cmds = append(cmds, cmd)

	m.Viewport, cmd = m.Viewport.Update(msg)
	cmds = append(cmds, cmd)

	return m, tea.Batch(cmds...)
}

func (m Model) View() string {
	// Define styles
	headerStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#FAFAFA")).
		Background(lipgloss.Color("#874BFD")).
		Padding(0, 1)

	inputStyle := lipgloss.NewStyle().
		Border(lipgloss.NormalBorder(), false, false, false, true).
		BorderForeground(lipgloss.Color("63")).
		Padding(0, 1)

	statusStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("201")).
		Background(lipgloss.Color("57")).
		Bold(true)

	var status string
	if m.Loading {
		status = statusStyle.Render(m.Spinner.View() + " Thinking...")
	} else {
		status = ""
	}

	// Compose UI sections
	header := headerStyle.Render("Console Buddy")
	input := inputStyle.Render(m.TextInput.View())

	return lipgloss.JoinVertical(lipgloss.Left,
		header,
		m.Viewport.View(),
		input,
		status,
		"(ctrl+c to quit)",
	) + "\n"
}

func (m *Model) UnderstandAndExecute() tea.Cmd {
	input := m.TextInput.Value()
	m.TextInput.Reset()

	// Check if input is a command
	allowedCommands := map[string]bool{
		"dir":     true,
		"echo":    true,
		"type":    true,
		"copy":    true,
		"del":     true,
		"cls":     true,
		"cd":      true,
		"analyze": true,
	}
	firstWord := strings.Fields(input)
	isCommand := false
	if len(firstWord) > 0 {
		baseCmd := strings.Trim(strings.ToLower(firstWord[0]), ".:;!`'\"")
		if allowedCommands[baseCmd] {
			isCommand = true
		}
	}

	if isCommand {
		if firstWord[0] == "analyze" {
			return func() tea.Msg {
				output, err := gemini.AnalyzeProject()
				if err != nil {
					return ErrMsg(err)
				}
				return SuccessMsg(output)
			}
		}
		// Run as command
		return func() tea.Msg {
			output, err := commander.ExecuteCommand(input)
			if err != nil {
				return ErrMsg(err)
			}
			// Save to conversation history
			m.ConversationHistory = append(m.ConversationHistory, "Command: "+input)
			m.ConversationHistory = append(m.ConversationHistory, "Output: "+output)
			history.SaveConversation(m.ConversationHistory)
			return SuccessMsg(output)
		}
	} else {
		// Treat as conversation
		ch := make(chan any)
		m.stream = ch
		go func() {
			defer close(ch)
			reply, err := gemini.ContinueConversation(m.Gemini, m.ConversationHistory, input, func(title, content string) {
				ch <- StreamMsg{Step: Step{Title: title, Content: content}}
			})
			if err != nil {
				ch <- ErrMsg(err)
				return
			}
			// Save to conversation history
			m.ConversationHistory = append(m.ConversationHistory, "User: "+input)
			m.ConversationHistory = append(m.ConversationHistory, "Assistant: "+reply)
			history.SaveConversation(m.ConversationHistory)
			ch <- SuccessMsg(reply)
			ch <- finalMsg{}
		}()
		return waitForStream(ch)
	}
}

func waitForStream(ch chan any) tea.Cmd {
	return func() tea.Msg {
		return <-ch
	}
}

func (m Model) renderSteps() string {
	var steps strings.Builder
	for _, step := range m.Steps {
		steps.WriteString(lipgloss.NewStyle().Bold(true).Render(step.Title))
		steps.WriteString("\n")
		steps.WriteString(step.Content)
		steps.WriteString("\n\n")
	}
	return steps.String()
}

type stream chan any