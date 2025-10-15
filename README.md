# Console AI - Intelligent Project Agent

A powerful AI-powered console application that acts as your intelligent development companion. Console AI analyzes your project structure, understands your codebase, and provides context-aware assistance for development tasks.

## Features

### 🧠 Project Intelligence
- **Automatic Project Analysis**: Detects programming languages, frameworks, dependencies, and project structure
- **Context-Aware Assistance**: Provides relevant help based on your project type and tools
- **Multi-Language Support**: Works with Go, JavaScript/TypeScript, Python, Rust, and more

### 🛠️ Advanced Tools
- **Code Generation**: Generate functions, classes, tests, and configuration files
- **Project Operations**: Install dependencies, run tests, and build projects using the correct tools
- **File Management**: Create, read, update, and delete files with safety checks
- **Command Execution**: Execute shell commands with security allowlists

### 🤖 AI-Powered Features
- **Intelligent Conversations**: Powered by Google's Gemini AI
- **Tool Integration**: AI can directly interact with your development environment
- **Safety First**: Built-in safety checks for potentially dangerous operations
- **Contextual Memory**: Maintains conversation history for continuity

### 🎨 User Experience
- **Beautiful TUI**: Elegant terminal user interface built with Bubble Tea
- **Real-time Streaming**: See AI responses as they're generated
- **Comprehensive Logging**: Detailed logging for debugging and monitoring
- **Flexible Configuration**: JSON config with environment variable overrides

## Installation

### Prerequisites
- Go 1.19 or higher
- A Google AI API key (Gemini)

### Build from Source
```bash
git clone <repository-url>
cd console-ai
go mod download
go build -o console-ai .
```

### Run
```bash
./console-ai
```

## Configuration

Console AI is **zero-configuration** - everything is hardcoded for immediate use:
- **Model**: `gemini-2.5-flash` (latest model)
- **API Key**: `AIzaSyC-gNO6yZPjN1XgS0k6ncidRMPeoQ72Z9U` (hardcoded and ready to use)
- **Storage**: `CB.hist` (binary format, stores conversation + project context)

**No config files are created or needed.** The application works out of the box with sensible defaults.

### Environment Variables

You can override configuration values using environment variables:

| Variable | Description |
|----------|-------------|
| `GEMINI_API_KEY` or `GOOGLE_API_KEY` | Your Google AI API key |
| `CONSOLE_AI_MODEL` | AI model to use |
| `CONSOLE_AI_HUMOR_LEVEL` | Humor level (0-100) |
| `CONSOLE_AI_LOG_LEVEL` | Logging level (DEBUG, INFO, WARN, ERROR, FATAL) |
| `CONSOLE_AI_LOG_FILE` | Log file path |
| `CONSOLE_AI_LOG_ENABLE_FILE` | Enable file logging (true/false) |
| `CONSOLE_AI_AUTO_ANALYZE` | Auto-analyze projects (true/false) |
| `CONSOLE_AI_CONTEXTUAL_HELP` | Enable contextual help (true/false) |
| `CONSOLE_AI_CODE_GENERATION` | Enable code generation (true/false) |
| `CONSOLE_AI_SAFETY_MODE` | Enable safety mode (true/false) |
| `CONSOLE_AI_ALLOWED_COMMANDS` | Comma-separated list of allowed commands |

### API Key

**Ready to Use**: The application comes with a working API key: `AIzaSyC-gNO6yZPjN1XgS0k6ncidRMPeoQ72Z9U`

The API key is hardcoded and ready to use immediately. If needed, you can override it with environment variables:
```bash
set GEMINI_API_KEY=YOUR_CUSTOM_API_KEY
```

### Smart Session Management

Console AI automatically manages everything in a single `CB.hist` file per directory:

**What's Stored in CB.hist:**
- 💬 **Conversation History**: All your chat messages
- 🔍 **Project Context**: Detected language, framework, dependencies
- 📊 **Session Stats**: Number of sessions, preferences
- ⚙️ **Settings**: Humor level, project-specific configurations

**Benefits:**
- 🗂️ **Per-Project Context**: Each directory gets its own intelligent context
- 🧠 **Smart Memory**: AI remembers your project details across sessions
- 📱 **Zero Setup**: No configuration files to manage
- 🔒 **Binary Format**: Efficient and secure storage

**Example:**
```bash
# Working in a Go project
cd /my/go-project
console-ai.exe  # Detects Go, creates CB.hist with Go context

# Working in a React project
cd /my/react-app  
console-ai.exe  # Detects React, creates CB.hist with React context
```

The AI automatically knows what type of project you're working on!

## Usage

### Basic Usage

1. Start the application: `./console-ai`
2. Type your question or request in the input field
3. Press Enter to send
4. The AI will analyze your request and respond with relevant help

