package gateway

import (
	"encoding/json"
	"net/http"
)

func WriteError(w http.ResponseWriter, status int, code, message, id string) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("X-Request-ID", id)
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(map[string]any{"error": map[string]string{"code": code, "message": message, "request_id": id}})
}
