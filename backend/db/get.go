package db

import (
	"database/sql"
	"fmt"
)

const userRetrivalQuery string = `
SELECT email, is_admin
FROM users
`

func GetUsers(db *sql.DB) ([]UserInfo, error) {
	rows, err := db.Query(userRetrivalQuery)

	if err != nil {
		return nil, err
	}

	var users []UserInfo

	for rows.Next() {
		var usr UserInfo
		if err := rows.Scan(&usr.Email, &usr.IsAdmin); err != nil {
			return nil, fmt.Errorf("scan failed: %w", err)
		}
		users = append(users, usr)
	}

	return users, nil
}

const getDocumentsQuery string = `
SELECT original_name, storage_name
FROM user_documents
WHERE user_id = ?
`

func GetDocuments(db *sql.DB, user_id int64) ([]DocumentRecord, error) {
	rows, err := db.Query(getDocumentsQuery, user_id)

	if err != nil {
		return nil, err
	}

	var documents []DocumentRecord

	for rows.Next() {
		var document DocumentRecord
		if err := rows.Scan(&document.OriginalName, &document.StorageName); err != nil {
			return nil, fmt.Errorf("scan failed: %w", err)
		}

		documents = append(documents, document)
	}

	return documents, nil
}

const getPrompt string = `
SELECT prompt
FROM prompts
WHERE user_id = ?
`

func GetPrompt(db *sql.DB, user_id int64) (string, error) {
	var prompt string

	err := db.QueryRow(getPrompt, user_id).Scan(&prompt)

	return prompt, err
}
