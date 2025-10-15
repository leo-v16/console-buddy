package gemini

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"console-ai/pkg/agent"
	"console-ai/pkg/commander"
	"console-ai/pkg/config"
	"console-ai/pkg/logger"

	"github.com/google/generative-ai-go/genai"
)

// defineTools declares the functions the AI can execute.
func defineTools() []*genai.Tool {
	return []*genai.Tool{
		{
			FunctionDeclarations: []*genai.FunctionDeclaration{
				{
					Name:        "execute_shell_command",
					Description: "Executes a shell command on the user's machine. Use this for general-purpose commands that are not related to file manipulation. For example, 'go run main.go' or 'npm install'.",
					Parameters: &genai.Schema{
						Type: genai.TypeObject,
						Properties: map[string]*genai.Schema{
							"command": {Type: genai.TypeString, Description: "The command to execute."},
						},
						Required: []string{"command"},
					},
				},
				{
					Name:        "create_file",
					Description: "Creates a new file with the given content. For example, to create a new Python file, you would use create_file('main.py', 'print(\"Hello, World!\")').",
					Parameters: &genai.Schema{
						Type: genai.TypeObject,
						Properties: map[string]*genai.Schema{
							"path":    {Type: genai.TypeString, Description: "The path of the file to create."},
							"content": {Type: genai.TypeString, Description: "The content to write to the file."},
						},
						Required: []string{"path", "content"},
					},
				},
				{
					Name:        "read_file",
					Description: "Reads the content of a file. For example, to read a file named 'main.go', you would use read_file('main.go').",
					Parameters: &genai.Schema{
						Type: genai.TypeObject,
						Properties: map[string]*genai.Schema{
							"path": {Type: genai.TypeString, Description: "The path of the file to read."},
						},
						Required: []string{"path"},
					},
				},
				{
					Name:        "update_file",
					Description: "Updates the content of an existing file. This overwrites the entire file.",
					Parameters: &genai.Schema{
						Type: genai.TypeObject,
						Properties: map[string]*genai.Schema{
							"path":    {Type: genai.TypeString, Description: "The path of the file to update."},
							"content": {Type: genai.TypeString, Description: "The new content to write to the file."},
						},
						Required: []string{"path", "content"},
					},
				},
				{
					Name:        "delete_file",
					Description: "Deletes a file. For example, to delete a file named 'temp.txt', you would use delete_file('temp.txt').",
					Parameters: &genai.Schema{
						Type: genai.TypeObject,
						Properties: map[string]*genai.Schema{
							"path": {Type: genai.TypeString, Description: "The path of the file to delete."},
						},
						Required: []string{"path"},
					},
				},
				{
					Name:        "list_files",
					Description: "Lists all files and directories in a given path. Use '.' for the current directory.",
					Parameters: &genai.Schema{
						Type: genai.TypeObject,
						Properties: map[string]*genai.Schema{
							"path": {Type: genai.TypeString, Description: "The path of the directory to list."},
						},
						Required: []string{"path"},
					},
				},
				{
					Name:        "analyze_project",
					Description: "Analyzes the current project structure, detects programming language, framework, dependencies, and provides context about the project.",
					Parameters: &genai.Schema{
						Type: genai.TypeObject,
						Properties: map[string]*genai.Schema{
							"path": {Type: genai.TypeString, Description: "The root path of the project to analyze. Use '.' for current directory."},
						},
						Required: []string{"path"},
					},
				},
				{
					Name:        "generate_code",
					Description: "Generates code based on specifications and project context. Can generate functions, classes, tests, and configuration files.",
					Parameters: &genai.Schema{
						Type: genai.TypeObject,
						Properties: map[string]*genai.Schema{
							"type":        {Type: genai.TypeString, Description: "Type of code to generate: 'function', 'class', 'test', 'config'."},
							"name":        {Type: genai.TypeString, Description: "Name of the item to generate."},
							"description": {Type: genai.TypeString, Description: "Description of what the code should do."},
							"spec":        {Type: genai.TypeString, Description: "JSON specification for the code (parameters, fields, options)."},
						},
						Required: []string{"type", "name", "description"},
					},
				},
				{
					Name:        "install_dependencies",
					Description: "Installs project dependencies using the appropriate package manager.",
					Parameters: &genai.Schema{
						Type: genai.TypeObject,
						Properties: map[string]*genai.Schema{
							"packages": {Type: genai.TypeString, Description: "Space-separated list of packages to install (optional)."},
						},
					},
				},
				{
					Name:        "run_tests",
					Description: "Runs the project's test suite using the appropriate test framework.",
					Parameters: &genai.Schema{
						Type: genai.TypeObject,
						Properties: map[string]*genai.Schema{
							"pattern": {Type: genai.TypeString, Description: "Test pattern or specific test file to run (optional)."},
						},
					},
				},
				{
					Name:        "build_project",
					Description: "Builds the project using the appropriate build tool.",
					Parameters: &genai.Schema{
						Type: genai.TypeObject,
						Properties: map[string]*genai.Schema{
							"target": {Type: genai.TypeString, Description: "Build target (optional)."},
						},
					},
				},
				{
					Name:        "generate_web_file",
					Description: "Generates unique HTML, CSS, or JavaScript files using original patterns to avoid recitation blocks. Use this for web development instead of create_file for HTML/CSS/JS.",
					Parameters: &genai.Schema{
						Type: genai.TypeObject,
						Properties: map[string]*genai.Schema{
							"file_type": {Type: genai.TypeString, Description: "Type of web file: 'html', 'css', or 'js'."},
							"filename":  {Type: genai.TypeString, Description: "Name of the file to create (e.g., 'index.html', 'styles.css')."},
							"options":   {Type: genai.TypeString, Description: "JSON object with customization options (title, colors, features, etc.)."},
						},
						Required: []string{"file_type", "filename"},
					},
				},
			},
		},
	}
}

