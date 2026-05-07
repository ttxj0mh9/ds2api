package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

// Version is set at build time via ldflags
var Version = "dev"

func main() {
	// Load environment variables from .env file if present
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using environment variables")
	}

	port := getEnv("PORT", "3000")
	host := getEnv("HOST", "0.0.0.0")

	// Validate port is a valid number
	if _, err := strconv.Atoi(port); err != nil {
		log.Fatalf("Invalid PORT value: %s", port)
	}

	addr := fmt.Sprintf("%s:%s", host, port)

	router := setupRouter()

	log.Printf("ds2api %s starting on %s", Version, addr)

	server := &http.Server{
		Addr:    addr,
		Handler: router,
	}

	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("Server failed to start: %v", err)
	}
}

// setupRouter initialises and returns the HTTP router with all routes registered.
func setupRouter() http.Handler {
	mux := http.NewServeMux()

	// Health / readiness probe
	mux.HandleFunc("/health", healthHandler)

	// API version info
	mux.HandleFunc("/version", versionHandler)

	return mux
}

// healthHandler responds with a simple 200 OK to indicate the service is alive.
func healthHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, `{"status":"ok"}`)
}

// versionHandler returns the current build version of the service.
func versionHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, `{"version":%q}`, Version)
}

// getEnv returns the value of the environment variable named by key,
// or fallback if the variable is not set or empty.
func getEnv(key, fallback string) string {
	if val, ok := os.LookupEnv(key); ok && val != "" {
		return val
	}
	return fallback
}
