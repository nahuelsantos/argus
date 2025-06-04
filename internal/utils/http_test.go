package utils

import (
	"bytes"
	"encoding/json"
	"log"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEncodeJSON(t *testing.T) {
	tests := []struct {
		name           string
		data           interface{}
		expectedStatus int
		expectedBody   string
		shouldSucceed  bool
	}{
		{
			name: "encode simple struct",
			data: struct {
				Name string `json:"name"`
				Age  int    `json:"age"`
			}{
				Name: "John",
				Age:  30,
			},
			expectedStatus: http.StatusOK,
			expectedBody:   `{"name":"John","age":30}`,
			shouldSucceed:  true,
		},
		{
			name:           "encode nil",
			data:           nil,
			expectedStatus: http.StatusOK,
			expectedBody:   "null",
			shouldSucceed:  true,
		},
		{
			name:           "encode empty string",
			data:           "",
			expectedStatus: http.StatusOK,
			expectedBody:   `""`,
			shouldSucceed:  true,
		},
		{
			name:           "encode slice",
			data:           []string{"apple", "banana", "cherry"},
			expectedStatus: http.StatusOK,
			expectedBody:   `["apple","banana","cherry"]`,
			shouldSucceed:  true,
		},
		{
			name: "encode map",
			data: map[string]interface{}{
				"status": "success",
				"count":  42,
				"items":  []string{"a", "b"},
			},
			expectedStatus: http.StatusOK,
			expectedBody:   `{"count":42,"items":["a","b"],"status":"success"}`,
			shouldSucceed:  true,
		},
		{
			name:           "encode boolean",
			data:           true,
			expectedStatus: http.StatusOK,
			expectedBody:   "true",
			shouldSucceed:  true,
		},
		{
			name:           "encode number",
			data:           123.45,
			expectedStatus: http.StatusOK,
			expectedBody:   "123.45",
			shouldSucceed:  true,
		},
		{
			name: "encode complex nested structure",
			data: struct {
				User struct {
					ID       int      `json:"id"`
					Name     string   `json:"name"`
					Tags     []string `json:"tags"`
					Settings struct {
						Theme string `json:"theme"`
						Count int    `json:"count"`
					} `json:"settings"`
				} `json:"user"`
			}{
				User: struct {
					ID       int      `json:"id"`
					Name     string   `json:"name"`
					Tags     []string `json:"tags"`
					Settings struct {
						Theme string `json:"theme"`
						Count int    `json:"count"`
					} `json:"settings"`
				}{
					ID:   1,
					Name: "Alice",
					Tags: []string{"admin", "user"},
					Settings: struct {
						Theme string `json:"theme"`
						Count int    `json:"count"`
					}{
						Theme: "dark",
						Count: 5,
					},
				},
			},
			expectedStatus: http.StatusOK,
			expectedBody:   `{"user":{"id":1,"name":"Alice","tags":["admin","user"],"settings":{"theme":"dark","count":5}}}`,
			shouldSucceed:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a response recorder
			w := httptest.NewRecorder()

			// Call the function
			result := EncodeJSON(w, tt.data)

			// Check the return value
			assert.Equal(t, tt.shouldSucceed, result)

			// Check status code
			assert.Equal(t, tt.expectedStatus, w.Code)

			// Check response body
			body := strings.TrimSpace(w.Body.String())
			if tt.expectedBody == `{"count":42,"items":["a","b"],"status":"success"}` {
				// For maps, order is not guaranteed, so we need to unmarshal and compare
				var expected, actual map[string]interface{}
				err := json.Unmarshal([]byte(tt.expectedBody), &expected)
				require.NoError(t, err)
				err = json.Unmarshal([]byte(body), &actual)
				require.NoError(t, err)
				assert.Equal(t, expected, actual)
			} else {
				assert.JSONEq(t, tt.expectedBody, body)
			}

			// Check Content-Type header - httptest.ResponseRecorder sets text/plain for JSON
			if tt.shouldSucceed {
				contentType := w.Header().Get("Content-Type")
				assert.Equal(t, "text/plain; charset=utf-8", contentType, "ResponseRecorder sets Content-Type for written content")
			}
		})
	}
}

