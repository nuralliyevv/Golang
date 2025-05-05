package main

import (
	"log"
	"net/http"
	"habit-tracker/tracker-service/internal/habit"
	"github.com/gorilla/mux"
)

func main() {
	// Initialize router
	router := mux.NewRouter()

	// Habit routes
	router.HandleFunc("/habits", habit.HabitsHandler).Methods("POST", "GET")
	router.HandleFunc("/habits/{id}/track", habit.TrackHandler).Methods("POST")
	router.HandleFunc("/habits/{id}/stats", habit.StatsHandler).Methods("GET")
	router.PathPrefix("/motivation").HandlerFunc(habit.MotivationHandler).Methods("GET")

	// Start server
	log.Println("Starting Tracker Service on port 8081")
	log.Fatal(http.ListenAndServe(":8081", router))
} 