func generateToolDefinitions() string {
	var builder strings.Builder
	builder.WriteString("**Available Tools:**\n\n")
	tools := defineTools()
	for _, tool := range tools {
		for _, decl := range tool.FunctionDeclarations {
			builder.WriteString(fmt.Sprintf("- **%s**: %s\n", decl.Name, decl.Description))
		}
	}
	return builder.String()
}

type ToolExecutor struct {
	config      *config.Config
	projectInfo *agent.ProjectInfo
	analyzer    *agent.ProjectAnalyzer
	generator   *agent.CodeGenerator
}

func NewToolExecutor(config *config.Config) *ToolExecutor {
	cwd, _ := os.Getwd()
	analyzer := agent.NewProjectAnalyzer(cwd)
	
	return &ToolExecutor{
		config:   config,
		analyzer: analyzer,
	}
}

// executeTool is a dispatcher that calls the appropriate Go function for a given tool name.
func (e *ToolExecutor) Execute(fc genai.FunctionCall) (string, error) {
	switch fc.Name {
	case "execute_shell_command":
		if command, ok := fc.Args["command"].(string); ok {
			return commander.ExecuteCommand(command, e.config.AllowedCommands)
		}
		return "", fmt.Errorf("invalid or missing 'command' argument")
	case "create_file", "update_file":
		path, okPath := fc.Args["path"].(string)
		content, okContent := fc.Args["content"].(string)
		if !okPath || !okContent {
			return "", fmt.Errorf("invalid arguments for %s", fc.Name)
		}
		err := os.WriteFile(path, []byte(content), 0644)
		if err != nil {
			return "", err
		}
		return fmt.Sprintf("File '%s' was %sd successfully.", path, fc.Name), nil
	case "read_file":
		if path, ok := fc.Args["path"].(string); ok {
			content, err := os.ReadFile(path)
			if err != nil {
				return "", err
			}
			return string(content), nil
		}
		return "", fmt.Errorf("invalid or missing 'path' argument")
	case "delete_file":
		if path, ok := fc.Args["path"].(string); ok {
			err := os.Remove(path)
			if err != nil {
				return "", err
			}
			return "File deleted successfully.", nil
		}
		return "", fmt.Errorf("invalid or missing 'path' argument")
	case "list_files":
		if path, ok := fc.Args["path"].(string); ok {
			files, err := os.ReadDir(path)
			if err != nil {
				return "", err
			}
			var fileNames []string
			for _, file := range files {
				fileNames = append(fileNames, file.Name())
			}
			return strings.Join(fileNames, "\n"), nil
		}
		return "", fmt.Errorf("invalid or missing 'path' argument")
	case "analyze_project":
		if path, ok := fc.Args["path"].(string); ok {
			return e.analyzeProject(path)
		}
		return "", fmt.Errorf("invalid or missing 'path' argument")
	case "generate_code":
		return e.generateCode(fc)
	case "install_dependencies":
		return e.installDependencies(fc)
	case "run_tests":
		return e.runTests(fc)
	case "build_project":
		return e.buildProject(fc)
	case "generate_web_file":
		return e.generateWebFile(fc)
	default:
		return "", fmt.Errorf("unknown function call: %s", fc.Name)
	}
}

