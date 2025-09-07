package auth

import (
	"encoding/json"
)

type LoginCredentials struct {
	Email    string
	Password string
}

type LoginResponse struct {
	IsAdmin   bool
	IsPremium bool
	JWTToken  string
	Username  string
}

type JWTToken struct {
	ExpirationTime int64
	ID             int64
	IsAdmin        bool
}

// Returns the authorization data
type AuthorizationResult struct {
	IsAdmin bool
	ID      int64
}

func (jwt JWTToken) JWTTokenToJson() ([]byte, error) {
	return json.Marshal(jwt)
}

func JsonToJWTToken(buffer []byte) (JWTToken, error) {
	var data JWTToken
	err := json.Unmarshal(buffer, &data)

	if err != nil {
		return data, err
	}
	return data, nil
}
