package api

import (
	"backend/auth"
	"backend/db"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
)

// Legal_libary handles requests related to legal library operations.
//
// Expects a JSON payload in the request body with LegalLibary structure.
// Returns 200 OK with success status on valid requests.
//
// Error responses:
//   - 400 Bad Request: For body read errors or invalid JSON
func Legal_libary(w http.ResponseWriter, r *http.Request) {
	if r.Method != "PUT" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	data, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Failed to read request body", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	var legal_library LegalLibary
	if err := json.Unmarshal(data, &legal_library); err != nil {
		http.Error(w, fmt.Sprintf("Invalid JSON: %v", err), http.StatusBadRequest)
		return
	}

	println(legal_library.Legal_library)

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status": "success"}`))
}

// Local_only handles requests to update local-only application state.
//
// Path: /api/update/local_only
// Expects a JSON payload in the request body with LocalOnly structure.
// Returns 200 OK with success status on valid requests.
//
// Error responses:
//   - 400 Bad Request: For body read errors or invalid JSON
func Local_only(w http.ResponseWriter, r *http.Request) {
	if r.Method != "PUT" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	data, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Failed to read request body", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	var local_only LocalOnly
	if err := json.Unmarshal(data, &local_only); err != nil {
		http.Error(w, fmt.Sprintf("Invalid JSON: %v", err), http.StatusBadRequest)
		return
	}

	println(local_only.Local_only)

	// 4. Success response
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status": "success"}`))
}

// AcceptSignupRequest processes pending signup requests by:
//  1. Removing the request from signup_requests table
//  2. Creating a new user account with the credentials
//
// Expects the email of the user to process as raw bytes in request body.
// Returns empty 200 OK response on success.
//
// Error responses:
//   - 400 Bad Request: For body read errors
//   - 500 Internal Server Error: For database operation failures
//
// Note:
//   - New users are created with admin=false privileges
//   - Logs errors but doesn't expose detailed error messages to client
func AcceptSignupRequest(db_handle *sql.DB, w http.ResponseWriter, r *http.Request) {
	if r.Method != "PUT" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	data, err := io.ReadAll(r.Body)
	defer r.Body.Close()

	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	entry, err := db.DeleteSignupRequest(db_handle, string(data[:]))
	if err != nil {
		log.Println("Failed to delete user", err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	err = db.AddUser(
		db_handle,
		db.CreateUser(entry.Name, entry.Password, entry.Email, false),
	)

	if err != nil {
		log.Println("Failed to add user", err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	user_id, err := db.GetUserID(db_handle, entry.Email)
	if err != nil {
		log.Println("Failed to get user", err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	err = db.AddPrompt(db_handle, user_id, DEFAULT_PREPROMPT)
	if err != nil {
		log.Println("Failed to create promt for user", err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte{})
}

// UpdatePrompt handles HTTP requests to update a user's custom prompt.
//
// This is a PUT-only endpoint that:
// - Requires authentication (via auth_result)
// - Accepts either:
//   - An empty body to reset to DEFAULT_PREPROMPT
//   - A new prompt string in the request body
//
// Parameters:
//   - auth_result: Authenticated user's authorization details containing their ID
//   - db_handle:   Active database connection pool
//   - w:           HTTP response writer
//   - r:           HTTP request object
//
// Responses:
//   - 200 OK:                   Prompt updated successfully
//   - 400 Bad Request:          Malformed request body
//   - 401 Unauthorized:         Implicit via auth middleware (not shown in function)
//   - 405 Method Not Allowed:   Non-PUT requests
//   - 500 Internal Server Error: Database operation failure
//
// Security:
// - Requires pre-authentication (handled by middleware)
// - Uses parameterized queries via db.UpdatePrompt
func UpdatePrompt(
	auth_result auth.AuthorizationResult,
	db_handle *sql.DB,
	w http.ResponseWriter,
	r *http.Request,
) {
	if r.Method != "PUT" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	data, err := io.ReadAll(r.Body)
	defer r.Body.Close()

	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if len(data) == 0 {
		err = db.UpdatePrompt(db_handle, auth_result.ID, DEFAULT_PREPROMPT)

		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	} else {
		prompt := string(data[:])

		err = db.UpdatePrompt(db_handle, auth_result.ID, prompt)

		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte{})
}

// UpdateDefaultPrompt handles HTTP requests to update the default system prompt.
//
// This endpoint expects a plain text request body containing the new prompt.
// On success, it returns HTTP 200 with an empty body.
//
// Parameters:
//   - db_handle: Database connection handle (currently unused, preserved for future extensions)
//   - w: HTTP response writer
//   - r: HTTP request containing the new prompt in its body
//
// Possible status codes:
//   - 200 OK: Prompt updated successfully
//   - 400 Bad Request: Invalid request body
//   - 500 Internal Server Error: (future use for database operations)
func UpdateDefaultPromt(db_handle *sql.DB, w http.ResponseWriter, r *http.Request) {
	data, err := io.ReadAll(r.Body)
	defer r.Body.Close()

	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	var new_default_promt string = string(data[:])

	if len(new_default_promt) == 0 {
		new_default_promt = BACK_UP_PROMPT
	}
	err = SetDefaultPrompt(new_default_promt)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Write([]byte{})
}

// PromoteUser handles HTTP requests to promote a user's privileges.
//
// This is a PUT-only endpoint that expects a user email in the request body.
// It calls the database layer to perform the promotion operation.
//
// Parameters:
//   - db_handle: Database connection pool
//   - w: HTTP response writer
//   - r: HTTP request object
//
// Responses:
//   - 200 OK: On successful promotion
//   - 400 Bad Request: If body cannot be read
//   - 405 Method Not Allowed: If request method isn't PUT
//   - 500 Internal Server Error: If database operation fails
func PromoteUser(db_handle *sql.DB, w http.ResponseWriter, r *http.Request) {
	if r.Method != "PUT" {
		http.Error(w, "Method not followed", http.StatusMethodNotAllowed)
		return
	}

	data, err := io.ReadAll(r.Body)
	defer r.Body.Close()

	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	err = db.PromoteUser(db_handle, string(data[:])) // Email is in data

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte{})
}

// UpdateModelSelection handles HTTP requests to update the active AI model.
//
// This is a PUT-only endpoint that expects the model name in the request body.
// It updates the application's selected model through the database layer.
//
// Parameters:
//   - w: HTTP response writer
//   - r: HTTP request object
//
// Responses:
//   - 200 OK: On successful model update
//   - 400 Bad Request: If body cannot be read
//   - 405 Method Not Allowed: If request method isn't PUT
func UpdateModelSelection(w http.ResponseWriter, r *http.Request) {
	if r.Method != "PUT" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	data, err := io.ReadAll(r.Body)
	defer r.Body.Close()

	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	var model_selection string = string(data[:])
	db.SetModel(model_selection)

	w.WriteHeader(http.StatusOK)
	w.Write([]byte{})
}
