package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/zipkin"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.21.0"
)

type CEPResponse struct {
	CEP        string `json:"cep"`
	Localidade string `json:"localidade"`
	UF         string `json:"uf"`
	Error      bool   `json:"erro"`
}

type WeatherResponse struct {
	Current struct {
		TempC float64 `json:"temp_c"`
	} `json:"current"`
}

type WeatherResult struct {
	City  string  `json:"city"`
	TempC float64 `json:"temp_C"`
	TempF float64 `json:"temp_F"`
	TempK float64 `json:"temp_K"`
}

func initTracer() (*sdktrace.TracerProvider, error) {
	zipkinURL := os.Getenv("OTEL_EXPORTER_ZIPKIN_ENDPOINT")
	exporter, err := zipkin.New(zipkinURL)
	if err != nil {
		return nil, err
	}

	tracerProvider := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exporter),
		sdktrace.WithResource(resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceName("service-b"),
		)),
	)
	otel.SetTracerProvider(tracerProvider)
	return tracerProvider, nil
}

func main() {
	tp, err := initTracer()
	if err != nil {
		log.Fatal(err)
	}
	defer tp.Shutdown(context.Background())

	mux := http.NewServeMux()
	mux.HandleFunc("/weather/", handleWeather)

	port := ":8081"
	log.Printf("Service B starting on port %s", port)
	log.Fatal(http.ListenAndServe(port, mux))
}

func handleWeather(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	ctx := r.Context()
	tracer := otel.Tracer("service-b")
	ctx, span := tracer.Start(ctx, "handle-weather-request")
	defer span.End()

	parts := strings.Split(r.URL.Path, "/")
	if len(parts) != 3 {
		http.Error(w, "invalid request path", http.StatusBadRequest)
		return
	}

	cep := parts[2]

	_, locationSpan := tracer.Start(ctx, "get-location")
	location, err := getLocation(cep)
	locationSpan.End()
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	_, weatherSpan := tracer.Start(ctx, "get-weather")
	weather, err := getWeather(location.Localidade)
	weatherSpan.End()
	if err != nil {
		http.Error(w, "error getting weather data", http.StatusInternalServerError)
		return
	}

	tempC := weather.Current.TempC
	tempF := tempC*1.8 + 32
	tempK := tempC + 273.15

	result := WeatherResult{
		City:  location.Localidade,
		TempC: tempC,
		TempF: tempF,
		TempK: tempK,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}

func getLocation(cep string) (*CEPResponse, error) {
	resp, err := http.Get(fmt.Sprintf("https://viacep.com.br/ws/%s/json/", cep))
	if err != nil {
		return nil, fmt.Errorf("can not find zipcode")
	}
	defer resp.Body.Close()

	var location CEPResponse
	if err := json.NewDecoder(resp.Body).Decode(&location); err != nil {
		return nil, fmt.Errorf("can not find zipcode")
	}

	if location.Error {
		return nil, fmt.Errorf("can not find zipcode")
	}

	return &location, nil
}

func getWeather(city string) (*WeatherResponse, error) {
	apiKey := os.Getenv("WEATHER_API_KEY")
	escapedCity := url.QueryEscape(city)

	url := fmt.Sprintf("https://api.weatherapi.com/v1/current.json?key=%s&q=%s", apiKey, escapedCity)

	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var weather WeatherResponse
	if err := json.Unmarshal(body, &weather); err != nil {
		return nil, err
	}

	return &weather, nil
}
