package handlers

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"github.com/hejijunhao/heimdall/backend/internal/api/middleware"
	"github.com/hejijunhao/heimdall/backend/internal/db"
)

// resolveOrgAndRole resolves the target org and the caller's role within it.
// If the request includes an X-Org-ID header, that org is used (after
// validating membership). Otherwise falls back to the user's primary org.
// Returns the org, the role, and true on success. On failure it writes an
// HTTP error and returns zero values with false.
func (s *Server) resolveOrgAndRole(w http.ResponseWriter, r *http.Request) (db.Organization, db.OrgMemberRole, bool) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		jsonError(w, "missing user context", http.StatusUnauthorized)
		return db.Organization{}, "", false
	}

	org, membership, err := s.resolveOrgForUser(r, userID)
	if err != nil {
		jsonError(w, "organization not found or access denied", http.StatusNotFound)
		return db.Organization{}, "", false
	}

	return org, membership.Role, true
}

// resolveOrgForUser reads the X-Org-ID header to determine which org the
// request targets, validates membership, and returns the org + membership.
// Falls back to the user's primary org when the header is absent.
//
// All reads run inside one UserQueries scope so org_members and
// organizations are queried under app.current_user_id == userID. The
// transaction is read-only — we don't commit, just close, which rolls
// back any incidental state. This is the entry-point helper that nearly
// every JWT-authed handler funnels through; routing every read through
// it (rather than direct s.Queries) is the lowest-friction way to bring
// the bulk of Class D into the handoff pattern.
func (s *Server) resolveOrgForUser(r *http.Request, userID uuid.UUID) (db.Organization, db.OrgMember, error) {
	queries, _, done, err := s.UserQueries(r.Context(), userID)
	if err != nil {
		return db.Organization{}, db.OrgMember{}, err
	}
	defer done()

	if orgIDStr := r.Header.Get("X-Org-ID"); orgIDStr != "" {
		orgID, err := uuid.Parse(orgIDStr)
		if err != nil {
			return db.Organization{}, db.OrgMember{}, err
		}
		membership, err := queries.GetOrgMembership(r.Context(), db.GetOrgMembershipParams{
			UserID: userID,
			OrgID:  orgID,
		})
		if err != nil {
			return db.Organization{}, db.OrgMember{}, err
		}
		org, err := queries.GetOrganization(r.Context(), orgID)
		if err != nil {
			return db.Organization{}, db.OrgMember{}, err
		}
		return org, membership, nil
	}

	// Fallback: primary org (earliest membership).
	org, err := queries.GetOrganizationByUser(r.Context(), userID)
	if err != nil {
		return db.Organization{}, db.OrgMember{}, err
	}
	membership, err := queries.GetOrgMembership(r.Context(), db.GetOrgMembershipParams{
		UserID: userID,
		OrgID:  org.ID,
	})
	if err != nil {
		return db.Organization{}, db.OrgMember{}, err
	}
	return org, membership, nil
}

// hasMinRole checks whether the given role meets the required minimum.
// Hierarchy: owner > admin > member.
func hasMinRole(actual db.OrgMemberRole, required db.OrgMemberRole) bool {
	hierarchy := map[db.OrgMemberRole]int{
		db.OrgMemberRoleOwner:  3,
		db.OrgMemberRoleAdmin:  2,
		db.OrgMemberRoleMember: 1,
	}
	return hierarchy[actual] >= hierarchy[required]
}

