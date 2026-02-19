package handlers

import (
	"encoding/json"
	"net/http"
)

func ListConnections(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode([]any{})
}

func GetConnection(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	http.Error(w, `{"error":"not implemented"}`, http.StatusNotImplemented)
}

func CreateConnection(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	http.Error(w, `{"error":"not implemented"}`, http.StatusNotImplemented)
}

func UpdateConnection(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	http.Error(w, `{"error":"not implemented"}`, http.StatusNotImplemented)
}

func DeleteConnection(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	http.Error(w, `{"error":"not implemented"}`, http.StatusNotImplemented)
}
