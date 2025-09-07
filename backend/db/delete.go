package db

import (
	"database/sql"
	"fmt"
)

const deletesignupRequest string = `
DELETE FROM signup_requests
WHERE email = ?
RETURNING name, password, email
`

// DeleteSignupRequest removes a signup request from the database by email.
//
// Performs the following operations:
//  1. Deletes the record from signup_requests table
//  2. Returns the deleted row data if successful
//
// Parameters:
//   - db: Database connection handle
//   - email: Unique email identifying the request to delete
//
// Returns:
//   - SignupRequest: The deleted request data if found
//   - error: Detailed failure reason including:
//   - sql.ErrNoRows if no matching request exists
//   - Database errors for other failures
//
// Note:
//   - Email is used as the unique identifier
//   - Returns empty SignupRequest and error if deletion fails
func DeleteSignupRequest(db *sql.DB, email string) (SignupRequest, error) {
	var request SignupRequest

	// Use QueryRow since we expect only one row (email should be unique)
	err := db.QueryRow(deletesignupRequest, email).Scan(
		&request.Name,
		&request.Password,
		&request.Email,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return SignupRequest{}, fmt.Errorf("no signup request found for email: %s", email)
		}
		return SignupRequest{}, fmt.Errorf("failed to delete signup request: %w", err)
	}

	return request, nil
}

const deleteUser string = `
DELETE FROM users
WHERE email = ?
RETURNING id, name, password, email, is_admin, is_premium
`

func DeleteUser(db *sql.DB, email string) (DataBaseUser, error) {
	var user DataBaseUser

	err := db.QueryRow(deleteUser, email).Scan(
		&user.ID,
		&user.Name,
		&user.Password,
		&user.Email,
		&user.IsAdmin,
		&user.IsPremium,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return DataBaseUser{}, fmt.Errorf("no signup request found for email: %s", email)
		}
		return DataBaseUser{}, fmt.Errorf("failed to delete signup request: %w", err)
	}

	return user, nil
}

const deleteDocument string = `
DELETE FROM user_documents
WHERE user_id = $1 AND storage_name = $2
`

func DeleteDocument(db *sql.DB, storage_name string, id int64) error {
	result, err := db.Exec(deleteDocument, id, storage_name)
	if err != nil {
		return fmt.Errorf("error executing delete: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("error getting rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("no documents deleted - check if user_id=%d and storage_name=%s exist", id, storage_name)
	}

	return nil
}
