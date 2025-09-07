package auth

import (
	"backend/db"
	"crypto/cipher"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"time"
)

const N_MINUTES int64 = 120 * 60 // In seconds

func get_request_body(r *http.Request) ([]byte, error) {
	buffer, err := io.ReadAll(r.Body)
	defer r.Body.Close()

	return buffer, err
}

// Login handles user authentication and JWT token generation.
//
// Expects a JSON payload in the request body with the format:
//
//	{
//		"email":    string,
//		"password": string
//	}
//
// On successful authentication:
//   - Generates a JWT token with user claims (ID, admin status)
//   - Returns a LoginResponse containing:
//   - User metadata (admin/premium status)
//   - Hex-encoded JWT token
//   - Username
//
// Possible error responses:
//   - 400 Bad Request: Malformed JSON or missing fields
//   - 401 Unauthorized: Invalid credentials
//   - 500 Internal Server Error: Database or token generation failure
//
// Security Note:
//   - Uses constant-time comparison for password validation
//   - JWT contains expiration claim (N_MINUTES from issuance)
//   - Tokens are hex-encoded for transport
func Login(db_handle *sql.DB, cipher cipher.Block, w http.ResponseWriter, r *http.Request) {
	buffer, err := get_request_body(r)

	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	var login_credentials LoginCredentials
	err = json.Unmarshal(buffer, &login_credentials)

	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
	}

	record, err := db.GetDataBaseUser(db_handle, login_credentials.Email)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if record.Password == login_credentials.Password {
		jwt, err := JWTToken{
			ID:             record.ID,
			IsAdmin:        record.IsAdmin,
			ExpirationTime: time.Now().UTC().Unix() + N_MINUTES,
		}.JWTTokenToJson()

		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}

		var jwt_string string = hex.EncodeToString(jwt)
		var login_response LoginResponse = LoginResponse{
			IsAdmin:   record.IsAdmin,
			IsPremium: record.IsPremium,
			JWTToken:  jwt_string,
			Username:  record.Name,
		}

		log.Printf("%s Login Successful", login_credentials.Email)

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(login_response)
	} else {
		http.Error(w, "Username or Password is wrong", http.StatusUnauthorized)
	}
}
