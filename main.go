package main

import (
	"fmt"
	"os"
	"strings"

	tea "github.com/charmbracelet/bubbletea"

	"console-ai/pkg/agent"
	"console-ai/pkg/config"
	"console-ai/pkg/gemini"
	"console-ai/pkg/history"
	"console-ai/pkg/logger"
	"console-ai/pkg/tui"
)

func main() {
	// Use hardcoded configuration - no config files created:
	// - API Key: AIzaSyC-gNO6yZPjN1XgS0k6ncidRMPeoQ72Z9U
	// - Model: gemini-2.5-flash
	// - History + Project Context: CB.hist (binary format, created in current working directory)
	cfg, err := config.GetConfig()
	if err != nil {
		fmt.Printf("Error getting config: %v\n", err)
		os.Exit(1)
	}

	// Initialize logging
	logLevel := parseLogLevel(cfg.Logging.Level)
	loggerConfig := &logger.Config{
		Level:      logLevel,
		Output:     os.Stdout,
		LogFile:    cfg.Logging.File,
		EnableFile: cfg.Logging.EnableFile,
		Prefix:     "[Console-AI] ",
	}
	if err := logger.Initialize(loggerConfig); err != nil {
		fmt.Printf("Failed to initialize logger: %v\n", err)
		os.Exit(1)
	}
	defer logger.Shutdown()

	logger.Info("Console AI starting up...")
	logger.Debug("Configuration loaded: Model=%s, HumorLevel=%d", cfg.ModelName, cfg.HumorLevel)

	geminiClient, err := gemini.NewClient(cfg.GeminiAPIKey, cfg.ModelName)
	if err != nil {
		logger.Fatal("Failed to create Gemini client: %v", err)
	}

	// Load existing session data from CB.hist
	sessionData, err := history.LoadSession(cfg.ConversationHistory)
	if err != nil {
		logger.Warn("Error loading session data: %v", err)
		sessionData = nil
	}

	var projectInfo *agent.ProjectInfo
	var conversationHistory []string
	
	if sessionData != nil {
		projectInfo = sessionData.ProjectInfo
		conversationHistory = sessionData.Conversations
		// Update humor level from session if available
		if sessionData.HumorLevel > 0 {
			cfg.HumorLevel = sessionData.HumorLevel
		}
		logger.Info("Loaded session: %d conversations, %d total sessions", len(conversationHistory), sessionData.TotalSessions)
		if projectInfo != nil {
			logger.Info("Project context loaded: %s (%s)", projectInfo.Language, projectInfo.Framework)
		}
	} else {
		conversationHistory = []string{}
	}

	// Auto-analyze project if enabled and no project context exists
	if cfg.Agent.AutoAnalyze && (sessionData == nil || sessionData.ProjectInfo == nil) {
		logger.Info("Auto-analyzing project structure...")
		cwd, err := os.Getwd()
		if err == nil {
			analyzer := agent.NewProjectAnalyzer(cwd)
			if newProjectInfo, err := analyzer.AnalyzeProject(); err == nil {
				projectInfo = newProjectInfo
				logger.Info("Project analyzed: %s (%s)", projectInfo.Language, projectInfo.Framework)
				// Save the new project info to session
				history.SaveSession(cfg.ConversationHistory, conversationHistory, projectInfo, cfg.HumorLevel)
			} else {
				logger.Warn("Failed to analyze project: %v", err)
			}
		}
	}

	m := tui.InitialModel(cfg)
	m.Gemini = geminiClient
	m.ConversationHistory = conversationHistory
	m.ProjectInfo = projectInfo

	logger.Info("Starting TUI interface...")
	p := tea.NewProgram(m)

	if _, err := p.Run(); err != nil {
		logger.Fatal("TUI interface error: %v", err)
	}

	logger.Info("Console AI shutting down...")
}

// parseLogLevel converts string log level to logger.LogLevel
func parseLogLevel(level string) logger.LogLevel {
	switch strings.ToUpper(level) {
	case "DEBUG":
		return logger.DEBUG
	case "INFO":
		return logger.INFO
	case "WARN", "WARNING":
		return logger.WARN
	case "ERROR":
		return logger.ERROR
	case "FATAL":
		return logger.FATAL
	default:
		return logger.INFO
	}
}