// analyzeProject analyzes the project structure and provides context
func (e *ToolExecutor) analyzeProject(path string) (string, error) {
	logger.Info("Analyzing project at path: %s", path)
	
	if path == "." {
		cwd, err := os.Getwd()
		if err != nil {
			return "", fmt.Errorf("failed to get current directory: %w", err)
		}
		path = cwd
	}
	
	analyzer := agent.NewProjectAnalyzer(path)
	projectInfo, err := analyzer.AnalyzeProject()
	if err != nil {
		logger.Error("Project analysis failed: %v", err)
		return "", fmt.Errorf("project analysis failed: %w", err)
	}
	
	// Cache the project info for future use
	e.projectInfo = projectInfo
	e.generator = agent.NewCodeGenerator(projectInfo)
	
	// Format the analysis result
	result, err := json.MarshalIndent(projectInfo, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to format analysis result: %w", err)
	}
	
	logger.Info("Project analysis completed successfully for %s project", projectInfo.Language)
	return fmt.Sprintf("Project Analysis Results:\n%s", string(result)), nil
}

// generateCode generates code based on specifications
func (e *ToolExecutor) generateCode(fc genai.FunctionCall) (string, error) {
	codeType, ok1 := fc.Args["type"].(string)
	name, ok2 := fc.Args["name"].(string)
	description, ok3 := fc.Args["description"].(string)
	
	if !ok1 || !ok2 || !ok3 {
		return "", fmt.Errorf("missing required arguments for code generation")
	}
	
	// Ensure we have project context
	if e.generator == nil {
		// Analyze project first
		if _, err := e.analyzeProject("."); err != nil {
			return "", fmt.Errorf("failed to analyze project context: %w", err)
		}
	}
	
	logger.Info("Generating %s code: %s", codeType, name)
	
	var code string
	var filename string
	var err error
	
	switch strings.ToLower(codeType) {
	case "function":
		// Parse function specification if provided
		var params, returns []string
		if spec, ok := fc.Args["spec"].(string); ok && spec != "" {
			var funcSpec struct {
				Params  []string `json:"params"`
				Returns []string `json:"returns"`
			}
			if err := json.Unmarshal([]byte(spec), &funcSpec); err == nil {
				params = funcSpec.Params
				returns = funcSpec.Returns
			}
		}
		code, err = e.generator.GenerateFunction(name, description, params, returns)
		filename = e.generator.GetSuggestedFilename("function", name)
		
	case "class", "struct":
		// Parse class specification if provided
		var fields []agent.Field
		if spec, ok := fc.Args["spec"].(string); ok && spec != "" {
			var classSpec struct {
				Fields []agent.Field `json:"fields"`
			}
			if err := json.Unmarshal([]byte(spec), &classSpec); err == nil {
				fields = classSpec.Fields
			}
		}
		code, err = e.generator.GenerateClass(name, description, fields)
		filename = e.generator.GetSuggestedFilename("class", name)
		
	case "test":
		code, err = e.generator.GenerateTest(name, "unit")
		filename = e.generator.GetSuggestedTestFilename(name)
		
	case "config":
		// Parse config options if provided
		options := make(map[string]interface{})
		if spec, ok := fc.Args["spec"].(string); ok && spec != "" {
			if err := json.Unmarshal([]byte(spec), &options); err != nil {
				logger.Warn("Failed to parse config spec: %v", err)
			}
		}
		code, err = e.generator.GenerateConfigFile(name, options)
		filename = name
		
	default:
		return "", fmt.Errorf("unsupported code type: %s", codeType)
	}
	
	if err != nil {
		logger.Error("Code generation failed: %v", err)
		return "", fmt.Errorf("code generation failed: %w", err)
	}
	
	result := fmt.Sprintf("Generated %s code for '%s':\n\nSuggested filename: %s\n\nCode:\n```\n%s\n```", 
		codeType, name, filename, code)
	
	logger.Info("Code generation completed successfully")
	return result, nil
}

