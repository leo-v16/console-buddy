package commander

import (
	"fmt"
	"os/exec"
	"strings"
)

func ExecuteCommand(command string) (string, error) {
	// Sanitize command: remove leading/trailing quotes, backticks, and whitespace
	command = strings.TrimSpace(command)
	command = strings.Trim(command, "`\"")
	if len(command) == 0 {
		return "", fmt.Errorf("empty command")
	}

	// List of allowed commands (add more as needed)
	allowedCommands := map[string]bool{
		"dir":  true,
		"echo": true,
		"type": true,
		"copy": true,
		"del":  true,
		"cls":  true,
		"cd":   true,
		// Add more Windows commands as needed
	}

	// Get the first word to check if allowed (allow punctuation, e.g., echo.)
	firstWord := strings.Fields(command)
	if len(firstWord) == 0 {
		return "", fmt.Errorf("command '%s' is not allowed or not recognized", command)
	}
	baseCmd := strings.Trim(strings.ToLower(firstWord[0]), ".:;`'!")
	if !allowedCommands[baseCmd] {
		return "", fmt.Errorf("command '%s' is not allowed or not recognized", command)
	}

	cmd := exec.Command("cmd.exe", "/C", command)
	output, err := cmd.CombinedOutput()
	return string(output), err
}
