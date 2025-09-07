package db

import (
	"database/sql"
	"fmt"
)

const promoteUserQuery string = `
UPDATE users
SET is_admin = TRUE
WHERE email = ?
`

func PromoteUser(db *sql.DB, email string) error {
	fmt.Printf("Attempting to promote email: '%s'\n", email) // Debug log
	result, err := db.Exec(promoteUserQuery, email)
	if err != nil {
		return fmt.Errorf("promote failed: %w", err)
	}
	rowsAffected, _ := result.RowsAffected()
	fmt.Printf("Rows affected: %d\n", rowsAffected) // Debug log
	if rowsAffected == 0 {
		return fmt.Errorf("no user found with email '%s'", email)
	}
	return nil
}

const updatePrompt string = `
UPDATE prompts
SET prompt = $1
WHERE user_id = $2
`

// UpdatePrompt updates the stored prompt for a specific user in the database.
//
// Parameters:
//   - db:      Database connection handle (*sql.DB)
//   - id:      User ID to update (int64)
//   - prompt:  New prompt text to store (string)
//
// Returns:
//   - error:   Any database error that occurred during the update operation
//     Returns nil on successful update
func UpdatePrompt(db *sql.DB, id int64, prompt string) error {
	_, err := db.Exec(updatePrompt, prompt, id)
	return err
}