// installDependencies installs project dependencies
func (e *ToolExecutor) installDependencies(fc genai.FunctionCall) (string, error) {
	// Ensure we have project context
	if e.projectInfo == nil {
		if _, err := e.analyzeProject("."); err != nil {
			return "", fmt.Errorf("failed to analyze project context: %w", err)
		}
	}
	
	packages, _ := fc.Args["packages"].(string)
	
	var command string
	switch e.projectInfo.PackageManager {
	case "npm":
		if packages != "" {
			command = fmt.Sprintf("npm install %s", packages)
		} else {
			command = "npm install"
		}
	case "yarn":
		if packages != "" {
			command = fmt.Sprintf("yarn add %s", packages)
		} else {
			command = "yarn install"
		}
	case "pnpm":
		if packages != "" {
			command = fmt.Sprintf("pnpm add %s", packages)
		} else {
			command = "pnpm install"
		}
	case "go":
		if packages != "" {
			command = fmt.Sprintf("go get %s", packages)
		} else {
			command = "go mod tidy"
		}
	case "pip":
		if packages != "" {
			command = fmt.Sprintf("pip install %s", packages)
		} else {
			command = "pip install -r requirements.txt"
		}
	case "cargo":
		if packages != "" {
			return "", fmt.Errorf("cargo doesn't support installing individual packages via command line")
		} else {
			command = "cargo build"
		}
	default:
		return "", fmt.Errorf("unknown package manager: %s", e.projectInfo.PackageManager)
	}
	
	logger.Info("Installing dependencies with command: %s", command)
	return commander.ExecuteCommand(command, e.config.AllowedCommands)
}

// runTests runs the project's test suite
func (e *ToolExecutor) runTests(fc genai.FunctionCall) (string, error) {
	// Ensure we have project context
	if e.projectInfo == nil {
		if _, err := e.analyzeProject("."); err != nil {
			return "", fmt.Errorf("failed to analyze project context: %w", err)
		}
	}
	
	pattern, _ := fc.Args["pattern"].(string)
	
	var command string
	switch e.projectInfo.Language {
	case "Go":
		if pattern != "" {
			command = fmt.Sprintf("go test %s", pattern)
		} else {
			command = "go test ./..."
		}
	case "JavaScript", "TypeScript":
		if e.projectInfo.TestFramework == "Jest" {
			if pattern != "" {
				command = fmt.Sprintf("%s test %s", e.projectInfo.PackageManager, pattern)
			} else {
				command = fmt.Sprintf("%s test", e.projectInfo.PackageManager)
			}
		} else {
			command = fmt.Sprintf("%s test", e.projectInfo.PackageManager)
		}
	case "Python":
		if e.projectInfo.TestFramework == "pytest" {
			if pattern != "" {
				command = fmt.Sprintf("pytest %s", pattern)
			} else {
				command = "pytest"
			}
		} else {
			command = "python -m unittest discover"
		}
	case "Rust":
		if pattern != "" {
			command = fmt.Sprintf("cargo test %s", pattern)
		} else {
			command = "cargo test"
		}
	default:
		return "", fmt.Errorf("testing not supported for language: %s", e.projectInfo.Language)
	}
	
	logger.Info("Running tests with command: %s", command)
	return commander.ExecuteCommand(command, e.config.AllowedCommands)
}

