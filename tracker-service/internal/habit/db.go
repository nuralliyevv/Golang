package habit

import (
	"database/sql"
	"fmt"
	"log"

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

func initDB() {
	psqlInfo := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		host, port, user, password, dbname)

	var err error
	db, err = sql.Open("postgres", psqlInfo)
	if err != nil {
		log.Fatal(err)
	}

	err = db.Ping()
	if err != nil {
		log.Fatal(err)
	}

	createTables()
}

func createTables() {
	// Create habits table
	_, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS habits (
			id SERIAL PRIMARY KEY,
			user_id BIGINT NOT NULL,
			name VARCHAR(255) NOT NULL,
			description TEXT,
			created_at TIMESTAMP NOT NULL
		)
	`)
	if err != nil {
		log.Fatal(err)
	}

	// Create track_records table
	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS track_records (
			id SERIAL PRIMARY KEY,
			habit_id BIGINT NOT NULL,
			user_id BIGINT NOT NULL,
			completed BOOLEAN NOT NULL,
			date TIMESTAMP NOT NULL,
			FOREIGN KEY (habit_id) REFERENCES habits(id) ON DELETE CASCADE
		)
	`)
	if err != nil {
		log.Fatal(err)
	}
}

func saveHabit(habit *Habit) error {
	query := `
		INSERT INTO habits (user_id, name, description, created_at)
		VALUES ($1, $2, $3, $4)
		RETURNING id
	`
	err := db.QueryRow(query, habit.UserID, habit.Name, habit.Description, habit.CreatedAt).Scan(&habit.ID)
	if err != nil {
		return err
	}
	return nil
}

func saveTrackRecord(record *TrackRecord) error {
	query := `
		INSERT INTO track_records (habit_id, user_id, completed, date)
		VALUES ($1, $2, $3, $4)
		RETURNING id
	`
	err := db.QueryRow(query, record.HabitID, record.UserID, record.Completed, record.Date).Scan(&record.ID)
	if err != nil {
		return err
	}
	return nil
}

func loadHabits(userID int64) (map[int64]*Habit, error) {
	habits := make(map[int64]*Habit)
	query := `
		SELECT id, name, description, created_at
		FROM habits
		WHERE user_id = $1
	`
	rows, err := db.Query(query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var habit Habit
		err := rows.Scan(&habit.ID, &habit.Name, &habit.Description, &habit.CreatedAt)
		if err != nil {
			return nil, err
		}
		habit.UserID = userID
		habits[habit.ID] = &habit
	}

	return habits, nil
}

func loadTrackRecords(userID int64) (map[int64][]*TrackRecord, error) {
	records := make(map[int64][]*TrackRecord)
	query := `
		SELECT id, habit_id, completed, date
		FROM track_records
		WHERE user_id = $1
	`
	rows, err := db.Query(query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var record TrackRecord
		err := rows.Scan(&record.ID, &record.HabitID, &record.Completed, &record.Date)
		if err != nil {
			return nil, err
		}
		record.UserID = userID
		records[record.HabitID] = append(records[record.HabitID], &record)
	}

	return records, nil
}

func getNextHabitID(userID int64) (int64, error) {
	var maxID sql.NullInt64
	query := `
		SELECT MAX(id)
		FROM habits
		WHERE user_id = $1
	`
	err := db.QueryRow(query, userID).Scan(&maxID)
	if err != nil {
		return 0, err
	}
	if !maxID.Valid {
		return 1, nil
	}
	return maxID.Int64 + 1, nil
} 