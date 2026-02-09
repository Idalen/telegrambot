package weather

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"
)

type HourlyForecast struct {
	Time          []string
	Temperature2m []float64
	Humidity2m    []float64
	Precipitation []float64
}

type HourlyUnits struct {
	Temperature2m string
	Humidity2m    string
	Precipitation string
}

type Forecast struct {
	Hourly      HourlyForecast
	HourlyUnits HourlyUnits
}

type openMeteoHourlyResponse struct {
	Hourly struct {
		Time          []string  `json:"time"`
		Temperature2m []float64 `json:"temperature_2m"`
		Humidity2m    []float64 `json:"relative_humidity_2m"`
		Precipitation []float64 `json:"precipitation"`
	} `json:"hourly"`
	HourlyUnits struct {
		Temperature2m string `json:"temperature_2m"`
		Humidity2m    string `json:"relative_humidity_2m"`
		Precipitation string `json:"precipitation"`
	} `json:"hourly_units"`
}

func SaoPauloHourlyForecast() (Forecast, error) {
	u := url.URL{
		Scheme: "https",
		Host:   "api.open-meteo.com",
		Path:   "/v1/forecast",
	}
	q := u.Query()
	q.Set("latitude", "-23.5505")
	q.Set("longitude", "-46.6333")
	q.Set("timezone", "America/Sao_Paulo")
	q.Set("hourly", "temperature_2m,relative_humidity_2m,precipitation")
	q.Set("forecast_days", "2")
	u.RawQuery = q.Encode()

	client := &http.Client{Timeout: 10 * time.Second}
	req, err := http.NewRequest(http.MethodGet, u.String(), nil)
	if err != nil {
		return Forecast{}, fmt.Errorf("Sorry, I couldn't build the weather request.")
	}

	resp, err := client.Do(req)
	if err != nil {
		return Forecast{}, fmt.Errorf("Sorry, I couldn't fetch the weather right now.")
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return Forecast{}, fmt.Errorf("Sorry, I couldn't fetch the weather right now.")
	}

	var data openMeteoHourlyResponse
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return Forecast{}, fmt.Errorf("Sorry, I couldn't read the weather response.")
	}

	return Forecast{
		Hourly: HourlyForecast{
			Time:          data.Hourly.Time,
			Temperature2m: data.Hourly.Temperature2m,
			Humidity2m:    data.Hourly.Humidity2m,
			Precipitation: data.Hourly.Precipitation,
		},
		HourlyUnits: HourlyUnits{
			Temperature2m: data.HourlyUnits.Temperature2m,
			Humidity2m:    data.HourlyUnits.Humidity2m,
			Precipitation: data.HourlyUnits.Precipitation,
		},
	}, nil
}
