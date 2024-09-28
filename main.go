package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"gopkg.in/ini.v1"
)

// Configuration struct to hold server configuration
type Configuration struct {
	ServerName string
	Port       string
	LogFolder  string
	LogLevel   string
	HTMLDir    string
}

// Global variable to hold the configuration
var config Configuration

// Function to read configuration from configuration.ini file
func loadConfig(filePath string) error {
	cfg, err := ini.Load(filePath)
	if err != nil {
		return fmt.Errorf("failed to read config file: %v", err)
	}

	config.ServerName = cfg.Section("server").Key("name").String()
	config.Port = cfg.Section("server").Key("port").String()
	config.LogFolder = cfg.Section("server").Key("log_folder").String()
	config.LogLevel = cfg.Section("server").Key("log_level").String()
	config.HTMLDir = cfg.Section("paths").Key("html_dir").String()

	return nil
}

// Logging middleware to log HTTP requests
func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		fmt.Printf("Started %s %s\n", r.Method, r.URL.Path)
		log.Printf("Started %s %s", r.Method, r.URL.Path)

		next.ServeHTTP(w, r)

		log.Printf("Completed in %v", time.Since(start))
	})
}

// Setup logging to log file
func setupLogging(logFolder string) error {
	logPath := filepath.Join(logFolder, "server.log")
	logFile, err := os.OpenFile(logPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("could not open log file: %v", err)
	}

	// Direct log output to log file
	log.SetOutput(logFile)
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	return nil
}

func main() {
	// Load configuration
	if err := loadConfig("configuration.ini"); err != nil {
		fmt.Printf("Error loading config: %v\n", err)
		return
	}

	// Setup logging
	if err := setupLogging(config.LogFolder); err != nil {
		fmt.Printf("Error setting up logging: %v\n", err)
		return
	}

	// Log server startup details
	log.Printf("Starting %s on port %s", config.ServerName, config.Port)

	// Serve static files from the configured HTML directory
	fileServer := http.FileServer(http.Dir(config.HTMLDir))

	// Wrap the file server with logging middleware
	http.Handle("/", loggingMiddleware(fileServer))

	// Start HTTP server
	if err := http.ListenAndServe(":"+config.Port, nil); err != nil {
		log.Fatalf("Error starting server: %v", err)
	}
}
