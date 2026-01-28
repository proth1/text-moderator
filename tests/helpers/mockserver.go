package helpers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// ModerationResponse represents a mock HuggingFace API response
type ModerationResponse struct {
	Input    string          `json:"input"`
	Response json.RawMessage `json:"response,omitempty"`
	Error    string          `json:"error,omitempty"`
	Status   int             `json:"status,omitempty"`
}

// MockServer provides a mock HTTP server for HuggingFace API responses
type MockServer struct {
	Server    *httptest.Server
	Responses map[string]ModerationResponse
}

// SetupMockServer creates a mock HuggingFace API server
func SetupMockServer(t *testing.T) (*MockServer, func()) {
	t.Helper()

	// Load fixture responses
	responses, err := loadModerationResponses(t)
	if err != nil {
		t.Fatalf("Failed to load moderation responses: %v", err)
	}

	mock := &MockServer{
		Responses: responses,
	}

	// Create HTTP test server
	mock.Server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mock.handleRequest(t, w, r)
	}))

	cleanup := func() {
		mock.Server.Close()
	}

	return mock, cleanup
}

// loadModerationResponses loads mock responses from fixtures/moderation_responses.json
func loadModerationResponses(t *testing.T) (map[string]ModerationResponse, error) {
	t.Helper()

	// Find the fixture file
	fixtureFile := filepath.Join("tests", "fixtures", "moderation_responses.json")
	if _, err := os.Stat(fixtureFile); os.IsNotExist(err) {
		// Try alternative path
		fixtureFile = filepath.Join("..", "fixtures", "moderation_responses.json")
	}

	data, err := os.ReadFile(fixtureFile)
	if err != nil {
		return nil, fmt.Errorf("failed to read fixture file: %w", err)
	}

	var responses map[string]ModerationResponse
	if err := json.Unmarshal(data, &responses); err != nil {
		return nil, fmt.Errorf("failed to parse fixture file: %w", err)
	}

	return responses, nil
}

// handleRequest handles incoming HTTP requests and returns mock responses
func (m *MockServer) handleRequest(t *testing.T, w http.ResponseWriter, r *http.Request) {
	t.Helper()

	// Parse request body
	var reqBody struct {
		Inputs string `json:"inputs"`
	}

	if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Match input text to fixture response
	var response *ModerationResponse
	for _, resp := range m.Responses {
		if strings.Contains(reqBody.Inputs, resp.Input) || resp.Input == "any text" {
			response = &resp
			break
		}
	}

	// Default to safe_text if no match
	if response == nil {
		safeResp := m.Responses["safe_text"]
		response = &safeResp
	}

	// Handle error responses
	if response.Error != "" {
		status := response.Status
		if status == 0 {
			status = http.StatusInternalServerError
		}
		http.Error(w, response.Error, status)
		return
	}

	// Return successful response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(json.RawMessage(response.Response))
}

// AddCustomResponse adds a custom response to the mock server
func (m *MockServer) AddCustomResponse(key string, response ModerationResponse) {
	m.Responses[key] = response
}

// SetTimeout configures the mock server to return timeout errors
func (m *MockServer) SetTimeout(enable bool) {
	if enable {
		m.Responses["timeout"] = ModerationResponse{
			Input:  "any text",
			Error:  "timeout",
			Status: http.StatusGatewayTimeout,
		}
	} else {
		delete(m.Responses, "timeout")
	}
}

// SetRateLimit configures the mock server to return rate limit errors
func (m *MockServer) SetRateLimit(enable bool) {
	if enable {
		m.Responses["rate_limit"] = ModerationResponse{
			Input:  "any text",
			Error:  "rate_limit_exceeded",
			Status: http.StatusTooManyRequests,
		}
	} else {
		delete(m.Responses, "rate_limit")
	}
}

// URL returns the mock server URL
func (m *MockServer) URL() string {
	return m.Server.URL
}
