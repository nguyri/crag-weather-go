package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"crag-weather-go/providers"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

// 1. Define a central struct for our area of interest
type BBox struct {
	MinLon float64 `json:"min_lon"`
	MinLat float64 `json:"min_lat"`
	MaxLon float64 `json:"max_lon"`
	MaxLat float64 `json:"max_lat"`
}

func main() {
	// 2. Single source of truth for the coordinates
	activeArea := BBox{
		MinLon: -115.30,
		MinLat: 50.70,
		MaxLon: -114.40,
		MaxLat: 51.10,
	}

	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	r.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Access-Control-Allow-Origin", "*")
			w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
			if r.Method == "OPTIONS" {
				w.WriteHeader(http.StatusOK)
				return
			}
			next.ServeHTTP(w, r)
		})
	})

	// 3. Use the shared variable for the config API
	r.Get("/api/config", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{"bbox": activeArea})
	})

	// 4. Use the shared variable to build the ECCC string dynamically
	httpClient := &http.Client{Timeout: 10 * time.Second}
	var weatherService providers.WeatherProvider = &providers.ECCCProvider{Client: httpClient}

	r.Get("/api/weather/bow-valley", func(w http.ResponseWriter, r *http.Request) {
		// Formats the struct into the string the API expects: minLon,minLat,maxLon,maxLat
		regionalBbox := fmt.Sprintf("%f,%f,%f,%f", activeArea.MinLon, activeArea.MinLat, activeArea.MaxLon, activeArea.MaxLat)

		stations, err := weatherService.FetchWeather(regionalBbox)
		if err != nil {
			w.WriteHeader(http.StatusBadGateway)
			json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(stations)
	})

	fmt.Println("Local server starting on http://localhost:8080")
	http.ListenAndServe(":8080", r)
}
