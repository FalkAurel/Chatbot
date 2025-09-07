package api

import (
	"backend/db"
	"database/sql"
	"encoding/json"
	"io"
	"net/http"
)

// HandleSignupRequest is the top-level HTTP handler for processing user signup requests.
//
// It expects a JSON payload in the request body with the following structure:
//
//	{
//		"name":    string,
//		"password": string,
//		"email":   string
//	}
//
// The handler performs these steps:
//  1. Validates the input JSON structure
//  2. Verifies the email isn't already registered
//  3. Adds the signup request to the database using AddSignupRequest
//
// Possible error responses:
//   - 400 Bad Request: Invalid JSON or missing required fields
//   - 409 Conflict: Email already exists
//   - 500 Internal Server Error: Database operation failed
func HandleSignUpRequest(db_handler *sql.DB, w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	data, err := io.ReadAll(r.Body)
	defer r.Body.Close()

	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	var signup_request db.SignupRequest
	err = json.Unmarshal(data, &signup_request)

	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
	}

	user_exists, err := db.ExistsEmailInUser(db_handler, signup_request.Email)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	signup_exists, err := db.ExistsEmailInSignUp(db_handler, signup_request.Email)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	if signup_exists || user_exists {
		http.Error(w, "email already exists", http.StatusConflict)
		return
	}

	err = db.AddSignupRequest(db_handler, signup_request)

	if err != nil {
		http.Error(w, err.Error(), http.StatusConflict)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte{})
}
