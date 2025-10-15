package history

import (
	"encoding/gob"
	"os"
	"path/filepath"
	"time"

	"console-ai/pkg/agent"
)

// SessionData contains all data stored in CB.hist
type SessionData struct {
	ProjectInfo    *agent.ProjectInfo `json:"project_info"`
	Conversations  []string          `json:"conversations"`
	LastUpdated    time.Time         `json:"last_updated"`
	TotalSessions  int               `json:"total_sessions"`
	HumorLevel     int               `json:"humor_level"`
}

// SaveHistory saves the conversation history and project context to CB.hist.
// The file is saved as CB.hist in the current working directory.
func SaveHistory(path string, history []string) error {
	return SaveSession(path, history, nil, 0)
}

// SaveSession saves both conversation history and project context to CB.hist.
func SaveSession(path string, history []string, projectInfo *agent.ProjectInfo, humorLevel int) error {
	// Always use CB.hist in current working directory
	if path == "" || path == "conversation_history.json" || path == "CB.hist" {
		cwd, err := os.Getwd()
		if err != nil {
			// Fallback to current directory if we can't get working directory
			path = "CB.hist"
		} else {
			path = filepath.Join(cwd, "CB.hist")
		}
	}

	// Load existing session data if it exists
	existingData, _ := LoadSession(path)
	if existingData == nil {
		existingData = &SessionData{
			TotalSessions: 0,
			HumorLevel:    humorLevel,
		}
	}

	// Update session data
	existingData.Conversations = history
	existingData.LastUpdated = time.Now()
	existingData.TotalSessions++
	if projectInfo != nil {
		existingData.ProjectInfo = projectInfo
	}
	if humorLevel > 0 {
		existingData.HumorLevel = humorLevel
	}

	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()

	enc := gob.NewEncoder(f)
	return enc.Encode(existingData)
}

// LoadHistory loads just the conversation history from CB.hist for backward compatibility.
func LoadHistory(path string) ([]string, error) {
	sessionData, err := LoadSession(path)
	if err != nil || sessionData == nil {
		return []string{}, nil
	}
	return sessionData.Conversations, nil
}

// LoadSession loads the complete session data from CB.hist binary file.
// Looks for CB.hist in the current working directory.
func LoadSession(path string) (*SessionData, error) {
	// Always use CB.hist in current working directory
	if path == "" || path == "conversation_history.json" || path == "CB.hist" {
		cwd, err := os.Getwd()
		if err != nil {
			// Fallback to current directory if we can't get working directory
			path = "CB.hist"
		} else {
			path = filepath.Join(cwd, "CB.hist")
		}
	}

	f, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			// Return nil if file doesn't exist
			return nil, nil
		}
		return nil, err
	}
	defer f.Close()

	dec := gob.NewDecoder(f)
	
	// Try to decode as SessionData first
	var sessionData SessionData
	if err := dec.Decode(&sessionData); err != nil {
		// If that fails, try to decode as old format ([]string)
		f.Seek(0, 0)
		dec = gob.NewDecoder(f)
		var oldHistory []string
		if err2 := dec.Decode(&oldHistory); err2 != nil {
			// Both failed, return empty
			return nil, nil
		}
		// Convert old format to new format
		return &SessionData{
			Conversations: oldHistory,
			LastUpdated:   time.Now(),
			TotalSessions: 1,
			HumorLevel:    0,
		}, nil
	}
	
	return &sessionData, nil
}
