package user

import (
	"database/sql"
	"fmt"
	_ "github.com/lib/pq"
)

const (
	host     = "localhost"
	port     = 5432
	user     = "postgres"
	password = "postgres"
	dbname   = "habit_tracker"
)

var db *sql.DB

func InitDB() error {
	psqlInfo := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		host, port, user, password, dbname)

	var err error
	db, err = sql.Open("postgres", psqlInfo)
	if err != nil {
		return err
	}

	err = db.Ping()
	if err != nil {
		return err
	}

	// Create users table if it doesn't exist
	createTable := `
	CREATE TABLE IF NOT EXISTS users (
		id SERIAL PRIMARY KEY,
		username VARCHAR(50) UNIQUE NOT NULL,
		email VARCHAR(100) UNIQUE NOT NULL,
		password VARCHAR(100) NOT NULL,
		last_login TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	);`

	_, err = db.Exec(createTable)
	return err
}

func saveUser(user User) error {
	query := `INSERT INTO users (username, email, password) VALUES ($1, $2, $3) RETURNING id`
	return db.QueryRow(query, user.Username, user.Email, user.Password).Scan(&user.ID)
}

func getUserByUsername(username string) (User, error) {
	var user User
	query := `SELECT id, username, email, password FROM users WHERE username = $1`
	err := db.QueryRow(query, username).Scan(&user.ID, &user.Username, &user.Email, &user.Password)
	return user, err
}

func updateLastLogin(username string) error {
	// First check if user exists
	var exists bool
	err := db.QueryRow("SELECT EXISTS(SELECT 1 FROM users WHERE username = $1)", username).Scan(&exists)
	if err != nil {
		return fmt.Errorf("failed to check if user exists: %v", err)
	}
	if !exists {
		return fmt.Errorf("user not found")
	}

	// Update last login time
	query := `UPDATE users SET last_login = CURRENT_TIMESTAMP WHERE username = $1`
	result, err := db.Exec(query, username)
	if err != nil {
		return fmt.Errorf("failed to update login time: %v", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %v", err)
	}
	if rowsAffected == 0 {
		return fmt.Errorf("no rows were updated")
	}

	return nil
}

func getLastLoggedInUser() (User, error) {
	var user User
	query := `SELECT id, username, email, password FROM users ORDER BY last_login DESC LIMIT 1`
	err := db.QueryRow(query).Scan(&user.ID, &user.Username, &user.Email, &user.Password)
	return user, err
} 