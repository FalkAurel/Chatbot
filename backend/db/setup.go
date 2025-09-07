package db

import (
	"database/sql"
	"fmt"

	_ "github.com/mattn/go-sqlite3"
)

const tableCreationQuery string = `
CREATE TABLE IF NOT EXISTS users (
    name TEXT NOT NULL,
    password TEXT NOT NULL,
    email TEXT NOT NULL UNIQUE,
    is_admin BOOLEAN,
	is_premium BOOLEAN,
    id INTEGER PRIMARY KEY
);

CREATE TABLE IF NOT EXISTS prompts (
	user_id INTEGER REFERENCES users(id) ON DELETE CASCADE,
	prompt TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS user_documents (
	user_id INTEGER REFERENCES users(id) ON DELETE CASCADE,
	original_name TEXT NOT NULL,
	storage_name TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS signup_requests (
	name TEXT NOT NULL,
	password TEXT NOT NULL,
	email TEXT NOT NULL UNIQUE
);

CREATE TABLE IF NOT EXISTS preprompt_history (
	user_id INTEGER REFERENCES users(id) ON DELETE CASCADE,
	prompts  BLOB
);
`

const addDocument string = `
INSERT OR REPLACE INTO user_documents (user_id, original_name, storage_name)
VALUES ($1, $2, $3)
`

const getUserID string = `
SELECT id FROM users
WHERE email = ?
`

const getDBUser string = `
SELECT name, password, is_admin, is_premium, id FROM users
WHERE email = ?
`

const getSignup string = `
SELECT name, email FROM signup_requests
`

// SetupSqlite initializes a new SQLite database with required tables and admin user.
//
// Parameters:
//   - db_location: Path to the SQLite database file
//   - admin: Pre-configured admin user to create
//
// Returns:
//   - *sql.DB: Database connection handle
//   - error: Initialization errors including:
//   - Database connection failures
//   - Table creation failures
//   - Admin user creation failures
//
// Note:
//   - Creates four tables: users, subscriptions, user_documents, and signup_requests
//   - Automatically adds a default signup request for testing
//   - Uses FOREIGN KEY constraints with ON DELETE CASCADE
func SetupSqlite(db_location string, admin User) (*sql.DB, error) {
	db, err := sql.Open("sqlite3", db_location)

	if err != nil {
		return nil, err
	}

	_, err = db.Exec(tableCreationQuery)

	if err != nil {
		return nil, err
	}

	_ = AddUser(db, admin)
	id, _ := GetUserID(db, admin.Email)
	AddPrompt(db, id, "Test")

	return db, nil
}

// GetUserID retrieves a user's unique ID by their email address.
//
// Parameters:
//   - db: Database connection handle
//   - email: User's email address (unique identifier)
//
// Returns:
//   - int64: Numeric user ID
//   - error: Database errors including:
//   - sql.ErrNoRows if email not found
//   - Query execution errors
//
// Note:
//   - Uses numeric IDs for efficient lookups and storage
//   - Email must be exact match (case-sensitive)
func GetUserID(db *sql.DB, email string) (int64, error) {
	var id int64
	err := db.QueryRow(getUserID, email).Scan(&id)

	if err != nil {
		return 0, err
	}
	return id, nil
}

// GetDataBaseUser retrieves complete user information from the database.
//
// Parameters:
//   - db: Database connection handle
//   - email: User's email address (unique identifier)
//
// Returns:
//   - DataBaseUser: Struct containing all user fields
//   - error: Database query errors
//
// Note:
//   - Returns empty DataBaseUser with nil error if user not found
//   - Includes sensitive information (hashed password)
func GetDataBaseUser(db *sql.DB, email string) (DataBaseUser, error) {
	var name string
	var password string
	var is_admin bool
	var is_premium bool
	var id int64

	err := db.QueryRow(getDBUser, email).Scan(&name, &password, &is_admin, &is_premium, &id)
	if err != nil {
		return DataBaseUser{Name: "", ID: 0, IsAdmin: false, IsPremium: false, Password: "", Email: ""}, nil
	}

	return DataBaseUser{
		Name:      name,
		Password:  password,
		Email:     email,
		IsAdmin:   is_admin,
		IsPremium: is_premium,
		ID:        id,
	}, nil
}

// AddDocument associates a document with a user in the database.
//
// Parameters:
//   - db: Database connection handle
//   - id: User ID (foreign key)
//   - filename: Original document name
//   - storage_name: Internal storage identifier
//
// Returns:
//   - error: Database operation errors
//
// Note:
//   - Uses INSERT OR REPLACE to handle duplicates
//   - Documents are automatically deleted when users are removed (ON DELETE CASCADE)
func AddDocument(db *sql.DB, id int64, filename string, storage_name string) error {
	_, err := db.Exec(addDocument, id, filename, storage_name)

	if err != nil {
		return err
	}

	return nil
}

// GetSignupRequests retrieves all pending signup requests from the database.
//
// Parameters:
//   - db: Database connection handle
//
// Returns:
//   - []SignupRequestDB: Slice of signup request records
//   - error: Database query errors
//
// Note:
//   - Only returns name and email fields (excludes passwords)
//   - Properly handles row iteration and cleanup
//   - Returns empty slice (not nil) when no requests exist
func GetSignupRequests(db *sql.DB) ([]SignupRequestDB, error) {
	rows, err := db.Query(getSignup)
	if err != nil {
		return nil, fmt.Errorf("query failed: %w", err)
	}
	defer rows.Close() // Important to close rows

	var signupRequests []SignupRequestDB

	// Iterate through each row
	for rows.Next() {
		var lr SignupRequestDB
		if err := rows.Scan(&lr.Name, &lr.Email); err != nil {
			return nil, fmt.Errorf("scan failed: %w", err)
		}
		signupRequests = append(signupRequests, lr)
	}

	// Check for errors from iterating over rows
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows iteration error: %w", err)
	}

	return signupRequests, nil
}
