package commander

import (
	"fmt"
	"os/exec"
	"strings"
)

// ExecuteCommand runs a shell command after validating it against an allowlist.
func ExecuteCommand(command string) (string, error) {
	// Sanitize command: remove leading/trailing quotes and whitespace
	command = strings.TrimSpace(command)
	command = strings.Trim(command, `"'`)
	if len(command) == 0 {
		return "", fmt.Errorf("empty command")
	}

	// List of allowed commands to prevent arbitrary code execution.
	// This is a security measure.
	allowedCommands := map[string]bool{
		// Windows specific
		"dir":  true,
		"type": true,
		"copy": true,
		"del":  true,
		"cls":  true,
		"cd":   true,
		"md":   true,
		"rd":   true,

		// General development tools
		"go":   true,
		"git":  true,
		"npm":  true,
		"node": true,
		"pip":  true,
		"py":   true,
	}

	parts := strings.Fields(command)
	if len(parts) == 0 {
		return "", fmt.Errorf("empty command")
	}

	baseCmd := strings.ToLower(parts[0])
	if !allowedCommands[baseCmd] {
		return "", fmt.Errorf("command '%s' is not allowed", baseCmd)
	}

	// Execute the command using cmd.exe on Windows.
	cmd := exec.Command("cmd.exe", "/C", command)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return string(output), fmt.Errorf("command execution failed: %w\nOutput: %s", err, string(output))
	}
	return string(output), nil
}
