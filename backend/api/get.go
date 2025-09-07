package api

import (
	"backend/auth"
	"backend/db"
	"database/sql"
	"encoding/json"
	"io"
	"net/http"
)

// GetHistory retrieves the message history for an authenticated user.
//
// Parameters:
//   - auth_result: Contains authenticated user's ID and authorization status
//   - w: HTTP response writer
//   - r: HTTP request object
//
// Behavior:
//   - Requires successful authorization via auth_result
//   - Proxies the request to GetMessageHistory service
//   - Returns the exact status code and body from the history service
//
// Error Responses:
//   - 500 Internal Server Error: If history retrieval fails
//   - Propagates any error from GetMessageHistory service
func GetHistory(auth_result auth.AuthorizationResult, w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	response, err := GetMessageHistory(auth_result.ID)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer response.Body.Close()

	body, err := io.ReadAll(response.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(response.StatusCode)
	w.Write(body)
}

// GetSignupRequests retrieves all pending signup requests from the database.
//
// Parameters:
//   - db_handle: Database connection handle
//   - w: HTTP response writer
//
// Responses:
//   - 200 OK: With JSON array of signup requests on success
//   - 400 Bad Request: If database query fails
//
// Note:
//   - Returns minimal request info (name and email) without sensitive data
//   - Uses db.GetSignupRequests for the actual database operation
func GetSignupRequests(db_handle *sql.DB, w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	login_requests, err := db.GetSignupRequests(db_handle)

	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(login_requests)
}

// GetUser retrieves and returns all user information from the system.
//
// This endpoint:
//   - Only accepts GET requests
//   - Returns a JSON array of user objects
//   - Returns 400 Bad Request for non-GET methods or database errors
//   - Returns 200 OK with user data on success
//
// Response Format:
//
// The response body contains a JSON array of user objects.
// Refer to the User struct in the db package for the exact field structure.
//
// Example Response:
//
//	[
//	  {
//		"Email": string
//		"IsAdmin": bool
//	  }
//	]
func GetUser(db_handle *sql.DB, w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, "Method not allowed", http.StatusBadRequest)
		return
	}

	user_info, err := db.GetUsers(db_handle)

	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(user_info)
}

// GetDocuments retrieves all documents for the authenticated user.
//
// This endpoint:
//   - Requires authentication (via auth_result)
//   - Returns a JSON array of document objects
//   - Returns 500 Internal Server Error for database failures
//   - Returns 200 OK with document data on success
//
// Response Format:
//
//	The response body contains a JSON array of document objects.
//	Refer to the Document struct in the db package for the exact field structure.
//
// Example Response:
//
//		[
//		  {
//			"OriginalName": string
//	     	"StorageName": string
//		  }
//		]
func GetDocuments(
	auth_result auth.AuthorizationResult,
	db_handle *sql.DB,
	w http.ResponseWriter,
	r *http.Request,
) {
	if r.Method != "GET" {
		http.Error(w, "Method not allowed", http.StatusBadRequest)
		return
	}

	documents, err := db.GetDocuments(db_handle, auth_result.ID)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(documents)
}

// GetModels handles HTTP requests to retrieve available AI models from the ML service.
//
// This is a GET-only endpoint that:
// - Proxies the request to the ML service
// - Returns the model list as a JSON response
// - Maintains the original status code from the ML service
//
// Parameters:
//   - w: HTTP response writer
//   - r: HTTP request object
//
// Responses:
//   - 200 OK: JSON array of available models
//   - 400 Bad Request: If called with non-GET method
//   - 5xx: Propagates any error from the ML service
func GetModels(
	w http.ResponseWriter,
	r *http.Request,
) {
	if r.Method != "GET" {
		http.Error(w, "Method not allowed", http.StatusBadRequest)
		return
	}

	response, err := MLGetModels()

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	body, err := io.ReadAll(response.Body)
	defer response.Body.Close()

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	println(string(body[:])) // Prints out correctly a JSON-Object

	if response.StatusCode != 200 {
		http.Error(w, string(body[:]), response.StatusCode)
		return
	}

	println("Why does it not send the JSON body")

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(body)
}

// GetCurrentModel handles HTTP requests to retrieve the currently selected AI model.
//
// This endpoint:
// - Returns the model name as plain text
// - Does not require any parameters
//
// Parameters:
//   - w: HTTP response writer
//
// Responses:
//   - 200 OK: Plain text response with the current model name
func GetCurrentModel(w http.ResponseWriter) {
	var current_llm string = db.GetModel()

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(current_llm))
}

// GetPrompt handles HTTP requests to retrieve a user's custom prompt.
//
// This is a GET-only endpoint that:
// - Requires authentication (via auth_result)
// - Retrieves the user's stored prompt from the database
// - Returns the prompt as plain text
//
// Parameters:
//   - auth_result: Authenticated user's authorization details containing their ID
//   - db_handle:   Active database connection pool
//   - w:           HTTP response writer
//   - r:           HTTP request object
//
// Responses:
//   - 200 OK:                   Returns the user's stored prompt as plain text
//   - 401 Unauthorized:         Implicit via auth middleware
//   - 405 Method Not Allowed:   Non-GET requests
//   - 500 Internal Server Error: Database operation failure
//
// Security:
// - Requires pre-authentication (handled by middleware)
// - Only returns prompts for the authenticated user
//
// Example Request:
//
//	GET /api/prompt
//	Headers: Authorization: Bearer <token>
//
// Example Response:
//
//	"You are a helpful AI assistant that speaks German."
func GetPrompt(
	auth_result auth.AuthorizationResult,
	db_handle *sql.DB,
	w http.ResponseWriter,
	r *http.Request,
) {
	if r.Method != "GET" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	prompt, err := db.GetPrompt(db_handle, auth_result.ID)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(prompt))
}

// GetPrompt handles HTTP requests to retrieve a user's custom prompt.
//
// This is a GET-only endpoint that:
// - Requires authentication (via auth_result)
// - Retrieves the user's stored prompt from the database
// - Returns the prompt as plain text
//
// Parameters:
//   - auth_result: Authenticated user's authorization details containing their ID
//   - db_handle:   Active database connection pool
//   - w:           HTTP response writer
//   - r:           HTTP request object
//
// Responses:
//   - 200 OK:                   Returns the user's stored prompt as plain text
//   - 401 Unauthorized:         Implicit via auth middleware
//   - 405 Method Not Allowed:   Non-GET requests
//   - 500 Internal Server Error: Database operation failure
//
// Security:
// - Requires pre-authentication (handled by middleware)
// - Only returns prompts for the authenticated user
//
// Example Request:
//
//	GET /api/prompt
//	Headers: Authorization: Bearer <token>
//
// Example Response:
//
//	"You are a helpful AI assistant that speaks German."
func GetDefaultPrompt(
	w http.ResponseWriter,
	r *http.Request,
) {
	if r.Method != "GET" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	prompt := getDefaultPrompt()

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(prompt))
}
