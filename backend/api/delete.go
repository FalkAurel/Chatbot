package api

import (
	"backend/auth"
	"backend/db"
	"database/sql"
	"io"
	"net/http"
)

// DeleteSignupRequest removes a pending signup request from the database.
//
// Parameters:
//   - db_handle: Active database connection
//   - w: HTTP response writer
//   - r: HTTP request containing the email to delete
//
// Request Format:
//   - Plain text email in request body (no JSON wrapping)
//
// Behavior:
//   - Looks up request by exact email match
//   - Deletes the entire signup request record
//   - Returns empty 200 response on success
//
// Response Codes:
//   - 200 OK: Successful deletion
//   - 400 Bad Request: Invalid/missing email in request body
//   - 500 Internal Server Error: Database operation failed
//
// Security:
//   - Email matching is case-sensitive
func DeleteSignupRequest(db_handle *sql.DB, w http.ResponseWriter, r *http.Request) {
	data, err := io.ReadAll(r.Body)
	defer r.Body.Close()

	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
	}

	_, err = db.DeleteSignupRequest(db_handle, string(data[:]))

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte{})
}

// DeleteUser handles the deletion of a user and all associated documents from the system.
//
// This endpoint performs a cascading deletion that:
//  1. Reads the user's email from the request body
//  2. Deletes the user record from the database (which cascades to linked documents)
//  3. Sends a deletion request to the ML pipeline for the user
//  4. Returns the ML pipeline's response to the client
//
// Parameters:
//   - db_handle: Database connection handle
//   - w: HTTP response writer
//   - r: HTTP request containing the user's email in its body
//
// Behavior:
//   - Expects the user's email as raw bytes in the request body
//   - Returns 400 Bad Request if the body cannot be read
//   - Returns 500 Internal Server Error if database deletion fails
//   - Returns 500 Internal Server Error if ML pipeline communication fails
//   - Propagates the ML pipeline's response status code and body to the client
//
// Database Notes:
//   - Uses foreign key constraints with cascading delete to automatically remove
//     all documents associated with the user when the user is deleted
//
// ML Pipeline Integration:
//   - Forwards the deletion request to maintain consistency across all services
//   - Returns whatever response the ML pipeline provides
func DeleteUser(db_handle *sql.DB, w http.ResponseWriter, r *http.Request) {
	data, err := io.ReadAll(r.Body)
	defer r.Body.Close()

	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
	}

	var email string = string(data[:])

	// The document database is linked via a foreing key to the user record,
	// meaning deleting the user deletes also all other document records
	user, err := db.DeleteUser(db_handle, email)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	response, err := MLDeleteUser(user.ID)

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

// DeleteDocument handles the deletion of a document from the system architecture.
//
// This endpoint performs the following operations:
//  1. Reads the document identifier (storage_name) from the request body
//  2. Sends a deletion request to the ML pipeline
//  3. On successful ML pipeline response, deletes the document from the database
//
// Parameters:
//   - auth_result: Authorization context containing user ID and permissions
//   - db_handle: Database connection handle
//   - w: HTTP response writer
//   - r: HTTP request containing the document identifier in its body
//
// Behavior:
//   - Expects the storage_name (document identifier) as raw bytes in the request body
//   - Returns 400 Bad Request if the body cannot be read
//   - Returns 500 Internal Server Error if ML pipeline communication fails
//   - Propagates the ML pipeline's error status code if deletion fails there
//   - Returns 500 Internal Server Error if database deletion fails
//   - Returns 200 OK on successful deletion
//
// Note: The function closes all request/response bodies automatically via defer statements.
func DeleteDocument(
	auth_result auth.AuthorizationResult,
	db_handle *sql.DB,
	w http.ResponseWriter,
	r *http.Request,
) {
	data, err := io.ReadAll(r.Body)
	defer r.Body.Close()

	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	response, err := MLDeleteDocument(auth_result.ID, string(data[:]))

	if err != nil {
		http.Error(w, "Sending to ML-Pipeline failed", http.StatusInternalServerError)
		return
	}

	if response.StatusCode != 200 {
		data, err = io.ReadAll(response.Body)

		defer response.Body.Close()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		http.Error(w, string(data[:]), response.StatusCode)
		return
	}

	defer response.Body.Close()

	err = db.DeleteDocument(db_handle, string(data[:]), auth_result.ID)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte{})
}

// DeleteChat handles the deletion of a chat session from the ML pipeline.
//
// This endpoint performs the following operations:
//  1. Sends a deletion request to the ML pipeline for the user's chat session
//  2. On successful ML pipeline response, returns success to the client
//
// Parameters:
//   - auth_result: Authorization context containing user ID
//   - w: HTTP response writer
//   - r: HTTP request (unused body, but closed automatically)
//
// Behavior:
//   - Returns 500 Internal Server Error if ML pipeline communication fails
//   - Propagates the ML pipeline's error status code if deletion fails there
//   - Returns 200 OK on successful deletion
//
// Note:
//   - Unlike DeleteDocument, this only interacts with the ML pipeline and not the database
//   - The function closes all request/response bodies automatically via defer statements
func DeleteChat(auth_result auth.AuthorizationResult, w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	response, err := MLDeleteChat(auth_result.ID)

	if err != nil {
		http.Error(w, "Sending to ML-Pipeline failed", http.StatusInternalServerError)
		return
	}

	if response.StatusCode != 200 {
		data, err := io.ReadAll(response.Body)

		defer response.Body.Close()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		http.Error(w, string(data[:]), response.StatusCode)
		return
	}

	defer response.Body.Close()

	w.WriteHeader(http.StatusOK)
	w.Write([]byte{})

}
