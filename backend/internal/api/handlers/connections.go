package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"github.com/hejijunhao/heimdall/backend/internal/api/middleware"
	"github.com/hejijunhao/heimdall/backend/internal/db"
)

func (s *Server) ListConnections(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		jsonError(w, "missing user context", http.StatusUnauthorized)
		return
	}

	connections, err := s.Queries.ListConnectionsByUser(r.Context(), userID)
	if err != nil {
		jsonError(w, "failed to list connections", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(connections)
}

func (s *Server) GetConnection(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		jsonError(w, "missing user context", http.StatusUnauthorized)
		return
	}

	connID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		jsonError(w, "invalid connection id", http.StatusBadRequest)
		return
	}

	conn, err := s.Queries.GetConnectionByUser(r.Context(), db.GetConnectionByUserParams{
		ID:     connID,
		UserID: userID,
	})
	if err != nil {
		jsonError(w, "connection not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(conn)
}

type createConnectionRequest struct {
	Name      string          `json:"name"`
	Type      string          `json:"type"`
	Direction string          `json:"direction"`
	Config    json.RawMessage `json:"config"`
}

func (s *Server) CreateConnection(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		jsonError(w, "missing user context", http.StatusUnauthorized)
		return
	}

	var req createConnectionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonError(w, "invalid request body", http.StatusBadRequest)
		return
	}

	if req.Name == "" || req.Type == "" {
		jsonError(w, "name and type are required", http.StatusBadRequest)
		return
	}

	direction := req.Direction
	if direction == "" {
		direction = "one_way"
	}
	config := req.Config
	if config == nil {
		config = json.RawMessage(`{}`)
	}

	conn, err := s.Queries.CreateConnection(r.Context(), db.CreateConnectionParams{
		UserID:    userID,
		Name:      req.Name,
		Type:      req.Type,
		Direction: direction,
		Config:    config,
		Status:    "inactive",
	})
	if err != nil {
		jsonError(w, "failed to create connection", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(conn)
}

type updateConnectionRequest struct {
	Name      string          `json:"name"`
	Type      string          `json:"type"`
	Direction string          `json:"direction"`
	Config    json.RawMessage `json:"config"`
	Status    string          `json:"status"`
}

func (s *Server) UpdateConnection(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		jsonError(w, "missing user context", http.StatusUnauthorized)
		return
	}

	connID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		jsonError(w, "invalid connection id", http.StatusBadRequest)
		return
	}

	var req updateConnectionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonError(w, "invalid request body", http.StatusBadRequest)
		return
	}

	if req.Name == "" || req.Type == "" {
		jsonError(w, "name and type are required", http.StatusBadRequest)
		return
	}

	direction := req.Direction
	if direction == "" {
		direction = "one_way"
	}
	config := req.Config
	if config == nil {
		config = json.RawMessage(`{}`)
	}
	status := req.Status
	if status == "" {
		status = "inactive"
	}

	conn, err := s.Queries.UpdateConnection(r.Context(), db.UpdateConnectionParams{
		ID:        connID,
		Name:      req.Name,
		Type:      req.Type,
		Direction: direction,
		Config:    config,
		Status:    status,
		UserID:    userID,
	})
	if err != nil {
		jsonError(w, "connection not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(conn)
}

func (s *Server) DeleteConnection(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		jsonError(w, "missing user context", http.StatusUnauthorized)
		return
	}

	connID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		jsonError(w, "invalid connection id", http.StatusBadRequest)
		return
	}

	err = s.Queries.DeleteConnectionByUser(r.Context(), db.DeleteConnectionByUserParams{
		ID:     connID,
		UserID: userID,
	})
	if err != nil {
		jsonError(w, "failed to delete connection", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
