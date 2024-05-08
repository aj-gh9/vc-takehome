package main

import (
	"bytes"         // Import for creating byte slices for body of HTTP requests
	"context"       // Import for managing deadlines, cancelation signals, and other request-scoped values across API boundaries
	"encoding/json" // Import for encoding and decoding JSON data
	"fmt"           // Import for formatted I/O operations (like printing output)
	"io"            // Import for basic input/output operations across multiple formats
	"math/big"      // Import for handling large integers (big numbers)
	"net/http"      // Import for making HTTP client requests
	"os"            // Import for interacting with the operating system (like reading files)
	"os/signal"     // Import for handling operating system signals
	"syscall"       // Import for interface to the low-level operating system primitives
	"time"          // Import for time-related operations (like time intervals)

	"log" // Import for logging messages in a standardized or customized format

	"gopkg.in/yaml.v2" // Import for YAML file parsing
)

// requestData defines the structure for a JSON-RPC request.
type requestData struct {
	Jsonrpc string        `json:"jsonrpc"` // JSON-RPC protocol version
	Method  string        `json:"method"`  // The method to be invoked on the JSON-RPC server
	Params  []interface{} `json:"params"`  // List of parameters for the method
	ID      int           `json:"id"`      // Identifier for the request, to match response to request
}

// response defines the structure for a JSON-RPC response.
type response struct {
	Jsonrpc string `json:"jsonrpc"` // JSON-RPC protocol version
	ID      int    `json:"id"`      // The ID from the request matched here in the response
	Result  string `json:"result"`  // The result returned by the JSON-RPC method call
}

// Config holds the configuration parameters from the YAML file.
type Config struct {
	URL             string `yaml:"url"`             // URL of the JSON-RPC server
	IntervalSeconds int    `yaml:"intervalSeconds"` // Time interval between each request in seconds
}

// HTTP client declared globally to be reused across requests, allows TCP connection reuse
var client *http.Client

// main is the entry point of the program.
func main() {
	// Initialize the HTTP client
	client = &http.Client{}

	// Setup context for handling graceful shutdown
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop() // Make sure to cancel context to prevent context leak

	// Load and validate configuration from a YAML file
	config, err := readConfig("config/config.yaml")
	if err != nil {
		log.Fatalf("Error reading config: %s", err) // Fatal log since application cannot proceed without config
	}
	if err := validateConfig(config); err != nil {
		log.Fatalf("Config validation error: %s", err) // Configuration must be valid to proceed
	}

	// Create a ticker that triggers events based on a time interval
	interval := time.Duration(config.IntervalSeconds) * time.Second
	ticker := time.NewTicker(interval)
	defer ticker.Stop() // Ensure ticker is stopped to avoid resource leaks

	// Event loop to handle ticker and shutdown signals
	for {
		select {
		case <-ticker.C: // When ticker event fires
			if err := sendPostRequest(client, config.URL); err != nil {
				log.Printf("Error sending request: %s", err) // Log any error during request sending
			}
		case <-ctx.Done(): // When shutdown signal is received
			log.Println("Shutdown signal received, exiting...")
			return // Graceful exit from the program
		}
	}
}

// sendPostRequest sends a POST request to the JSON-RPC server and handles the response.
func sendPostRequest(client *http.Client, url string) error {
	// Prepare the JSON-RPC request data
	data := requestData{
		Jsonrpc: "2.0",
		Method:  "net_peerCount",
		Params:  make([]interface{}, 0), // This method does not require any parameters
		ID:      67,                     // Arbitrary ID for the request
	}

	// Encode the requestData into JSON format for the request body
	payloadBytes, err := json.Marshal(data)
	if err != nil {
		return err // Return encoding error
	}

	// Create an HTTP POST request with JSON body
	req, err := http.NewRequest("POST", url, bytes.NewReader(payloadBytes))
	if err != nil {
		return err // Return error if request creation fails
	}
	req.Header.Set("Content-Type", "application/json") // Set content type of request

	// Send the request using the global HTTP client
	resp, err := client.Do(req)
	if err != nil {
		return err // Return error if sending the request fails
	}
	defer resp.Body.Close() // Ensure the response body is closed after reading

	// Check the status code of the response
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("server returned non-OK status: %s", resp.Status) // Return error if status is not OK
	}

	// Read the response body
	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return err // Return error if reading the response body fails
	}

	// Decode the JSON response into the 'response' struct
	var res response
	if err := json.Unmarshal(responseBody, &res); err != nil {
		return err // Return error if JSON unmarshaling fails
	}

	// Convert the hexadecimal result (as a string) to a decimal big integer
	resultInt, ok := new(big.Int).SetString(res.Result[2:], 16) // Assume result is in hex format starting with '0x'
	if !ok {
		return fmt.Errorf("failed to convert hex to decimal") // Return error if conversion fails
	}

	log.Printf("Number of Geth peers: %d", resultInt) // Log the number of peers as an integer
	return nil
}

// readConfig reads and parses the YAML configuration file.
func readConfig(filePath string) (*Config, error) {
	file, err := os.ReadFile(filePath) // Read the file content
	if err != nil {
		return nil, err // Return error if file reading fails
	}
	var config Config
	if err := yaml.Unmarshal(file, &config); err != nil {
		return nil, err // Return error if YAML unmarshaling fails
	}
	return &config, nil
}

// validateConfig checks the loaded configuration for necessary fields and valid values.
func validateConfig(config *Config) error {
	if config.URL == "" {
		return fmt.Errorf("url must be provided in the config") // Return error if URL is empty
	}
	if config.IntervalSeconds <= 0 {
		return fmt.Errorf("intervalSeconds must be greater than zero") // Return error if interval is not positive
	}
	return nil
}
