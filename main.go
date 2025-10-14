package main

import (
	"log"

	tea "github.com/charmbracelet/bubbletea"

	"console-ai/pkg/gemini"
	"console-ai/pkg/tui"
)

func main() {
	gemini, err := gemini.NewClient()
	if err != nil {
		log.Fatalf("Failed to create Gemini client: %v", err)
	}

	m := tui.InitialModel()
	m.Gemini = gemini

	p := tea.NewProgram(m)

	_, err = p.Run()
	if err != nil {
		log.Fatalf("Alas, there's been an error: %v", err)
	}
}
