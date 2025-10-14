package history

import (
	"encoding/gob"
	"os"
)

// SaveConversation saves the conversation history to a binary file.
func SaveConversation(history []string) {
	f, err := os.OpenFile("conversation_history.bin", os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		return
	}
	defer f.Close()
	enc := gob.NewEncoder(f)
	enc.Encode(history)
}

// LoadConversation loads the conversation history from a binary file.
func LoadConversation() []string {
	var history []string
	f, err := os.Open("conversation_history.bin")
	if err == nil {
		dec := gob.NewDecoder(f)
		dec.Decode(&history)
		f.Close()
	}
	return history
}
