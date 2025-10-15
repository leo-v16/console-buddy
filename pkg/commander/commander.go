package commander

import (
	"fmt"
	"os/exec"
	"runtime"
	"strings"
)

// ExecuteCommand runs a shell command after validating it against an allowlist.
func ExecuteCommand(command string, allowedCommands []string) (string, error) {
	command = strings.TrimSpace(command)
	if command == "" {
		return "", fmt.Errorf("empty command")
	}

	parts := strings.Fields(command)
	baseCmd := strings.ToLower(parts[0])

	isAllowed := false
	for _, allowed := range allowedCommands {
		if baseCmd == allowed {
			isAllowed = true
			break
		}
	}

	if !isAllowed {
		return "", fmt.Errorf("command '%s' is not allowed", baseCmd)
	}

	var cmd *exec.Cmd
	if runtime.GOOS == "windows" {
		cmd = exec.Command("cmd.exe", "/C", command)
	} else {
		cmd = exec.Command("sh", "-c", command)
	}

	output, err := cmd.CombinedOutput()
	if err != nil {
		return string(output), fmt.Errorf("command execution failed: %w\nOutput: %s", err, string(output))
	}
	return string(output), nil
}
