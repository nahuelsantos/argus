package utils

import (
	"encoding/json"
	"log"
	"net/http"
)

// EncodeJSON safely encodes data to JSON and writes to response writer
// Returns true if successful, false if there was an error
func EncodeJSON(w http.ResponseWriter, data interface{}) bool {
	if err := json.NewEncoder(w).Encode(data); err != nil {
		log.Printf("Error encoding JSON response: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return false
	}
	return true
}
