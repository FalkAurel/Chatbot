package auth

import (
	"encoding/hex"
	"errors"
	"time"
)

// Authorization verifies and validates a JWT token from the given hex-encoded string.
//
// Returns an AuthorizationResult with user metadata if successful. Possible error conditions:
//   - Invalid hex encoding
//   - Malformed JWT token
//   - Expired token
//   - Invalid signature
//
// Any error indicates the request is not authorized, suggesting either:
//   - Expired session
//   - Tampered credentials
//   - Invalid token format
//
// Returns:
//   - AuthorizationResult: Contains user ID and admin status on success
//   - error: Detailed authorization failure reason
func Authorization(buffer_string string) (AuthorizationResult, error) {
	buffer, err := hex.DecodeString(buffer_string)

	if err != nil {
		return AuthorizationResult{IsAdmin: false, ID: 0}, err
	}

	jwt, err := JsonToJWTToken(buffer)
	if err != nil {
		return AuthorizationResult{IsAdmin: false, ID: 0}, err
	}

	if jwt.ExpirationTime < time.Now().Unix() {
		return AuthorizationResult{IsAdmin: false, ID: 0}, errors.New("token is expired. Please refresh the browser")
	}

	return AuthorizationResult{IsAdmin: jwt.IsAdmin, ID: jwt.ID}, nil
}

// AdminAuthorization performs authorization specifically requiring admin privileges.
//
// This wraps the standard Authorization check and adds an additional admin privilege
// verification. It will only succeed if:
//   - The base Authorization succeeds
//   - The authenticated user has IsAdmin flag set
//
// Returns:
//   - AuthorizationResult: User metadata if authorized as admin
//   - error: nil (always returns nil error, check IsAdmin flag instead)
//
// Note: Unlike Authorization, this intentionally swallows errors to prevent
// leaking information about authorization failures.
func AdminAuthorization(buffer_string string) (AuthorizationResult, error) {
	auth_result, err := Authorization(buffer_string)

	if err != nil {
		return AuthorizationResult{IsAdmin: false, ID: 0}, nil
	}

	if auth_result.IsAdmin {
		return auth_result, nil
	} else {
		return AuthorizationResult{IsAdmin: false, ID: 0}, nil
	}
}
