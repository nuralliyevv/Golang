package habit

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"
)

type Habit struct {
	ID          int64     `json:"id"`
	UserID      int64     `json:"user_id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	CreatedAt   time.Time `json:"created_at"`
}

type HabitRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

type HabitResponse struct {
	Message string `json:"message"`
	Habit   *Habit `json:"habit,omitempty"`
}

type TrackRecord struct {
	ID        int64     `json:"id"`
	HabitID   int64     `json:"habit_id"`
	UserID    int64     `json:"user_id"`
	Completed bool      `json:"completed"`
	Date      time.Time `json:"date"`
}

type StatsResponse struct {
	HabitName     string `json:"habit_name"`
	TotalTrackings int    `json:"total_trackings"`
	CompletedDays  int    `json:"completed_days"`
	SkippedDays    int    `json:"skipped_days"`
	FirstTracked   string `json:"first_tracked"`
	LastTracked    string `json:"last_tracked"`
}

type MotivationResponse struct {
	Quote     string `json:"quote"`
	Author    string `json:"author"`
	Category  string `json:"category"`
}

// In-memory storage
var habits = make(map[int64]map[int64]*Habit)  // userID -> habitID -> Habit
var trackRecords = make(map[int64]map[int64][]*TrackRecord) // userID -> habitID -> []TrackRecord
var nextHabitIDs = make(map[int64]int64) // userID -> nextHabitID
var nextTrackID int64 = 1

func init() {
	initDB()
}

// Get the last logged-in user from the User Service
func getLastLoggedInUser() (int64, error) {
	// In a real application, this would be a proper service call
	// For now, we'll use a simple HTTP request
	resp, err := http.Get("http://localhost:8080/me")
	if err != nil {
		return 0, fmt.Errorf("failed to get last logged in user: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return 0, fmt.Errorf("failed to get last logged in user: status %d", resp.StatusCode)
	}

	var user struct {
		ID int64 `json:"id"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&user); err != nil {
		return 0, fmt.Errorf("failed to decode user response: %v", err)
	}

	return user.ID, nil
}

func HabitsHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	userID, err := getLastLoggedInUser()
	if err != nil {
		http.Error(w, fmt.Sprintf(`{"error": "%v"}`, err), http.StatusUnauthorized)
		return
	}

	switch r.Method {
	case http.MethodPost:
		createHabit(w, r, userID)
	case http.MethodGet:
		listHabits(w, r, userID)
	default:
		http.Error(w, `{"error": "Method not allowed"}`, http.StatusMethodNotAllowed)
	}
}

func TrackHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	userID, err := getLastLoggedInUser()
	if err != nil {
		http.Error(w, fmt.Sprintf(`{"error": "%v"}`, err), http.StatusUnauthorized)
		return
	}

	if r.Method != http.MethodPost {
		http.Error(w, `{"error": "Method not allowed"}`, http.StatusMethodNotAllowed)
		return
	}

	// Extract habit ID from URL
	idStr := r.URL.Path[len("/habits/"):]
	idStr = idStr[:len(idStr)-len("/track")]
	habitID, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		http.Error(w, `{"error": "Invalid habit ID"}`, http.StatusBadRequest)
		return
	}

	// Check if habit exists and belongs to current user
	userHabits, exists := habits[userID]
	if !exists {
		http.Error(w, `{"error": "Habit not found"}`, http.StatusNotFound)
		return
	}
	habit, exists := userHabits[habitID]
	if !exists {
		http.Error(w, `{"error": "Habit not found"}`, http.StatusNotFound)
		return
	}

	// Initialize user's track records map if it doesn't exist
	if _, exists := trackRecords[userID]; !exists {
		trackRecords[userID] = make(map[int64][]*TrackRecord)
	}

	// Create new track record
	record := &TrackRecord{
		HabitID:   habitID,
		UserID:    userID,
		Completed: true,
		Date:      time.Now(),
	}

	// Save to database
	err = saveTrackRecord(record)
	if err != nil {
		http.Error(w, fmt.Sprintf(`{"error": "Failed to save track record: %v"}`, err), http.StatusInternalServerError)
		return
	}

	// Update in-memory storage
	trackRecords[userID][habitID] = append(trackRecords[userID][habitID], record)
	nextTrackID = record.ID + 1

	// Create response with formatted dates
	response := map[string]interface{}{
		"message": "Habit tracked successfully",
		"habit": map[string]interface{}{
			"id":          habit.ID,
			"name":        habit.Name,
			"description": habit.Description,
			"created_at":  habit.CreatedAt.Format("2006-01-02 15:04:05"),
		},
		"tracked_at": record.Date.Format("2006-01-02 15:04:05"),
	}

	json.NewEncoder(w).Encode(response)
}

func StatsHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	userID, err := getLastLoggedInUser()
	if err != nil {
		http.Error(w, fmt.Sprintf(`{"error": "%v"}`, err), http.StatusUnauthorized)
		return
	}

	if r.Method != http.MethodGet {
		http.Error(w, `{"error": "Method not allowed"}`, http.StatusMethodNotAllowed)
		return
	}

	// Extract habit ID from URL
	idStr := r.URL.Path[len("/habits/"):]
	idStr = idStr[:len(idStr)-len("/stats")]
	habitID, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		http.Error(w, `{"error": "Invalid habit ID"}`, http.StatusBadRequest)
		return
	}

	// Check if habit exists and belongs to current user
	userHabits, exists := habits[userID]
	if !exists {
		http.Error(w, `{"error": "Habit not found"}`, http.StatusNotFound)
		return
	}
	habit, exists := userHabits[habitID]
	if !exists {
		http.Error(w, `{"error": "Habit not found"}`, http.StatusNotFound)
		return
	}

	// Load track records from database
	userTrackRecords, err := loadTrackRecords(userID)
	if err != nil {
		http.Error(w, fmt.Sprintf(`{"error": "Failed to load track records: %v"}`, err), http.StatusInternalServerError)
		return
	}

	// Update in-memory storage
	trackRecords[userID] = userTrackRecords
	records := userTrackRecords[habitID]

	// Calculate stats
	totalTrackings := len(records)
	completedDays := 0
	skippedDays := 0
	var firstTracked, lastTracked time.Time

	if len(records) > 0 {
		firstTracked = records[0].Date
		lastTracked = records[0].Date
	}

	// Track unique days and their status
	dayStatus := make(map[string]bool) // date string -> completed
	for _, record := range records {
		dateStr := record.Date.Format("2006-01-02")
		if record.Date.Before(firstTracked) {
			firstTracked = record.Date
		}
		if record.Date.After(lastTracked) {
			lastTracked = record.Date
		}
		if record.Completed {
			dayStatus[dateStr] = true
		} else {
			dayStatus[dateStr] = false
		}
	}

	// Count completed and skipped days
	for _, completed := range dayStatus {
		if completed {
			completedDays++
		} else {
			skippedDays++
		}
	}

	response := StatsResponse{
		HabitName:     habit.Name,
		TotalTrackings: totalTrackings,
		CompletedDays:  completedDays,
		SkippedDays:    skippedDays,
		FirstTracked:   firstTracked.Format("2006-01-02 15:04:05"),
		LastTracked:    lastTracked.Format("2006-01-02 15:04:05"),
	}

	json.NewEncoder(w).Encode(response)
}

func MotivationHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if r.Method != http.MethodGet {
		http.Error(w, `{"error": "Method not allowed"}`, http.StatusMethodNotAllowed)
		return
	}

	// Fetch quote from ZenQuotes API
	resp, err := http.Get("https://zenquotes.io/api/random")
	if err != nil {
		http.Error(w, `{"error": "Failed to fetch motivation quote"}`, http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	var quotes []struct {
		Quote  string `json:"q"`
		Author string `json:"a"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&quotes); err != nil || len(quotes) == 0 {
		http.Error(w, `{"error": "Failed to decode motivation quote"}`, http.StatusInternalServerError)
		return
	}

	// Create response with the fetched quote
	response := MotivationResponse{
		Quote:     quotes[0].Quote,
		Author:    quotes[0].Author,
		Category:  "Motivation",
	}

	json.NewEncoder(w).Encode(response)
}

func createHabit(w http.ResponseWriter, r *http.Request, userID int64) {
	var req HabitRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error": "Invalid request body"}`, http.StatusBadRequest)
		return
	}

	if req.Name == "" {
		http.Error(w, `{"error": "Name is required"}`, http.StatusBadRequest)
		return
	}

	// Initialize user's habit map if it doesn't exist
	if _, exists := habits[userID]; !exists {
		habits[userID] = make(map[int64]*Habit)
		nextHabitIDs[userID] = 1
	}

	habit := &Habit{
		UserID:      userID,
		Name:        req.Name,
		Description: req.Description,
		CreatedAt:   time.Now(),
	}

	// Save to database
	err := saveHabit(habit)
	if err != nil {
		http.Error(w, fmt.Sprintf(`{"error": "Failed to save habit: %v"}`, err), http.StatusInternalServerError)
		return
	}

	// Update in-memory storage
	habits[userID][habit.ID] = habit
	nextHabitIDs[userID] = habit.ID + 1

	// Create response with formatted date
	response := map[string]interface{}{
		"message": "Habit created successfully",
		"habit": map[string]interface{}{
			"id":          habit.ID,
			"name":        habit.Name,
			"description": habit.Description,
			"created_at":  habit.CreatedAt.Format("2006-01-02 15:04:05"),
		},
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(response)
}

func listHabits(w http.ResponseWriter, r *http.Request, userID int64) {
	// Load habits from database
	userHabits, err := loadHabits(userID)
	if err != nil {
		http.Error(w, fmt.Sprintf(`{"error": "Failed to load habits: %v"}`, err), http.StatusInternalServerError)
		return
	}

	// Update in-memory storage
	habits[userID] = userHabits

	habitList := make([]*Habit, 0, len(userHabits))
	for _, habit := range userHabits {
		habitList = append(habitList, habit)
	}

	json.NewEncoder(w).Encode(habitList)
}

func ClearAllHabits() {
	// Clear all maps
	habits = make(map[int64]map[int64]*Habit)
	trackRecords = make(map[int64]map[int64][]*TrackRecord)
	nextHabitIDs = make(map[int64]int64)
	// Reset counters
	nextTrackID = 1
} 