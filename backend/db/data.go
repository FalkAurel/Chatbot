package db

import "encoding/json"

// Defines an insertible user.
// User defines a user object for insertion into the system.
// This represents the basic user information required for account creation.
// Fields:
//   - Name:     The full name of the user (required)
//   - Password: The user's password (stored in plaintext - should be hashed before storage)
//   - Email:    The user's email address (must be unique)
//   - IsAdmin:  Flag indicating administrator privileges (default: false)
type User struct {
	Name      string
	Password  string
	Email     string
	IsAdmin   bool
	IsPremium bool
}

// DataBaseUser represents a complete user record as stored in the database.
// This extends the basic User struct with database-specific fields.
// Fields:
//   - Name:     The full name of the user (indexed)
//   - Password: The hashed password (should never be stored in plaintext)
//   - Email:    The user's email address (unique constraint)
//   - ID:       The auto-incremented primary key from the database
//   - IsAdmin:  Administrator status flag (default: false)
type DataBaseUser struct {
	Name      string
	Password  string
	Email     string
	ID        int64
	IsAdmin   bool
	IsPremium bool
}

// Isolated UserInfo to not reveal sensitive information.
type UserInfo struct {
	Email   string
	IsAdmin bool
}

func CreateUser(name string, password string, email string, is_premium bool) User {
	return User{Name: name, Password: password, Email: email, IsAdmin: false, IsPremium: is_premium}
}

func CreateAdmin(name string, password string, email string) User {
	return User{Name: name, Password: password, Email: email, IsAdmin: true, IsPremium: true}
}

type SignupRequest struct {
	Name     string
	Password string
	Email    string
}

// This type will be returned when requesting all the login requests.
// This ensures that the password of the user is never read out.
// A type is needed to prove this statically at compile time
type SignupRequestDB struct {
	Name  string
	Email string
}

type DocumentRecord struct {
	OriginalName string
	StorageName  string
}

type PreviousPrompts struct {
	prompts [10]string
}

// Adds a prompt, maintaining a rotating cache of the last 10 prompts.
func (p *PreviousPrompts) AddPrompt(prompt string) {
	// Shift all elements left (drop oldest)
	copy(p.prompts[:], p.prompts[1:])
	// Add new prompt at the end
	p.prompts[9] = prompt
}

// Returns a copy of the prompts (safe from external modification).
func (p *PreviousPrompts) GetPrompts() []string {
	return append([]string(nil), p.prompts[:]...)
}

// Serializes the prompts to JSON.
func (p *PreviousPrompts) Serialize() ([]byte, error) {
	return json.Marshal(p.prompts)
}

// Deserializes JSON into an existing PreviousPrompts struct.
func (p *PreviousPrompts) Deserialize(bytes []byte) error {
	return json.Unmarshal(bytes, &p.prompts)
}
