package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/jackc/pgx/v5/pgtype"

	"github.com/hejijunhao/heimdall/backend/internal/api/middleware"
	"github.com/hejijunhao/heimdall/backend/internal/db"
)

func (s *Server) GetOrganization(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		jsonError(w, "missing user context", http.StatusUnauthorized)
		return
	}

	org, err := s.Queries.GetOrganizationByUser(r.Context(), userID)
	if err != nil {
		jsonError(w, "organization not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(org)
}

type onboardingRequest struct {
	OrgName string `json:"org_name"`
	OrgSlug string `json:"org_slug"`
	AppName string `json:"app_name"`
}

func (s *Server) Onboard(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		jsonError(w, "missing user context", http.StatusUnauthorized)
		return
	}

	var req onboardingRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonError(w, "invalid request body", http.StatusBadRequest)
		return
	}

	if req.OrgName == "" || req.OrgSlug == "" {
		jsonError(w, "org_name and org_slug are required", http.StatusBadRequest)
		return
	}
	if req.AppName == "" {
		req.AppName = "My Application"
	}

	// Idempotency guard: if user already has an org, return conflict.
	user, err := s.Queries.GetUser(r.Context(), userID)
	if err != nil {
		jsonServerError(w, "failed to look up user", err)
		return
	}
	if user.OrgID.Valid {
		jsonError(w, "user already belongs to an organization", http.StatusConflict)
		return
	}

	tx, err := s.Pool.Begin(r.Context())
	if err != nil {
		jsonServerError(w, "database error", err)
		return
	}
	defer tx.Rollback(r.Context())

	qtx := s.Queries.WithTx(tx)

	// Create org
	org, err := qtx.CreateOrganization(r.Context(), db.CreateOrganizationParams{
		Name: req.OrgName,
		Slug: req.OrgSlug,
	})
	if err != nil {
		jsonError(w, "failed to create organization (slug may be taken)", http.StatusConflict)
		return
	}

	// Link user to org
	if err := qtx.SetUserOrg(r.Context(), db.SetUserOrgParams{
		OrgID: pgtype.UUID{Bytes: org.ID, Valid: true},
		ID:    userID,
	}); err != nil {
		jsonServerError(w, "failed to link user to organization", err)
		return
	}

	// Create default app
	app, err := qtx.CreateApplication(r.Context(), db.CreateApplicationParams{
		OrgID:  org.ID,
		Name:   req.AppName,
		Status: "active",
	})
	if err != nil {
		jsonServerError(w, "failed to create application", err)
		return
	}

	// Create default agent config for the app
	_, err = qtx.UpsertAppAgentConfig(r.Context(), db.UpsertAppAgentConfigParams{
		AppID:                app.ID,
		Model:                "claude-sonnet-4-6",
		Mode:                 "off",
		ScheduleIntervalSecs: 60,
	})
	if err != nil {
		jsonServerError(w, "failed to create agent config", err)
		return
	}

	if err := tx.Commit(r.Context()); err != nil {
		jsonServerError(w, "failed to complete onboarding", err)
		return
	}

	type onboardingResponse struct {
		Organization db.Organization `json:"organization"`
		Application  db.Application  `json:"application"`
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(onboardingResponse{
		Organization: org,
		Application:  app,
	})
}
