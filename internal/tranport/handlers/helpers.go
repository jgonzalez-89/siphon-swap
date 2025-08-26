package handlers

import (
	"encoding/json"
	"net/http"
)

// respondWithJSON escribe una respuesta JSON
func respondWithJSON(w http.ResponseWriter, data interface{}, status int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

// respondWithError escribe un error en formato JSON
func respondWithError(w http.ResponseWriter, message string, status int) {
	respondWithJSON(w, map[string]string{"error": message}, status)
}
