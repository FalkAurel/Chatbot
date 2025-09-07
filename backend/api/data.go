package api

import (
	"backend/db"
	"errors"
	"os"
	"sync"
)

// Defines a JSON-Object for transferring Chat Messages.
// This structure is used for communication between the user interface and the AI Agent.
//
//   - Kind defines if the message originates from a user or from the AI Agent (0 = AI, 1 = User).
//     This field is crucial for maintaining conversation context and identifying the message source.
//   - Message is the content of the message in string form.  This holds the actual text being communicated.
type Message struct {
	Kind    int
	Message string
}

// MLMessage extends the Message struct to include metadata relevant for the AI model.
// This structure is used for messages that require the AI model's processing and response generation.
//
//   - Kind defines if the message originates from a user or from the AI Agent (0 = AI, 1 = User).
//   - Message is the content of the message in string form.
//   - Model identifies the specific AI model used to process the message (e.g., "GPT-3", "LLaMA").
//   - Preprompt provides the initial instructions or context for the AI model to use when generating a response.
//     *Important Legal Note:* The `Preprompt` field is particularly sensitive.
//     Modifying this field can drastically alter the AI Agent’s behavior and responses.
//     Handle with care and ensure appropriate access controls.
//     Incorrect use may lead to unexpected or undesirable outcomes.
type MLMessage struct {
	Kind      int
	Message   string
	Model     string
	Preprompt string
}

type LegalLibary struct {
	Legal_library bool
}

type LocalOnly struct {
	Local_only bool
}

type LoginRequestReponse struct {
	Requests []db.SignupRequestDB
}

const BACK_UP_PROMPT string = `
## SPRACH- UND ANTWORTREGELN (STRENG)

1. **Sprache**  
   - Antworte immer auf Deutsch, auch bei englischen Eingaben.  
   - Fachbegriffe übersetzen: "AI" → "KI", "Machine Learning" → "ML", etc.  

2. **Rechtsbezogene Angaben (Legal AI)**  
   - Bei Kontextnutzung: Dokumententitel in Klammern angeben („Beispiel.pdf“).  
   - Keine relevanten Infos im Kontext? Vermerke: (Keine passenden Informationen gefunden.)  

3. **Stil & Präzision**  
   - Strukturierte Antworten (Absätze/Aufzählungen).  
   - Sachlich, aber verständlich formuliert.
`

// Package level variables to manage the default pre-prompt.
//
// DEFAULT_PREPROMPT holds the current default pre-prompt string. It is initialized with the value of BACK_UP_PROMPT.
// defaultPromptMutex is a read/write mutex used to synchronize access to the DEFAULT_PREPROMPT variable, ensuring thread safety.
var (
	DEFAULT_PREPROMPT  string = BACK_UP_PROMPT
	defaultPromptMutex sync.RWMutex
)

// SetDefaultPrompt sets the default pre-prompt string.
// It acquires a write lock on defaultPromptMutex to prevent concurrent access,
// then updates the DEFAULT_PREPROMPT variable and persists the new prompt to disk.
//
// Parameters:
//
//	new: The new default pre-prompt string to set.
//
// Returns:
//
//	error: An error if writing the prompt to disk fails; nil otherwise.
//
// **Legal Note:** This function modifies a global variable and writes to a file.  Proper error handling and consideration of potential race conditions should be implemented in the calling code.
func SetDefaultPrompt(new string) error {
	defaultPromptMutex.Lock()
	defer defaultPromptMutex.Unlock()
	DEFAULT_PREPROMPT = new

	return Write_default_prompt("./data/default_prompt.txt", new)
}

// getDefaultPrompt retrieves the current default prompt string.
// It acquires a read lock on defaultPromptMutex to allow concurrent read access while preventing write access.
//
// Returns:
//
//	string: The current default prompt string.
func getDefaultPrompt() string {
	defaultPromptMutex.RLock()
	defer defaultPromptMutex.RUnlock()
	return DEFAULT_PREPROMPT
}

// Load_default_prompt reads the default prompt from a file.
// This function is crucial for initializing the AI Agent with its core instructions.
// It is intended to be run before the start of the program.
//
// Parameters:
//
//	path: The path to the file containing the default prompt.
//
// Returns:
//
//	string: The content of the file as a string.
//	error: An error if the file cannot be read; nil otherwise.  An error is also returned if the file is empty.
func Load_default_prompt(path string) (string, error) {
	content, err := os.ReadFile(path)

	if err != nil {
		return string([]byte{}), err
	}

	if len(content) == 0 {
		return string([]byte{}), errors.New("no content, creating a new file")
	}

	return string(content[:]), nil
}

// Write_default_prompt writes the default prompt to a file.
// This function is used to persist changes to the AI Agent's core instructions.
//
// Parameters:
//
//	path: The path to the file where the default prompt should be written.
//	prompt: The new default prompt string.
//
// Returns:
//
//	error: An error if the file cannot be written; nil otherwise.
//
// **Legal Note:**  Care should be taken when modifying the default prompt,
// as this directly impacts the AI Agent’s behavior. Ensure appropriate access controls are in
// place to prevent unauthorized modifications.
func Write_default_prompt(path string, prompt string) error {
	err := os.WriteFile(path, []byte(prompt), os.FileMode(os.O_WRONLY))

	return err
}
