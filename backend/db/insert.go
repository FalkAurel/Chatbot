package db

import (
	"database/sql"
	"log"
)

const insertUser string = `
INSERT INTO users (name, password, email, is_admin, is_premium)
VALUES (?, ?, ?, ?, ?)
`

// AddUser inserts a new user record into the database.
//
// The function requires a complete User struct containing:
//   - Name: User's full name
//   - Password: Hashed password string
//   - Email: Unique email address (used as identifier)
//   - IsAdmin: Boolean admin flag
//   - IsPremium: Boolean premium status flag
//
// Returns:
//   - error: nil on success, or:
//   - sql.ErrNoRows if email already exists (preventing duplication)
//   - Other database errors for connection/query failures
//
// Note:
//   - Email must be unique (enforced by database constraints)
//   - Password should be hashed before calling this function
//   - Logs the addition attempt including name and email
func AddUser(db *sql.DB, user User) error {
	log.Printf("Adding user %s with email %s", user.Name, user.Email)
	_, err := db.Exec(
		insertUser,
		user.Name,
		user.Password,
		user.Email,
		user.IsAdmin,
		user.IsPremium,
	)

	return err
}

const createSignupRequest string = `
INSERT INTO signup_requests (name, password, email)
VALUES ($1, $2, $3)
`

// AddSignupRequest adds a new signup request to the database.
//
// It inserts a new record into the signup_requests table with the following fields:
//
//	{
//		"name":     string,  // User's full name
//		"password": string,  // Hashed password
//		"email":    string   // User's email address (must be unique)
//	}
//
// Parameters:
//   - db: Database connection handle
//   - r: SignupRequest containing user credentials
//
// Returns:
//   - error: Database operation error if insertion fails, nil on success
//
// Note:
//   - The password should be hashed before calling this function
//   - Email uniqueness should be verified before insertion
func AddSignupRequest(db *sql.DB, r SignupRequest) error {
	_, err := db.Exec(createSignupRequest, r.Name, r.Password, r.Email)
	return err
}

const createPreprompt string = `
INSERT INTO prompts (user_id, prompt)
VALUES ($1, $2)
`

func AddPrompt(db *sql.DB, id int64, pre_prompt string) error {
	_, err := db.Exec(createPreprompt, id, pre_prompt)
	return err
}

const createPrePromptHistory string = `
INSERT INTO preprompt_history (user_id, prompts)
VALUES ($1, $2)
`

// func AddToPrePromptHistory(db *sql.DB, id int64, pre_prompt string) error {
// 	_, err := db.Exec(createPrePromptHistory, id, nil)
// }
