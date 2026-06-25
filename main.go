package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	// Import your newly created providers package
	"crag-weather-go/providers"
)

func main() {
	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	// Initialize our standard HTTP Client
	httpClient := &http.Client{Timeout: 10 * time.Second}

	// Instantiating our Canadian Provider wrapper
	var weatherService providers.WeatherProvider = &providers.ECCCProvider{Client: httpClient}

	r.Get("/api/weather/bow-valley", func(w http.ResponseWriter, r *http.Request) {
		// Moose Mountain area bounding box
		bbox := "-115.10,50.80,-114.60,51.10"

		// Fire off the decoupled service method
		data, err := weatherService.FetchWeather(bbox)
		if err != nil {
			w.WriteHeader(http.StatusBadGateway)
			json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(data)
	})

	fmt.Println("Local server starting on http://localhost:8080")
	http.ListenAndServe(":8080", r)
}
