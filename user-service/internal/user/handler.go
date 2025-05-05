package user

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
)

func RegisterHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if r.Method != http.MethodPost {
		http.Error(w, `{"error": "Method not allowed"}`, http.StatusMethodNotAllowed)
		return
	}

	var req RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error": "Invalid request body"}`, http.StatusBadRequest)
		return
	}

	if req.Username == "" {
		http.Error(w, `{"error": "Username is required"}`, http.StatusBadRequest)
		return
	}
	if req.Email == "" {
		http.Error(w, `{"error": "Email is required"}`, http.StatusBadRequest)
		return
	}
	if req.Password == "" {
		http.Error(w, `{"error": "Password is required"}`, http.StatusBadRequest)
		return
	}

	// Create new user
	user := User{
		Username: req.Username,
		Email:    req.Email,
		Password: req.Password,
	}

	// Save to database
	if err := saveUser(user); err != nil {
		http.Error(w, fmt.Sprintf(`{"error": "Failed to create user: %v"}`, err), http.StatusInternalServerError)
		return
	}

	// Return success message
	response := RegisterResponse{
		Message: "Registration successful",
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(response)
}

func LoginHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if r.Method != http.MethodPost {
		http.Error(w, `{"error": "Method not allowed"}`, http.StatusMethodNotAllowed)
		return
	}

	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error": "Invalid request body"}`, http.StatusBadRequest)
		return
	}

	user, err := getUserByUsername(req.Username)
	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, `{"error": "User not found"}`, http.StatusNotFound)
		} else {
			http.Error(w, fmt.Sprintf(`{"error": "Failed to get user: %v"}`, err), http.StatusInternalServerError)
		}
		return
	}

	if user.Password != req.Password {
		http.Error(w, `{"error": "Invalid password"}`, http.StatusUnauthorized)
		return
	}

	// Update last login time
	if err := updateLastLogin(req.Username); err != nil {
		http.Error(w, fmt.Sprintf(`{"error": "Failed to update login time: %v"}`, err), http.StatusInternalServerError)
		return
	}

	response := LoginResponse{
		Message: "Login successful",
		User: User{
			ID:       user.ID,
			Username: user.Username,
			Email:    user.Email,
		},
	}

	json.NewEncoder(w).Encode(response)
}

func MeHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if r.Method != http.MethodGet {
		http.Error(w, `{"error": "Method not allowed"}`, http.StatusMethodNotAllowed)
		return
	}

	user, err := getLastLoggedInUser()
	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, `{"error": "No user is currently logged in"}`, http.StatusUnauthorized)
		} else {
			http.Error(w, fmt.Sprintf(`{"error": "Failed to get last logged in user: %v"}`, err), http.StatusInternalServerError)
		}
		return
	}

	response := UserResponse{
		ID:       user.ID,
		Username: user.Username,
		Email:    user.Email,
	}

	json.NewEncoder(w).Encode(response)
} 