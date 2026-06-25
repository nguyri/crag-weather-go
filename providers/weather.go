package providers

import (
	"encoding/json"
	"fmt"
	"net/http"
)

type WeatherCondition struct {
	StationName string  `json:"station_name"`
	Temperature float64 `json:"temperature_celsius"`
	Humidity    float64 `json:"humidity_percent"`
	WindSpeed   float64 `json:"wind_speed_kmh"`
	WindDir     float64 `json:"wind_direction"`
	Latitude    float64 `json:"latitude"`
	Longitude   float64 `json:"longitude"`
}

type WeatherProvider interface {
	FetchWeather(bbox string) ([]WeatherCondition, error)
}

type ECCCProvider struct {
	Client *http.Client
}

type ecccRawResponse struct {
	Features []struct {
		Geometry struct {
			Coordinates []float64 `json:"coordinates"`
		} `json:"geometry"`
		Properties struct {
			StationName string  `json:"stn_nam-value"`
			AirTemp     float64 `json:"air_temp"`
			RelHum      float64 `json:"rel_hum"`
			WindSpd     float64 `json:"avg_wnd_spd_10m_pst1hr"`
			WindDir     float64 `json:"avg_wnd_dir_10m_pst10mts"`
		} `json:"properties"`
	} `json:"features"`
}

func (e *ECCCProvider) FetchWeather(bbox string) ([]WeatherCondition, error) {
	u := fmt.Sprintf("https://api.weather.gc.ca/collections/swob-realtime/items?bbox=%s&limit=5&f=json", bbox)

	req, err := http.NewRequest("GET", u, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "(bowvalleyclimbapp.ca, richard@example.com)")

	resp, err := e.Client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("provider returned status code: %d", resp.StatusCode)
	}

	var raw ecccRawResponse
	if err := json.NewDecoder(resp.Body).Decode(&raw); err != nil {
		return nil, err
	}

	// Use a map to ensure we only keep one entry per unique station name
	latestStations := make(map[string]WeatherCondition)

	for _, feature := range raw.Features {
		props := feature.Properties
		name := props.StationName

		// Skip if we already have this station (prevents duplicate time-series entries)
		if _, exists := latestStations[name]; exists {
			continue
		}

		coords := feature.Geometry.Coordinates
		lon, lat := 0.0, 0.0
		if len(coords) >= 2 {
			lon = coords[0]
			lat = coords[1]
		}

		// Add to map
		latestStations[name] = WeatherCondition{
			StationName: name,
			Temperature: props.AirTemp,
			Humidity:    props.RelHum,
			WindSpeed:   props.WindSpd,
			WindDir:     props.WindDir,
			Latitude:    lat,
			Longitude:   lon,
		}
	}

	// Convert the map back into a slice to return
	var finalConditions []WeatherCondition
	for _, cond := range latestStations {
		finalConditions = append(finalConditions, cond)
	}

	return finalConditions, nil
}
