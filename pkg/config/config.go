package config

import (
	"os"
	"strconv"
	"strings"
)

// Config holds the application's hardcoded configuration.
// No config file is generated - all values are hardcoded for simplicity.
type Config struct {
	GeminiAPIKey        string
	ConversationHistory string
	HumorLevel          int
	ModelName           string
	AllowedCommands     []string
	Logging             LogConfig
	Agent               AgentConfig
}

// LogConfig holds logging configuration
type LogConfig struct {
	Level      string // DEBUG, INFO, WARN, ERROR, FATAL
	File       string // Log file path
	EnableFile bool   // Whether to enable file logging
}

// AgentConfig holds agent-specific configuration
type AgentConfig struct {
	AutoAnalyze    bool // Automatically analyze project on startup
	ContextualHelp bool // Provide context-aware help
	CodeGeneration bool // Enable code generation features
	SafetyMode     bool // Enable safety checks for dangerous commands
}

// GetConfig returns the hardcoded configuration.
// All settings are hardcoded - no config file is created or read.
// Only environment variables can override settings.
func GetConfig() (*Config, error) {
	// Hardcoded configuration
	config := &Config{
		GeminiAPIKey:        "AIzaSyC-gNO6yZPjN1XgS0k6ncidRMPeoQ72Z9U", // Hardcoded API key
		ConversationHistory: "CB.hist",
		HumorLevel:          0,
		ModelName:           "gemini-2.5-flash",
		AllowedCommands: []string{
			// Programming Languages & Runtimes
			"go", "gofmt", "goimports", "python", "python3", "py", "node", "java", "javac",
			"ruby", "perl", "php", "rustc", "cargo", "dotnet", "lua",

			// Package Managers
			"npm", "npx", "yarn", "pnpm", "pip", "pip3", "gem", "composer", "bundle",
			"cargo", "poetry", "pipenv", "maven", "mvn", "gradle", "nuget",

			// Version Control
			"git", "gitk", "svn", "hg",

			// Build Tools
			"make", "cmake", "nmake", "msbuild", "ant", "webpack", "vite", "rollup",

			// Compilers & Interpreters
			"gcc", "g++", "clang", "clang++", "cl", "nvcc", "tsc", "babel",

			// Testing Tools
			"jest", "mocha", "pytest", "phpunit", "junit", "karma", "cypress",

			// Linters & Formatters
			"eslint", "prettier", "pylint", "black", "flake8", "rubocop", "phpstan",
			"golint", "rustfmt", "stylelint",

			// Database CLI Tools
			"mysql", "psql", "sqlite3", "mongo", "mongosh", "redis-cli",

			// Container & Orchestration
			"docker", "docker-compose", "kubectl", "podman", "vagrant",

			// Cloud Platform CLIs
			"aws", "az", "gcloud", "firebase", "heroku", "vercel", "netlify",

			// Windows Commands
			"dir", "type", "copy", "xcopy", "move", "del", "mkdir", "rmdir", "cd",
			"cls", "echo", "find", "findstr", "where", "tree", "attrib", "systeminfo",
			"ipconfig", "netstat", "ping", "tracert", "nslookup", "tasklist",

			// Unix/Linux Commands
			"ls", "cat", "grep", "cp", "mv", "rm", "pwd", "touch", "chmod", "head",
			"tail", "wc", "sort", "uniq", "diff", "sed", "awk", "tar", "gzip", "gunzip",
			"zip", "unzip", "curl", "wget", "ssh", "scp", "rsync",

			// Text Editors
			"vim", "nvim", "nano", "emacs", "code", "notepad",

			// Archive Tools
			"tar", "zip", "unzip", "7z", "gzip", "gunzip", "bzip2",

			// System Utilities
			"env", "printenv", "which", "whoami", "hostname", "date", "time",
			"clear", "history", "man", "help",

			// Package Managers (System Level)
			"choco", "scoop", "brew", "apt", "apt-get", "yum", "dnf",

			// Other Development Tools
			"jq", "base64", "openssl", "gpg", "nc", "telnet",
		},

		Logging: LogConfig{
			Level:      "INFO",
			File:       "logs/console-ai.log",
			EnableFile: false,
		},
		Agent: AgentConfig{
			AutoAnalyze:    true,
			ContextualHelp: true,
			CodeGeneration: true,
			SafetyMode:     true,
		},
	}

	// Override with environment variables if set
	if err := loadFromEnvironment(config); err != nil {
		return nil, err
	}

	return config, nil
}

// LoadConfig is kept for backward compatibility but just calls GetConfig
func LoadConfig(path string) (*Config, error) {
	return GetConfig()
}

// loadFromEnvironment loads configuration from environment variables
func loadFromEnvironment(config *Config) error {
	// Load API key from environment
	if apiKey := os.Getenv("GEMINI_API_KEY"); apiKey != "" {
		config.GeminiAPIKey = apiKey
	}
	if apiKey := os.Getenv("GOOGLE_API_KEY"); apiKey != "" {
		config.GeminiAPIKey = apiKey
	}

	// Load model name
	if modelName := os.Getenv("CONSOLE_AI_MODEL"); modelName != "" {
		config.ModelName = modelName
	}

	// Load humor level
	if humorStr := os.Getenv("CONSOLE_AI_HUMOR_LEVEL"); humorStr != "" {
		if humor, err := strconv.Atoi(humorStr); err == nil {
			config.HumorLevel = humor
		}
	}

	// Load logging configuration
	if logLevel := os.Getenv("CONSOLE_AI_LOG_LEVEL"); logLevel != "" {
		config.Logging.Level = strings.ToUpper(logLevel)
	}
	if logFile := os.Getenv("CONSOLE_AI_LOG_FILE"); logFile != "" {
		config.Logging.File = logFile
	}
	if enableFileStr := os.Getenv("CONSOLE_AI_LOG_ENABLE_FILE"); enableFileStr != "" {
		if enableFile, err := strconv.ParseBool(enableFileStr); err == nil {
			config.Logging.EnableFile = enableFile
		}
	}

	// Load agent configuration
	if autoAnalyzeStr := os.Getenv("CONSOLE_AI_AUTO_ANALYZE"); autoAnalyzeStr != "" {
		if autoAnalyze, err := strconv.ParseBool(autoAnalyzeStr); err == nil {
			config.Agent.AutoAnalyze = autoAnalyze
		}
	}
	if contextualHelpStr := os.Getenv("CONSOLE_AI_CONTEXTUAL_HELP"); contextualHelpStr != "" {
		if contextualHelp, err := strconv.ParseBool(contextualHelpStr); err == nil {
			config.Agent.ContextualHelp = contextualHelp
		}
	}
	if codeGenStr := os.Getenv("CONSOLE_AI_CODE_GENERATION"); codeGenStr != "" {
		if codeGen, err := strconv.ParseBool(codeGenStr); err == nil {
			config.Agent.CodeGeneration = codeGen
		}
	}
	if safetyModeStr := os.Getenv("CONSOLE_AI_SAFETY_MODE"); safetyModeStr != "" {
		if safetyMode, err := strconv.ParseBool(safetyModeStr); err == nil {
			config.Agent.SafetyMode = safetyMode
		}
	}

	// Load allowed commands
	if allowedCmds := os.Getenv("CONSOLE_AI_ALLOWED_COMMANDS"); allowedCmds != "" {
		config.AllowedCommands = strings.Split(allowedCmds, ",")
		for i, cmd := range config.AllowedCommands {
			config.AllowedCommands[i] = strings.TrimSpace(cmd)
		}
	}

	return nil
}
