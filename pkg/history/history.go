package history

import (
	"encoding/gob"
	"os"
)

const historyFileName = "conversation_history.gob"

// SaveHistory saves the conversation history to a binary file using gob encoding.
func SaveHistory(history []string) {
	f, err := os.Create(historyFileName)
	if err != nil {
		// Log error or handle it appropriately
		return
	}
	defer f.Close()
	enc := gob.NewEncoder(f)
	enc.Encode(history)
}

// LoadHistory loads the conversation history from the gob file.
func LoadHistory() []string {
	var history []string
	f, err := os.Open(historyFileName)
	if err != nil {
		// File might not exist on first run, which is fine.
		return []string{}
	}
	defer f.Close()

	dec := gob.NewDecoder(f)
	dec.Decode(&history)
	return history
}
