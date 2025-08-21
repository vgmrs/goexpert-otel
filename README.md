# goexpert-otel
Exercise on Observability and OpenTelemetry for Go Expert Postgraduate Course

## Description

An application that receives a ZIP code (CEP), identifies the city, and returns the current weather (temperature in Celsius, Fahrenheit, and Kelvin) along with the city name. This system implements OpenTelemetry (OTEL) and Zipkin for observability.

## How to Install

1. Make sure you have Go installed (version 1.20 or higher)

2. Install dependencies:
   ```bash
   make install
   ```

## How to Run

1. Start all services:
   ```bash
   make up
   ```

2. You can test the API by sending a POST request to `http://localhost:8080/cep` with a JSON body like this:
   ```json
   {
     "cep": "01001000"
   }
   ```
   Replace `01001000` with a valid Brazilian ZIP code (only numbers, e.g., 01001000 for São Paulo)

## Architecture

- **Service A**: Receives requests, forwards them to Service B, and returns the response
- **Service B**: Handles the business logic for ZIP code validation and weather information retrieval
- **OpenTelemetry**: Instrumentation for generating traces and metrics
- **Zipkin**: Distributed tracing system for visualizing request flows