// ListOrgMembers returns all members of the caller's organization.
// GET /api/org/members
func (s *Server) ListOrgMembers(w http.ResponseWriter, r *http.Request) {
	org, _, ok := s.resolveOrgAndRole(w, r)
	if !ok {
		return
	}
	userID, _ := middleware.UserIDFromContext(r.Context())

	var members []db.ListOrgMembersRow
	err := s.Pools.WithUserQueries(r.Context(), userID, func(q *db.Queries) error {
		var e error
		members, e = q.ListOrgMembers(r.Context(), org.ID)
		return e
	})
	if err != nil {
		jsonServerError(w, "failed to list members", err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(members)
}

type inviteMemberRequest struct {
	Email string `json:"email"`
	Role  string `json:"role"`
}

// InviteMember adds a user to the caller's organization by email.
// For V1 this is an "instant add" — the user must already exist.
//
// KNOWN LIMITATION: Leaf data tables (connections, log_buffer, conversations,
// agent_log, investigations) are still user-scoped via user_id, not org-scoped.
// New members will see an empty dashboard until their own queries populate data.
// Full org-scoped access requires rewriting leaf queries and RLS policies to join
// through org_members — tracked for a future release.
// POST /api/org/members/invite
func (s *Server) InviteMember(w http.ResponseWriter, r *http.Request) {
	org, callerRole, ok := s.resolveOrgAndRole(w, r)
	if !ok {
		return
	}

	if !hasMinRole(callerRole, db.OrgMemberRoleAdmin) {
		jsonError(w, "admin or owner role required", http.StatusForbidden)
		return
	}

	var req inviteMemberRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonError(w, "invalid request body", http.StatusBadRequest)
		return
	}

	if req.Email == "" {
		jsonError(w, "email is required", http.StatusBadRequest)
		return
	}
	if !isValidEmail(req.Email) {
		jsonError(w, "invalid email address", http.StatusBadRequest)
		return
	}

	// Default to member role if not specified.
	role := db.OrgMemberRoleMember
	if req.Role != "" {
		switch db.OrgMemberRole(req.Role) {
		case db.OrgMemberRoleMember, db.OrgMemberRoleAdmin:
			role = db.OrgMemberRole(req.Role)
		case db.OrgMemberRoleOwner:
			// Only owners can invite as owner.
			if callerRole != db.OrgMemberRoleOwner {
				jsonError(w, "only owners can grant owner role", http.StatusForbidden)
				return
			}
			role = db.OrgMemberRoleOwner
		default:
			jsonError(w, "role must be one of: owner, admin, member", http.StatusBadRequest)
			return
		}
	}

	userID, _ := middleware.UserIDFromContext(r.Context())

	// Resolve the invitee's user_id via the SECURITY DEFINER helper from
	// migration 041. Under FORCE the regular GetUserByEmail can't find
	// users outside the caller's orgs — which is exactly the population an
	// invite needs to address — so the helper is the only correct path.
	// The query COALESCEs missing-user into uuid.Nil (sqlc v1.30 won't
	// infer nullability through the SECURITY DEFINER call); we treat
	// uuid.Nil as the "not found" sentinel here. The membership check
	// still runs through the regular RLS-scoped query.
	var targetUserID uuid.UUID
	var alreadyMember bool
	if err := s.Pools.WithUserQueries(r.Context(), userID, func(q *db.Queries) error {
		var e error
		targetUserID, e = q.LookupUserIDForInvite(r.Context(), req.Email)
		if e != nil {
			return e
		}
		if targetUserID == uuid.Nil {
			return nil
		}
		_, e = q.GetOrgMembership(r.Context(), db.GetOrgMembershipParams{
			UserID: targetUserID,
			OrgID:  org.ID,
		})
		alreadyMember = (e == nil)
		return nil
	}); err != nil {
		jsonServerError(w, "database error", err)
		return
	}
	if targetUserID == uuid.Nil {
		jsonError(w, "user not found — they must have a Heimdall account first", http.StatusNotFound)
		return
	}
	if alreadyMember {
		jsonError(w, "user is already a member of this organization", http.StatusConflict)
		return
	}

	queries, commit, done, err := s.UserQueries(r.Context(), userID)
	if err != nil {
		jsonServerError(w, "database error", err)
		return
	}
	defer done()

	if err := queries.CreateOrgMember(r.Context(), db.CreateOrgMemberParams{
		UserID: targetUserID,
		OrgID:  org.ID,
		Role:   role,
	}); err != nil {
		// Catch unique constraint violation from TOCTOU race (concurrent invite).
		if strings.Contains(err.Error(), "duplicate key") || strings.Contains(err.Error(), "unique constraint") {
			jsonError(w, "user is already a member of this organization", http.StatusConflict)
			return
		}
		jsonServerError(w, "failed to add member", err)
		return
	}

	if err := commit(); err != nil {
		jsonServerError(w, "failed to save changes", err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{
		"user_id": targetUserID.String(),
		"email":   req.Email,
		"role":    string(role),
	})
}

type updateRoleRequest struct {
	Role string `json:"role"`
}

// UpdateMemberRole changes a member's role within the organization.
// Only owners can change roles.
// PUT /api/org/members/{userId}/role
func (s *Server) UpdateMemberRole(w http.ResponseWriter, r *http.Request) {
	org, callerRole, ok := s.resolveOrgAndRole(w, r)
	if !ok {
		return
	}

	if callerRole != db.OrgMemberRoleOwner {
		jsonError(w, "owner role required", http.StatusForbidden)
		return
	}

	targetUserID, err := uuid.Parse(chi.URLParam(r, "userId"))
	if err != nil {
		jsonError(w, "invalid user id", http.StatusBadRequest)
		return
	}

	var req updateRoleRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonError(w, "invalid request body", http.StatusBadRequest)
		return
	}

	newRole := db.OrgMemberRole(req.Role)
	switch newRole {
	case db.OrgMemberRoleOwner, db.OrgMemberRoleAdmin, db.OrgMemberRoleMember:
		// valid
	default:
		jsonError(w, "role must be one of: owner, admin, member", http.StatusBadRequest)
		return
	}

	callerID, _ := middleware.UserIDFromContext(r.Context())
	queries, commit, done, err := s.UserQueries(r.Context(), callerID)
	if err != nil {
		jsonServerError(w, "database error", err)
		return
	}
	defer done()

	// Verify target is actually a member (inside the transaction for consistency).
	currentMembership, err := queries.GetOrgMembership(r.Context(), db.GetOrgMembershipParams{
		UserID: targetUserID,
		OrgID:  org.ID,
	})
	if err != nil {
		jsonError(w, "member not found", http.StatusNotFound)
		return
	}

	// Prevent demoting the sole owner — the org would become unmanageable.
	// Checked inside the transaction to prevent TOCTOU races from concurrent
	// requests both passing the guard before either commits.
	if currentMembership.Role == db.OrgMemberRoleOwner && newRole != db.OrgMemberRoleOwner {
		ownerCount, err := queries.CountOrgOwners(r.Context(), org.ID)
		if err != nil {
			jsonServerError(w, "failed to count owners", err)
			return
		}
		if ownerCount <= 1 {
			jsonError(w, "cannot demote the sole owner of an organization", http.StatusConflict)
			return
		}
	}

	if err := queries.UpdateOrgMemberRole(r.Context(), db.UpdateOrgMemberRoleParams{
		UserID: targetUserID,
		OrgID:  org.ID,
		Role:   newRole,
	}); err != nil {
		jsonServerError(w, "failed to update role", err)
		return
	}

	if err := commit(); err != nil {
		jsonServerError(w, "failed to save changes", err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// RemoveMember removes a user from the organization.
// Admins and owners can remove members. Owners cannot remove themselves
// if they are the sole owner.
// DELETE /api/org/members/{userId}
func (s *Server) RemoveMember(w http.ResponseWriter, r *http.Request) {
	org, callerRole, ok := s.resolveOrgAndRole(w, r)
	if !ok {
		return
	}

	if !hasMinRole(callerRole, db.OrgMemberRoleAdmin) {
		jsonError(w, "admin or owner role required", http.StatusForbidden)
		return
	}

	targetUserID, err := uuid.Parse(chi.URLParam(r, "userId"))
	if err != nil {
		jsonError(w, "invalid user id", http.StatusBadRequest)
		return
	}

	callerID, _ := middleware.UserIDFromContext(r.Context())
	queries, commit, done, err := s.UserQueries(r.Context(), callerID)
	if err != nil {
		jsonServerError(w, "database error", err)
		return
	}
	defer done()

	// Verify target is actually a member (inside the transaction for consistency).
	targetMembership, err := queries.GetOrgMembership(r.Context(), db.GetOrgMembershipParams{
		UserID: targetUserID,
		OrgID:  org.ID,
	})
	if err != nil {
		jsonError(w, "member not found", http.StatusNotFound)
		return
	}

	// Prevent removing owners unless caller is also an owner.
	if targetMembership.Role == db.OrgMemberRoleOwner && callerRole != db.OrgMemberRoleOwner {
		jsonError(w, "only owners can remove other owners", http.StatusForbidden)
		return
	}

	// Prevent removing the sole owner — the org would become unmanageable.
	// Checked inside the transaction to prevent TOCTOU races from concurrent
	// requests both passing the guard before either commits.
	if targetMembership.Role == db.OrgMemberRoleOwner {
		ownerCount, err := queries.CountOrgOwners(r.Context(), org.ID)
		if err != nil {
			jsonServerError(w, "failed to count owners", err)
			return
		}
		if ownerCount <= 1 {
			jsonError(w, "cannot remove the sole owner of an organization", http.StatusConflict)
			return
		}
	}

	if err := queries.DeleteOrgMember(r.Context(), db.DeleteOrgMemberParams{
		UserID: targetUserID,
		OrgID:  org.ID,
	}); err != nil {
		jsonServerError(w, "failed to remove member", err)
		return
	}

	if err := commit(); err != nil {
		jsonServerError(w, "failed to save changes", err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
