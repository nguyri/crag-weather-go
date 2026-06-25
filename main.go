package main

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

type Response struct {
	Message string `json:"message"`
	Status  string `json:"status"`
}

func main() {
	r := chi.NewRouter()

	// Helpful built-in middleware for development
	r.Use(middleware.Logger)    // Logs every incoming request to your terminal
	r.Use(middleware.Recoverer) // Prevents the server from crashing if there's a panic

	// Define our status endpoint
	r.Get("/api/status", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		
		res := Response{
			Message: "Crag Weather API running locally on Mac!",
			Status:  "optimal",
		}
		
		json.NewEncoder(w).Encode(res)
	})

	println("Local server starting on http://localhost:8080")
	http.ListenAndServe(":8080", r)
}