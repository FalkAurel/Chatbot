package api

import (
	"backend/auth"
	"backend/db"
	"bytes"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"
)

const fileSizeLimit int64 = (1 << 20) * 200 // 200 MB

// FileUpload handles document uploads with PDF validation and storage.
//
// Requirements:
//   - POST method only
//   - X-Filename header must be present
//   - Title header must be present
//   - PDF file validation (signature and extension)
//   - Maximum file size: 200MB
//
// Process flow:
//  1. Validates headers and file type
//  2. Generates unique storage name
//  3. Uploads to document service
//  4. Records metadata in database
//
// Responses:
//   - 200 OK: Upload successful
//   - 400 Bad Request: Invalid headers, file type, or size
//   - 500 Internal ServerError: Upload or database failure
//
// Security:
//   - Requires valid authentication
//   - Uses SHA-256 hashed storage names
func FileUpload(db_handle *sql.DB, auth_result auth.AuthorizationResult, w http.ResponseWriter, r *http.Request) {
	// 1. Validate request headers
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	filename := r.Header.Get("X-Filename")
	if filename == "" {
		http.Error(w, "Missing X-Filename header", http.StatusBadRequest)
		return
	}

	title := r.Header.Get("Title")
	if title == "" {
		http.Error(w, "Missing Title header", http.StatusBadRequest)
		return
	}

	// 2. Limit and read body
	r.Body = http.MaxBytesReader(w, r.Body, fileSizeLimit)
	data, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Failed to read request body", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	// 3. Validate PDF
	if err := validate_legal_pdf(filename, data); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// 4. Process upload
	storage_name := create_storage_name(auth_result.ID, filename)
	response, err := UploadDocument(auth_result.ID, title, storage_name, data)
	if err != nil {
		http.Error(w, fmt.Sprintf("Upload failed: %v", err), http.StatusInternalServerError)
		return
	}
	defer response.Body.Close() // Always close the response body

	// 5. Check response status
	if response.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(response.Body) // Read error response if available
		http.Error(w, fmt.Sprintf("ML pipeline error: %s", string(body)), response.StatusCode)
		return
	}

	// 6. Store in database
	if err := db.AddDocument(db_handle, auth_result.ID, filename, storage_name); err != nil {
		http.Error(w, fmt.Sprintf("Database error: %v", err), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Upload successful"))
}

// validate_legal_pdf checks if a file is a valid PDF document.
//
// Parameters:
//   - filename: Original file name (checks extension)
//   - data: File content (checks header signature)
//
// Returns:
//   - error: Validation failure reason
//
// Validation rules:
//   - Must have .pdf extension
//   - Must start with %PDF-1. or %PDF-2. signature
func validate_legal_pdf(filename string, data []byte) error {
	if filename[len(filename)-4:] != ".pdf" {
		return errors.New("detected unsupported file format")
	}
	if !(bytes.HasPrefix(data, []byte("%PDF-1.")) || bytes.HasPrefix(data, []byte("%PDF-2."))) {
		return errors.New("invalid signature")
	}

	return nil
}

// create_storage_name generates a unique storage identifier for documents.
//
// Parameters:
//   - id: User ID (prefix)
//   - name: Original filename (used in hash)
//
// Returns:
//   - string: Unique name format "userID/sha256hash"
//
// Security:
//   - Uses SHA-256 hashing of userID+filename
//   - Returns hex-encoded string
func create_storage_name(id int64, name string) string {
	var int_str string = strconv.FormatInt(id, 10)
	var hashed_name [32]byte = sha256.Sum256([]byte(int_str + name))
	var hashed_name_str string = hex.EncodeToString(hashed_name[:])
	return fmt.Sprintf("%s/%s", int_str, hashed_name_str)
}

// MessageUpload handles message submission to processing pipeline.
//
// Process flow:
//  1. Reads message content from request body
//  2. Forwards to message processing service
//  3. Returns service response
//
// Responses:
//   - 200 OK: Message processed successfully
//   - 400 Bad Request: Invalid message format
//   - 500 Internal ServerError: Processing failure
//
// Note:
//   - Propagates errors from processing pipeline
//   - Requires valid authentication
func MessageUpload(auth_result auth.AuthorizationResult, w http.ResponseWriter, r *http.Request) {
	data, err := io.ReadAll(r.Body)
	defer r.Body.Close()

	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	response, err := UploadMessage(auth_result.ID, data)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(response.Body) // Read error response if available
		http.Error(w, fmt.Sprintf("ML pipeline error: %s", string(body)), response.StatusCode)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Success"))
}
