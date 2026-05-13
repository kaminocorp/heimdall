package handlers

import (
	"encoding/json"
	"net/http"
	"regexp"

	"github.com/hejijunhao/heimdall/backend/internal/agent"
	"github.com/hejijunhao/heimdall/backend/internal/api/middleware"
	"github.com/hejijunhao/heimdall/backend/internal/db"
)

var slugRe = regexp.MustCompile(`^[a-z0-9][a-z0-9-]{1,48}[a-z0-9]$`)

func (s *Server) GetOrganization(w http.ResponseWriter, r *http.Request) {
	org, callerRole, ok := s.resolveOrgAndRole(w, r)
	if !ok {
		return
	}

	type orgWithRole struct {
		db.Organization
		Role db.OrgMemberRole `json:"role"`
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(orgWithRole{
		Organization: org,
		Role:         callerRole,
	})
}

// ListUserOrganizations returns all organizations the authenticated user belongs to.
// GET /api/orgs
func (s *Server) ListUserOrganizations(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		jsonError(w, "missing user context", http.StatusUnauthorized)
		return
	}

	var orgs []db.ListOrganizationsByUserRow
	err := s.Pools.WithUserQueries(r.Context(), userID, func(q *db.Queries) error {
		var e error
		orgs, e = q.ListOrganizationsByUser(r.Context(), userID)
		return e
	})
	if err != nil {
		jsonServerError(w, "failed to list organizations", err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(orgs)
}

type createOrganizationRequest struct {
	Name string `json:"name"`
	Slug string `json:"slug"`
}

// CreateNewOrganization creates a new org and adds the caller as owner.
// Unlike Onboard (which also creates a default app), this just creates the org.
// POST /api/orgs
func (s *Server) CreateNewOrganization(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		jsonError(w, "missing user context", http.StatusUnauthorized)
		return
	}

	var req createOrganizationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonError(w, "invalid request body", http.StatusBadRequest)
		return
	}

	if req.Name == "" || req.Slug == "" {
		jsonError(w, "name and slug are required", http.StatusBadRequest)
		return
	}
	if !slugRe.MatchString(req.Slug) {
		jsonError(w, "slug must be 3-50 characters, lowercase alphanumeric and hyphens only, must start and end with a letter or digit", http.StatusBadRequest)
		return
	}

	queries, commit, done, err := s.UserQueries(r.Context(), userID)
	if err != nil {
		jsonServerError(w, "database error", err)
		return
	}
	defer done()

	org, err := queries.CreateOrganization(r.Context(), db.CreateOrganizationParams{
		Name: req.Name,
		Slug: req.Slug,
	})
	if err != nil {
		jsonError(w, "failed to create organization (slug may be taken)", http.StatusConflict)
		return
	}

	if err := queries.CreateOrgMember(r.Context(), db.CreateOrgMemberParams{
		UserID: userID,
		OrgID:  org.ID,
		Role:   db.OrgMemberRoleOwner,
	}); err != nil {
		jsonServerError(w, "failed to add owner membership", err)
		return
	}

	if err := commit(); err != nil {
		jsonServerError(w, "failed to create organization", err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(org)
}

type updateOrganizationRequest struct {
	Name string `json:"name"`
	Slug string `json:"slug"`
}

// UpdateOrganization updates the org name and/or slug. Requires admin+ role.
// PUT /api/org
func (s *Server) UpdateOrganization(w http.ResponseWriter, r *http.Request) {
	org, callerRole, ok := s.resolveOrgAndRole(w, r)
	if !ok {
		return
	}

	if !hasMinRole(callerRole, db.OrgMemberRoleAdmin) {
		jsonError(w, "admin or owner role required", http.StatusForbidden)
		return
	}

	var req updateOrganizationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonError(w, "invalid request body", http.StatusBadRequest)
		return
	}

	// Use existing values as defaults.
	if req.Name == "" {
		req.Name = org.Name
	}
	if req.Slug == "" {
		req.Slug = org.Slug
	}

	if !slugRe.MatchString(req.Slug) {
		jsonError(w, "slug must be 3-50 characters, lowercase alphanumeric and hyphens, no leading/trailing hyphen", http.StatusBadRequest)
		return
	}

	userID, _ := middleware.UserIDFromContext(r.Context())
	queries, commit, done, err := s.UserQueries(r.Context(), userID)
	if err != nil {
		jsonServerError(w, "database error", err)
		return
	}
	defer done()

	updated, err := queries.UpdateOrganization(r.Context(), db.UpdateOrganizationParams{
		ID:   org.ID,
		Name: req.Name,
		Slug: req.Slug,
	})
	if err != nil {
		jsonError(w, "failed to update organization (slug may be taken)", http.StatusConflict)
		return
	}

	if err := commit(); err != nil {
		jsonServerError(w, "failed to save changes", err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(updated)
}

type deleteOrganizationRequest struct {
	Confirm string `json:"confirm"` // Must match org slug
}

// DeleteOrganization permanently deletes the org and all associated data.
// Owner-only. Requires slug confirmation in the request body.
// DELETE /api/org
func (s *Server) DeleteOrganization(w http.ResponseWriter, r *http.Request) {
	org, callerRole, ok := s.resolveOrgAndRole(w, r)
	if !ok {
		return
	}

	if callerRole != db.OrgMemberRoleOwner {
		jsonError(w, "owner role required", http.StatusForbidden)
		return
	}

	var req deleteOrganizationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonError(w, "invalid request body", http.StatusBadRequest)
		return
	}

	if req.Confirm != org.Slug {
		jsonError(w, "confirmation does not match organization slug", http.StatusBadRequest)
		return
	}

	userID, _ := middleware.UserIDFromContext(r.Context())
	queries, commit, done, err := s.UserQueries(r.Context(), userID)
	if err != nil {
		jsonServerError(w, "database error", err)
		return
	}
	defer done()

	if err := queries.DeleteOrganization(r.Context(), org.ID); err != nil {
		jsonServerError(w, "failed to delete organization", err)
		return
	}

	if err := commit(); err != nil {
		jsonServerError(w, "failed to save changes", err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
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
	if !slugRe.MatchString(req.OrgSlug) {
		jsonError(w, "org_slug must be 3-50 characters, lowercase alphanumeric and hyphens only, must start and end with a letter or digit", http.StatusBadRequest)
		return
	}
	if req.AppName == "" {
		req.AppName = "My Application"
	}

	// Idempotency guard: if user already has an org, return conflict.
	var hasOrg bool
	err := s.Pools.WithUserQueries(r.Context(), userID, func(q *db.Queries) error {
		var e error
		hasOrg, e = q.HasOrgMembership(r.Context(), userID)
		return e
	})
	if err != nil {
		jsonServerError(w, "failed to look up user", err)
		return
	}
	if hasOrg {
		jsonError(w, "user already belongs to an organization", http.StatusConflict)
		return
	}

	queries, commit, done, err := s.UserQueries(r.Context(), userID)
	if err != nil {
		jsonServerError(w, "database error", err)
		return
	}
	defer done()

	// Create org
	org, err := queries.CreateOrganization(r.Context(), db.CreateOrganizationParams{
		Name: req.OrgName,
		Slug: req.OrgSlug,
	})
	if err != nil {
		jsonError(w, "failed to create organization (slug may be taken)", http.StatusConflict)
		return
	}

	// Link user to org as owner
	if err := queries.CreateOrgMember(r.Context(), db.CreateOrgMemberParams{
		UserID: userID,
		OrgID:  org.ID,
		Role:   db.OrgMemberRoleOwner,
	}); err != nil {
		jsonServerError(w, "failed to link user to organization", err)
		return
	}

	// Create default app
	app, err := queries.CreateApplication(r.Context(), db.CreateApplicationParams{
		OrgID:  org.ID,
		Name:   req.AppName,
		Status: "active",
	})
	if err != nil {
		jsonServerError(w, "failed to create application", err)
		return
	}

	// Create default agent config for the app
	_, err = queries.UpsertAppAgentConfig(r.Context(), db.UpsertAppAgentConfigParams{
		AppID:                app.ID,
		Model:                agent.DefaultModelID,
		Provider:             "anthropic",
		Mode:                 "off",
		ScheduleIntervalSecs: 60,
	})
	if err != nil {
		jsonServerError(w, "failed to create agent config", err)
		return
	}

	if err := commit(); err != nil {
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
