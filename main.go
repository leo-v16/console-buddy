package main

import (
	"log"

	tea "github.com/charmbracelet/bubbletea"

	"console-ai/pkg/gemini"
	"console-ai/pkg/history"
	"console-ai/pkg/tui"
)

func main() {
	// Initialize the Gemini client first.
	geminiClient, err := gemini.NewClient()
	if err != nil {
		log.Fatalf("Failed to create Gemini client: %v", err)
	}

	// Create the TUI model.
	m := tui.InitialModel()

	// Set the Gemini client and load the conversation history.
	m.Gemini = geminiClient
	m.ConversationHistory = history.LoadHistory()

	// Start the Bubble Tea program.
	p := tea.NewProgram(m)

	if _, err := p.Run(); err != nil {
		log.Fatalf("Alas, there's been an error: %v", err)
	}
}
