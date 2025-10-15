package gemini

import (
	"fmt"
	"os"
	"strings"

	"console-ai/pkg/commander"

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
			},
		},
	}
}

// executeTool is a dispatcher that calls the appropriate Go function for a given tool name.
func executeTool(fc genai.FunctionCall) (string, error) {
	switch fc.Name {
	case "execute_shell_command":
		if command, ok := fc.Args["command"].(string); ok {
			return commander.ExecuteCommand(command)
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
	default:
		return "", fmt.Errorf("unknown function call: %s", fc.Name)
	}
}