// buildProject builds the project
func (e *ToolExecutor) buildProject(fc genai.FunctionCall) (string, error) {
	// Ensure we have project context
	if e.projectInfo == nil {
		if _, err := e.analyzeProject("."); err != nil {
			return "", fmt.Errorf("failed to analyze project context: %w", err)
		}
	}
	
	target, _ := fc.Args["target"].(string)
	
	var command string
	switch e.projectInfo.Language {
	case "Go":
		if target != "" {
			command = fmt.Sprintf("go build -o %s .", target)
		} else {
			command = "go build ."
		}
	case "JavaScript", "TypeScript":
		if scripts, ok := e.projectInfo.Scripts["build"]; ok {
			command = fmt.Sprintf("%s run build", e.projectInfo.PackageManager)
			_ = scripts // Acknowledge that we have a build script
		} else {
			return "", fmt.Errorf("no build script found in package.json")
		}
	case "Python":
		if e.projectInfo.BuildTool == "poetry" {
			command = "poetry build"
		} else {
			command = "python setup.py build"
		}
	case "Rust":
		if target != "" {
			command = fmt.Sprintf("cargo build --bin %s", target)
		} else {
			command = "cargo build"
		}
	default:
		return "", fmt.Errorf("building not supported for language: %s", e.projectInfo.Language)
	}
	
	logger.Info("Building project with command: %s", command)
	return commander.ExecuteCommand(command, e.config.AllowedCommands)
}

// generateWebFile generates web files using unique patterns to avoid recitation blocks
func (e *ToolExecutor) generateWebFile(fc genai.FunctionCall) (string, error) {
	fileType, ok1 := fc.Args["file_type"].(string)
	filename, ok2 := fc.Args["filename"].(string)
	
	if !ok1 || !ok2 {
		return "", fmt.Errorf("missing required arguments for web file generation")
	}
	
	// Ensure we have project context
	if e.generator == nil {
		if _, err := e.analyzeProject("."); err != nil {
			return "", fmt.Errorf("failed to analyze project context: %w", err)
		}
	}
	
	logger.Info("Generating %s web file: %s", fileType, filename)
	
	// Parse options if provided
	options := make(map[string]interface{})
	if optionsStr, ok := fc.Args["options"].(string); ok && optionsStr != "" {
		if err := json.Unmarshal([]byte(optionsStr), &options); err != nil {
			logger.Warn("Failed to parse options: %v, using defaults", err)
		}
	}
	
	// Add unique elements to avoid recitation
	if options["appName"] == nil {
		options["appName"] = "Console Buddy"
	}
	if options["uniqueId"] == nil {
		options["uniqueId"] = "cb-app"
	}
	
	// Generate the web file content
	content, err := e.generator.GenerateWebFile(fileType, options)
	if err != nil {
		logger.Error("Web file generation failed: %v", err)
		return "", fmt.Errorf("web file generation failed: %w", err)
	}
	
	// Write the file
	if err := os.WriteFile(filename, []byte(content), 0644); err != nil {
		return "", fmt.Errorf("failed to write file %s: %w", filename, err)
	}
	
	logger.Info("Web file generation completed successfully: %s", filename)
	return fmt.Sprintf("Generated unique %s file '%s' successfully using Console Buddy templates to avoid recitation issues.", fileType, filename), nil
}
