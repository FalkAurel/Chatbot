package api

import (
	"bytes"
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"
	"strconv"
	"time"
)

const documentUpload string = "http://ml_pipeline:3030/api/document/upload"
const messageUpload string = "http://ml_pipeline:3030/api/message/upload"
const messageInference string = "http://ml_pipeline:3030/api/message/inference"
const messageHistory string = "http://ml_pipeline:3030/api/message/history"
const messageDeletion string = "http://ml_pipeline:3030/api/delete/history"
const userDeletion string = "http://ml_pipeline:3030/api/delete/user"
const documentDeletion string = "http://ml_pipeline:3030/api/delete/document"
const modelListing string = "http://ollama:11434/api/tags"

var client *http.Client = &http.Client{
	Timeout: 1200 * time.Second, // Always set a timeout
}

func SendToMLPipeline(request *http.Request) (*http.Response, error) {
	// Debug: Log the outgoing request
	if dump, err := httputil.DumpRequestOut(request, true); err == nil {
		log.Printf("Sending request:\n%s", dump)
	}

	response, err := client.Do(request)
	if err != nil {
		log.Printf("Request failed: %v", err)
		return nil, fmt.Errorf("request failed: %w", err)
	}

	// Debug: Log the response
	if dump, err := httputil.DumpResponse(response, true); err == nil {
		log.Printf("Received response:\n%s", dump)
	}

	return response, nil
}

func UploadDocument(id int64, title string, storage_name string, data []byte) (*http.Response, error) {
	if len(data) == 0 {
		return nil, fmt.Errorf("empty data provided")
	}

	request, err := http.NewRequest("POST", documentUpload, bytes.NewReader(data))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set required headers
	request.Header.Set("Title", title)
	request.Header.Set("X-Filename", storage_name)
	request.Header.Set("ID", strconv.FormatInt(id, 10))
	request.Header.Set("Content-Type", "application/octet-stream") // Important for binary data
	request.Header.Set("Content-Length", strconv.Itoa(len(data)))

	return SendToMLPipeline(request)
}

func UploadMessage(id int64, data []byte) (*http.Response, error) {
	if len(data) == 0 {
		return nil, fmt.Errorf("empty data provided")
	}

	request, err := http.NewRequest("POST", messageUpload, bytes.NewReader(data))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	request.Header.Set("ID", strconv.FormatInt(id, 10))
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("Content-Length", strconv.Itoa(len(data)))

	return SendToMLPipeline(request)
}

// Sends a message to be processed by the LLM
// Ensures that the message has the appropriate headers set
//   - Custom header: Deep_think: True or False
func InferenceMessage(id int64, deep_think_header string, data []byte) (*http.Response, error) {
	if len(data) == 0 {
		return nil, fmt.Errorf("empty data provided")
	}

	request, err := http.NewRequest("GET", messageInference, bytes.NewReader(data))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	//request.Header.Set("Model", model)
	request.Header.Set("ID", strconv.FormatInt(id, 10))
	request.Header.Set("Deep_think", deep_think_header)
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("Content-Length", strconv.Itoa(len(data)))

	return SendToMLPipeline(request)
}

func GetMessageHistory(id int64) (*http.Response, error) {
	request, err := http.NewRequest("GET", messageHistory, bytes.NewBuffer([]byte{}))

	if err != nil {
		return nil, err
	}

	request.Header.Set("ID", strconv.FormatInt(id, 10))
	return SendToMLPipeline(request)
}

func MLDeleteUser(id int64) (*http.Response, error) {
	request, err := http.NewRequest("DELETE", userDeletion, bytes.NewBuffer([]byte{}))

	if err != nil {
		return nil, err
	}

	request.Header.Set("ID", strconv.FormatInt(id, 10))
	return SendToMLPipeline(request)
}

func MLDeleteDocument(id int64, storage_name string) (*http.Response, error) {
	request, err := http.NewRequest("DELETE", documentDeletion, bytes.NewBuffer([]byte{}))

	if err != nil {
		return nil, err
	}
	request.Header.Set("ID", strconv.FormatInt(id, 10))
	request.Header.Set("X-Filename", storage_name)
	return SendToMLPipeline(request)
}

func MLDeleteChat(id int64) (*http.Response, error) {
	request, err := http.NewRequest("DELETE", messageDeletion, bytes.NewBuffer([]byte{}))

	if err != nil {
		return nil, err
	}
	request.Header.Set("ID", strconv.FormatInt(id, 10))

	return SendToMLPipeline(request)
}

func MLGetModels() (*http.Response, error) {
	request, err := http.NewRequest("GET", modelListing, bytes.NewBuffer([]byte{}))

	if err != nil {
		return nil, err
	}

	return SendToMLPipeline(request)
}
