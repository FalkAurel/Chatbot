package db

import "database/sql"

const existsEmailInUser string = `
SELECT EXISTS(
	SELECT 1 FROM users
	WHERE email = ?
)
`

const existsEmailinSignUp string = `
SELECT EXISTS(
	SELECT 1 FROM signup_requests
	WHERE email = ?
)
`

// ExistsEmailInUser checks if an email address exists in the users table.
//
// Parameters:
//   - db: Database connection handle
//   - email: Email address to check
//
// Returns:
//   - bool: True if email exists, false otherwise
//   - error: Database error if query fails
//
// Note:
//   - Performs a case-sensitive exact match query
//   - Uses EXISTS() for efficient checking without fetching full records
func ExistsEmailInUser(db *sql.DB, email string) (bool, error) {
	var boolean bool

	err := db.QueryRow(existsEmailInUser, email).Scan(&boolean)

	if err != nil {
		return false, err
	}

	return boolean, nil
}

// ExistsEmailInSignUp checks if an email address exists in the signup_requests table.
//
// Parameters:
//   - db: Database connection handle
//   - email: Email address to check
//
// Returns:
//   - bool: True if email exists in signup requests, false otherwise
//   - error: Database error if query fails
//
// Note:
//   - Useful for preventing duplicate signup requests
//   - Uses the same efficient EXISTS() pattern as ExistsEmailInUser
func ExistsEmailInSignUp(db *sql.DB, email string) (bool, error) {
	var boolean bool

	err := db.QueryRow(existsEmailinSignUp, email).Scan(&boolean)

	if err != nil {
		return false, err
	}

	return boolean, nil
}