func TestEncodeJSON_ErrorCases(t *testing.T) {
	// Test with a broken ResponseWriter that fails to write
	t.Run("broken response writer", func(t *testing.T) {
		// Capture log output
		var buf bytes.Buffer
		oldOutput := log.Writer()
		log.SetOutput(&buf)
		defer log.SetOutput(oldOutput)

		// Create a broken response writer (this is tricky since httptest.ResponseRecorder rarely fails)
		// We'll use a custom writer that always fails
		brokenWriter := &brokenResponseWriter{}

		result := EncodeJSON(brokenWriter, "test")

		// Should return false on error
		assert.False(t, result)

		// Should log an error
		logOutput := buf.String()
		assert.Contains(t, logOutput, "Error encoding JSON response")
	})

	// Test with data that can't be marshaled to JSON
	t.Run("unmarshalable data", func(t *testing.T) {
		// Capture log output
		var buf bytes.Buffer
		oldOutput := log.Writer()
		log.SetOutput(&buf)
		defer log.SetOutput(oldOutput)

		w := httptest.NewRecorder()

		// Function values cannot be marshaled to JSON
		unmarshalableData := func() {}

		result := EncodeJSON(w, unmarshalableData)

		// Should return false on error
		assert.False(t, result)

		// Should set error status
		assert.Equal(t, http.StatusInternalServerError, w.Code)

		// Should set error message
		body := strings.TrimSpace(w.Body.String())
		assert.Equal(t, "Internal server error", body)

		// Should log an error
		logOutput := buf.String()
		assert.Contains(t, logOutput, "Error encoding JSON response")
	})

	// Test with channel (another unmarshalable type)
	t.Run("channel data", func(t *testing.T) {
		var buf bytes.Buffer
		oldOutput := log.Writer()
		log.SetOutput(&buf)
		defer log.SetOutput(oldOutput)

		w := httptest.NewRecorder()
		ch := make(chan int)

		result := EncodeJSON(w, ch)

		assert.False(t, result)
		assert.Equal(t, http.StatusInternalServerError, w.Code)
		assert.Contains(t, buf.String(), "Error encoding JSON response")
	})
}

// Custom ResponseWriter that always fails to write
type brokenResponseWriter struct {
	header http.Header
}

func (w *brokenResponseWriter) Header() http.Header {
	if w.header == nil {
		w.header = make(http.Header)
	}
	return w.header
}

func (w *brokenResponseWriter) Write([]byte) (int, error) {
	return 0, assert.AnError // Always fail
}

func (w *brokenResponseWriter) WriteHeader(statusCode int) {
	// Do nothing
}

func TestEncodeJSON_ContentType(t *testing.T) {
	t.Run("sets expected content-type", func(t *testing.T) {
		w := httptest.NewRecorder()
		data := map[string]string{"test": "value"}

		result := EncodeJSON(w, data)

		assert.True(t, result)
		assert.Equal(t, "text/plain; charset=utf-8", w.Header().Get("Content-Type"))
	})

	t.Run("preserves existing content-type", func(t *testing.T) {
		w := httptest.NewRecorder()
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		data := map[string]string{"test": "value"}

		result := EncodeJSON(w, data)

		assert.True(t, result)
		assert.Equal(t, "application/json; charset=utf-8", w.Header().Get("Content-Type"))
	})
}

func TestEncodeJSON_ConcurrentUsage(t *testing.T) {
	// Test that the function is safe for concurrent use
	t.Run("concurrent calls", func(t *testing.T) {
		const numGoroutines = 100
		done := make(chan bool, numGoroutines)

		for i := 0; i < numGoroutines; i++ {
			go func(id int) {
				defer func() { done <- true }()

				w := httptest.NewRecorder()
				testData := map[string]interface{}{
					"message": "hello",
					"id":      id,
				}

				result := EncodeJSON(w, testData)
				assert.True(t, result)
				assert.Equal(t, http.StatusOK, w.Code)

				// Verify the response is valid JSON
				var response map[string]interface{}
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Equal(t, "hello", response["message"])
				assert.Equal(t, float64(id), response["id"]) // JSON numbers are float64
			}(i)
		}

		// Wait for all goroutines to complete
		for i := 0; i < numGoroutines; i++ {
			<-done
		}
	})
}

// Benchmark tests
func BenchmarkEncodeJSON_Simple(b *testing.B) {
	data := struct {
		Name string `json:"name"`
		Age  int    `json:"age"`
	}{
		Name: "John",
		Age:  30,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		w := httptest.NewRecorder()
		EncodeJSON(w, data)
	}
}

func BenchmarkEncodeJSON_Complex(b *testing.B) {
	data := map[string]interface{}{
		"users": []map[string]interface{}{
			{"id": 1, "name": "Alice", "tags": []string{"admin", "user"}},
			{"id": 2, "name": "Bob", "tags": []string{"user"}},
			{"id": 3, "name": "Charlie", "tags": []string{"guest"}},
		},
		"metadata": map[string]interface{}{
			"total":  3,
			"offset": 0,
			"limit":  10,
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		w := httptest.NewRecorder()
		EncodeJSON(w, data)
	}
}

// Example of how to use EncodeJSON
func ExampleEncodeJSON() {
	// Create test data
	data := struct {
		Message string `json:"message"`
		Status  int    `json:"status"`
	}{
		Message: "Hello, World!",
		Status:  200,
	}

	// Create a response recorder for testing
	w := httptest.NewRecorder()

	// Encode JSON
	success := EncodeJSON(w, data)
	if success {
		log.Printf("Response: %s", w.Body.String())
	}

	// Output will be logged
}