### Available Tools

The AI can use the following tools to help you:

#### Project Analysis
```
Ask: "Analyze this project"
```
- Detects programming language and framework
- Lists dependencies and scripts
- Identifies project structure

#### Code Generation
```
Ask: "Generate a function called calculateSum that takes two numbers"
Ask: "Create a React component for a user profile"
Ask: "Generate a test for the User class"
```

#### Project Operations
```
Ask: "Install the express package"
Ask: "Run the tests"
Ask: "Build the project"
```

#### File Operations
```
Ask: "Create a new file called utils.js with helper functions"
Ask: "Show me the contents of main.go"
Ask: "Update the package.json to add a new script"
```

### Keyboard Shortcuts

- `Enter`: Send message
- `Ctrl+C` or `Esc`: Quit
- `?`: Toggle help

## Project Structure

```
console-ai/
├── main.go                 # Application entry point
├── pkg/
│   ├── agent/             # Project analysis and code generation
│   │   ├── analyzer.go    # Project structure analysis
│   │   └── generator.go   # Code generation templates
│   ├── config/            # Configuration management
│   │   └── config.go      # Config loading and validation
│   ├── gemini/            # Gemini AI integration
│   │   ├── client.go      # AI client setup
│   │   ├── gemini.go      # Conversation handling
│   │   ├── tools.go       # AI tool definitions
│   │   └── constants.go   # System prompts
│   ├── history/           # Conversation persistence
│   │   └── history.go     # History management
│   ├── logger/            # Logging system
│   │   └── logger.go      # Structured logging
│   ├── commander/         # Command execution
│   │   └── commander.go   # Safe command runner
│   ├── tui/              # Terminal user interface
│   │   ├── tui.go        # Main TUI logic
│   │   └── help.go       # Help system
│   └── cat/              # Animation components
│       └── cat.go        # Loading animations
├── config.json           # Configuration file
├── go.mod               # Go module definition
├── go.sum               # Go module checksums
└── README.md           # This file
```

## Supported Languages and Frameworks

### Programming Languages
- **Go**: Full support with module analysis, testing, and building
- **JavaScript/TypeScript**: NPM, Yarn, PNPM support with framework detection
- **Python**: Pip, Poetry support with virtual environment awareness
- **Rust**: Cargo integration with dependency management

### Frameworks Detected
- **Frontend**: React, Vue.js, Angular
- **Backend**: Express.js, Django, Flask
- **Testing**: Jest, Mocha, PyTest, Go testing, Cargo test

## Development

### Contributing

1. Fork the repository
2. Create a feature branch: `git checkout -b feature-name`
3. Make your changes
4. Add tests if applicable
5. Commit your changes: `git commit -am 'Add feature'`
6. Push to the branch: `git push origin feature-name`
7. Submit a pull request

### Architecture

Console AI follows a modular architecture:

- **Agent Module**: Handles project analysis and code generation
- **Gemini Module**: Manages AI interactions and tool execution
- **TUI Module**: Provides the terminal user interface
- **Config Module**: Manages configuration and environment variables
- **Logger Module**: Provides structured logging with multiple outputs

### Adding New Tools

To add new AI tools:

1. Define the tool in `pkg/gemini/tools.go`
2. Add the implementation method to `ToolExecutor`
3. Update the tool dispatcher in the `Execute` method
4. Add tests for the new functionality

## Troubleshooting

### Common Issues

**API Key Not Working**
- Ensure your API key is valid and has Gemini API access
- Check that the key is properly set in config or environment

**Commands Not Executing**
- Verify the command is in the `allowed_commands` list
- Check that the required tools are installed (npm, go, python, etc.)

**Project Not Detected**
- Ensure you're running from the project root directory
- Check that standard project files exist (package.json, go.mod, etc.)

### Logging

Enable debug logging to troubleshoot issues:
```bash
export CONSOLE_AI_LOG_LEVEL=DEBUG
export CONSOLE_AI_LOG_ENABLE_FILE=true
./console-ai
```

Check the log file at `logs/console-ai.log` for detailed information.

## License

This project is licensed under the MIT License - see the LICENSE file for details.

## Acknowledgments

- [Bubble Tea](https://github.com/charmbracelet/bubbletea) - Excellent TUI framework
- [Google Generative AI](https://github.com/google/generative-ai-go) - AI integration
- [Lip Gloss](https://github.com/charmbracelet/lipgloss) - Terminal styling

## Roadmap

- [ ] Plugin system for custom tools
- [ ] Integration with more AI providers
- [ ] Web interface option
- [ ] Docker support
- [ ] CI/CD integration helpers
- [ ] Code review and analysis features
- [ ] Multi-project workspace support