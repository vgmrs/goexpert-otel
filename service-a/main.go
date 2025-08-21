package main

import (
	"context"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"os"
	"regexp"
	"strings"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/zipkin"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.21.0"
)

type ZipCodeRequest struct {
	CEP string `json:"cep"`
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
			semconv.ServiceName("service-a"),
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
	mux.HandleFunc("/cep", handleCEP)

	port := ":8080"
	log.Printf("Service A starting on port %s", port)
	log.Fatal(http.ListenAndServe(port, mux))
}

func handleCEP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	ctx := r.Context()
	tracer := otel.Tracer("service-a")
	ctx, span := tracer.Start(ctx, "handle-cep-request")
	defer span.End()

	var req ZipCodeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	if !isValidCEP(strings.TrimSpace(req.CEP)) {
		http.Error(w, "invalid zipcode", http.StatusUnprocessableEntity)
		return
	}

	BServiceURL := os.Getenv("SERVICE_B_URL")
	client := &http.Client{Timeout: 10 * time.Second}

	cep := strings.ReplaceAll(strings.TrimSpace(req.CEP), "-", "")

	serviceBReq, err := http.NewRequestWithContext(ctx, "GET", BServiceURL+"/weather/"+cep, nil)
	if err != nil {
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	resp, err := client.Do(serviceBReq)
	if err != nil {
		http.Error(w, "error calling service b", http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			http.Error(w, "error reading response body", http.StatusInternalServerError)
			return
		}

		var errorResponse map[string]string
		if err := json.Unmarshal(body, &errorResponse); err == nil && errorResponse["message"] != "" {
			http.Error(w, errorResponse["message"], resp.StatusCode)
		} else {
			http.Error(w, string(body), resp.StatusCode)
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		http.Error(w, "error parsing response", http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(result)
}

func isValidCEP(cep string) bool {
	return regexp.MustCompile(`^\d{5}-?\d{3}$`).MatchString(cep)
}
