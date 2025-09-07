package db

import (
	"sync"
)

var (
	modelMutex sync.Mutex // Protects concurrent access to `llm`
	llm        string     // Stores the model string in memory
)

// Initialize with default model ("gemma3:12b" as specified)
func InitModelSelection() {
	llm = "gemma3:12b"
}

// GetModel returns the current model (thread-safe read)
func GetModel() string {
	modelMutex.Lock()
	defer modelMutex.Unlock()
	return llm
}

// SetModel updates the model (thread-safe write)
func SetModel(model string) {
	modelMutex.Lock()
	defer modelMutex.Unlock()
	llm = model
	println("LLM:", llm, "Selected Model:", model)
}
