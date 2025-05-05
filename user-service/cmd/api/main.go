package main

import (
	"log"
	"net/http"
	"habit-tracker/user-service/internal/user"
)

func main() {
	// Initialize database
	if err := user.InitDB(); err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/register", user.RegisterHandler)
	mux.HandleFunc("/login", user.LoginHandler)
	mux.HandleFunc("/me", user.MeHandler)

	log.Println("Starting User Service on port 8080")
	if err := http.ListenAndServe(":8080", mux); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
} 