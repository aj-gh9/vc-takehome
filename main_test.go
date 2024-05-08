package main

import (
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
)

// TestSendPostRequest checks the sendPostRequest function with a mock server.
func TestSendPostRequest(t *testing.T) {
	// Define test cases for various response scenarios to simulate different behaviors of the JSON-RPC server.
	tests := []struct {
		name        string // Name of the test case for clarity
		status      int    // HTTP status to simulate
		response    string // JSON-RPC server response body to simulate
		expectError bool   // Whether an error is expected from sendPostRequest
	}{
		// A valid JSON-RPC response which should not cause an error
		{"ValidResponse", http.StatusOK, `{"jsonrpc":"2.0", "id":67, "result":"0x10"}`, false},
		// Server error scenario: Should result in an error in sendPostRequest
		{"ServerError", http.StatusInternalServerError, `{"error": "internal"}`, true},
		// Bad request status with non-JSON response: Should result in an error
		{"BadRequest", http.StatusBadRequest, `Bad request`, true},
		// Valid HTTP status but invalid JSON content (incorrect format for result)
		{"InvalidJSON", http.StatusOK, `{"jsonrpc": "2.0", "id": 67, "result": "not hex"}`, true},
	}

	// Iterate over each test case
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a new HTTP server to simulate responses for each test case
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				// Set the desired status code and response body for the simulated server
				w.WriteHeader(tt.status)
				w.Write([]byte(tt.response))
			}))
			defer server.Close() // Ensure server is closed after the test

			// Invoke sendPostRequest using the URL of our mock server
			err := sendPostRequest(&http.Client{}, server.URL)
			// Check if the result meets our expectations
			if (err != nil && !tt.expectError) || (err == nil && tt.expectError) {
				t.Errorf("sendPostRequest() error = %v, expectError %v", err, tt.expectError)
			}
		})
	}
}

// TestReadConfig tests the configuration reading and parsing.
func TestReadConfig(t *testing.T) {
	// Simulate reading a valid config file
	validConfig := "url: \"http://127.0.0.1:8545/\"\nintervalSeconds: 10"
	tmpFile, err := os.CreateTemp("", "config*.yaml")
	if err != nil {
		t.Fatal("Cannot create temporary file:", err)
	}
	defer os.Remove(tmpFile.Name()) // Clean up the file afterwards

	if _, err := tmpFile.Write([]byte(validConfig)); err != nil {
		t.Fatal("Failed to write to temporary file:", err)
	}
	if err := tmpFile.Close(); err != nil {
		t.Fatal("Failed to close the file:", err)
	}

	config, err := readConfig(tmpFile.Name())
	if err != nil {
		t.Errorf("readConfig returned an error: %v", err)
	}
	if config.URL != "http://127.0.0.1:8545/" || config.IntervalSeconds != 10 {
		t.Errorf("readConfig did not parse the contents correctly")
	}
}

// TestValidateConfig tests the validation logic for configuration settings.
func TestValidateConfig(t *testing.T) {
	validConfig := &Config{URL: "http://127.0.0.1:8545/", IntervalSeconds: 10}
	if err := validateConfig(validConfig); err != nil {
		t.Errorf("validateConfig should not have returned an error for valid input: %v", err)
	}

	invalidConfig := &Config{URL: "", IntervalSeconds: 10}
	if err := validateConfig(invalidConfig); err == nil {
		t.Errorf("validateConfig should have returned an error for invalid URL")
	}

	invalidConfig = &Config{URL: "http://127.0.0.1:8545/", IntervalSeconds: -1}
	if err := validateConfig(invalidConfig); err == nil {
		t.Errorf("validateConfig should have returned an error for invalid IntervalSeconds")
	}
}
