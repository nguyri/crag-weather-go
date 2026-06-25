package providers

import (
	"encoding/json"
	"fmt"
	"net/http"
)

// WeatherCondition is our universal internal layout.
// Every provider must convert its unique data into this unified struct.
type WeatherCondition struct {
	StationName string  `json:"station_name"`
	Temperature float64 `json:"temperature_celsius"`
	Humidity    float64 `json:"humidity_percent"`
	WindSpeed   float64 `json:"wind_speed_kmh"`
}

// WeatherProvider defines the contract that any weather service must follow.
type WeatherProvider interface {
	FetchWeather(bbox string) (*WeatherCondition, error)
}

// ============================================================================
// ENVIRONMENT CANADA (ECCC) IMPLEMENTATION
// ============================================================================

type ECCCProvider struct {
	Client *http.Client
}

// The raw external JSON schema unique to MSC GeoMet
type ecccRawResponse struct {
	Features []struct {
		Properties struct {
			StationName string  `json:"stn_nam-value"`
			AirTemp     float64 `json:"air_temp"`
			RelHum      float64 `json:"rel_hum"`
			WindSpd     float64 `json:"avg_wnd_spd_10m_pst1hr"`
			WindDir     float64 `json:"avg_wnd_dir_10m_pst10mts`
		} `json:"properties"`
	} `json:"features"`
}

func (e *ECCCProvider) FetchWeather(bbox string) (*WeatherCondition, error) {
	u := fmt.Sprintf("https://api.weather.gc.ca/collections/swob-realtime/items?bbox=%s&limit=1&f=json", bbox)

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

	if len(raw.Features) == 0 {
		return nil, fmt.Errorf("no weather station features found in bounding box")
	}

	// Map raw data seamlessly to our standard internal format
	props := raw.Features[0].Properties
	return &WeatherCondition{
		StationName: props.StationName,
		Temperature: props.AirTemp,
		Humidity:    props.RelHum,
		WindSpeed:   props.WindSpd,
	}, nil
}
