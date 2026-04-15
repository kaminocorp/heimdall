package handlers_test

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestListOrgMembers(t *testing.T) {
	env := testSetup(t)

	rr := env.request(t, http.MethodGet, "/api/org/members", nil)
	require.Equal(t, http.StatusOK, rr.Code)

	var members []map[string]interface{}
	require.NoError(t, json.NewDecoder(rr.Body).Decode(&members))

	// testSetup creates one owner.
	require.Len(t, members, 1)
	assert.Equal(t, env.UserID, members[0]["user_id"])
	assert.Equal(t, "owner", members[0]["role"])
}

func TestInviteMember(t *testing.T) {
	env := testSetup(t)

	// Create a second user to invite.
	inviteeID := createBareUser(t, env)

	// Look up invitee email.
	var email string
	err := env.Pool.QueryRow(t.Context(), "SELECT email FROM users WHERE id = $1", inviteeID).Scan(&email)
	require.NoError(t, err)

	body := map[string]string{
		"email": email,
		"role":  "member",
	}

	rr := env.request(t, http.MethodPost, "/api/org/members/invite", body)
	require.Equal(t, http.StatusCreated, rr.Code)

	var resp map[string]string
	require.NoError(t, json.NewDecoder(rr.Body).Decode(&resp))
	assert.Equal(t, inviteeID.String(), resp["user_id"])
	assert.Equal(t, "member", resp["role"])

	// Verify the member appears in the list.
	rr = env.request(t, http.MethodGet, "/api/org/members", nil)
	require.Equal(t, http.StatusOK, rr.Code)

	var members []map[string]interface{}
	require.NoError(t, json.NewDecoder(rr.Body).Decode(&members))
	assert.Len(t, members, 2)
}

func TestInviteMember_DuplicateReturns409(t *testing.T) {
	env := testSetup(t)

	inviteeID := createBareUser(t, env)
	var email string
	err := env.Pool.QueryRow(t.Context(), "SELECT email FROM users WHERE id = $1", inviteeID).Scan(&email)
	require.NoError(t, err)

	body := map[string]string{"email": email, "role": "member"}

	// First invite succeeds.
	rr := env.request(t, http.MethodPost, "/api/org/members/invite", body)
	require.Equal(t, http.StatusCreated, rr.Code)

	// Second invite returns conflict.
	rr = env.request(t, http.MethodPost, "/api/org/members/invite", body)
	assert.Equal(t, http.StatusConflict, rr.Code)
}

func TestInviteMember_MissingEmail(t *testing.T) {
	env := testSetup(t)

	rr := env.request(t, http.MethodPost, "/api/org/members/invite", map[string]string{"role": "member"})
	assert.Equal(t, http.StatusBadRequest, rr.Code)
}

func TestInviteMember_MemberCannotInvite(t *testing.T) {
	env := testSetup(t)

	// Add a member-role user.
	memberID := createBareUser(t, env)
	_, err := env.Pool.Exec(t.Context(),
		"INSERT INTO org_members (user_id, org_id, role) VALUES ($1, $2, 'member')",
		memberID, env.OrgID,
	)
	require.NoError(t, err)

	// Authenticate as the member.
	memberEnv := envForUser(t, env, memberID)

	// Create a third user to invite.
	thirdID := createBareUser(t, env)
	var email string
	err = env.Pool.QueryRow(t.Context(), "SELECT email FROM users WHERE id = $1", thirdID).Scan(&email)
	require.NoError(t, err)

	rr := memberEnv.request(t, http.MethodPost, "/api/org/members/invite", map[string]string{
		"email": email,
		"role":  "member",
	})
	assert.Equal(t, http.StatusForbidden, rr.Code)
}

func TestUpdateMemberRole(t *testing.T) {
	env := testSetup(t)

	// Add a member to update.
	memberID := createBareUser(t, env)
	_, err := env.Pool.Exec(t.Context(),
		"INSERT INTO org_members (user_id, org_id, role) VALUES ($1, $2, 'member')",
		memberID, env.OrgID,
	)
	require.NoError(t, err)

	// Promote to admin.
	rr := env.request(t, http.MethodPut, "/api/org/members/"+memberID.String()+"/role",
		map[string]string{"role": "admin"})
	assert.Equal(t, http.StatusNoContent, rr.Code)

	// Verify role changed.
	var role string
	err = env.Pool.QueryRow(t.Context(),
		"SELECT role FROM org_members WHERE user_id = $1 AND org_id = $2", memberID, env.OrgID,
	).Scan(&role)
	require.NoError(t, err)
	assert.Equal(t, "admin", role)
}

func TestUpdateMemberRole_NonOwnerForbidden(t *testing.T) {
	env := testSetup(t)

	// Add an admin.
	adminID := createBareUser(t, env)
	_, err := env.Pool.Exec(t.Context(),
		"INSERT INTO org_members (user_id, org_id, role) VALUES ($1, $2, 'admin')",
		adminID, env.OrgID,
	)
	require.NoError(t, err)

	// Add a member to target.
	memberID := createBareUser(t, env)
	_, err = env.Pool.Exec(t.Context(),
		"INSERT INTO org_members (user_id, org_id, role) VALUES ($1, $2, 'member')",
		memberID, env.OrgID,
	)
	require.NoError(t, err)

	// Admin tries to change role — should fail (owner required).
	adminEnv := envForUser(t, env, adminID)
	rr := adminEnv.request(t, http.MethodPut, "/api/org/members/"+memberID.String()+"/role",
		map[string]string{"role": "admin"})
	assert.Equal(t, http.StatusForbidden, rr.Code)
}

func TestRemoveMember(t *testing.T) {
	env := testSetup(t)

	memberID := createBareUser(t, env)
	_, err := env.Pool.Exec(t.Context(),
		"INSERT INTO org_members (user_id, org_id, role) VALUES ($1, $2, 'member')",
		memberID, env.OrgID,
	)
	require.NoError(t, err)

	rr := env.request(t, http.MethodDelete, "/api/org/members/"+memberID.String(), nil)
	assert.Equal(t, http.StatusNoContent, rr.Code)

	// Verify member is gone.
	rr = env.request(t, http.MethodGet, "/api/org/members", nil)
	require.Equal(t, http.StatusOK, rr.Code)

	var members []map[string]interface{}
	require.NoError(t, json.NewDecoder(rr.Body).Decode(&members))
	assert.Len(t, members, 1) // only the owner remains
}

func TestRemoveMember_SoleOwnerBlocked(t *testing.T) {
	env := testSetup(t)

	// Try to remove self (sole owner) — should be blocked.
	rr := env.request(t, http.MethodDelete, "/api/org/members/"+env.UserID, nil)
	assert.Equal(t, http.StatusConflict, rr.Code)
}

func TestRemoveMember_MemberCannotRemove(t *testing.T) {
	env := testSetup(t)

	// Add two members.
	member1ID := createBareUser(t, env)
	_, err := env.Pool.Exec(t.Context(),
		"INSERT INTO org_members (user_id, org_id, role) VALUES ($1, $2, 'member')",
		member1ID, env.OrgID,
	)
	require.NoError(t, err)

	member2ID := createBareUser(t, env)
	_, err = env.Pool.Exec(t.Context(),
		"INSERT INTO org_members (user_id, org_id, role) VALUES ($1, $2, 'member')",
		member2ID, env.OrgID,
	)
	require.NoError(t, err)

	// Member tries to remove another member — forbidden.
	memberEnv := envForUser(t, env, member1ID)
	rr := memberEnv.request(t, http.MethodDelete, "/api/org/members/"+member2ID.String(), nil)
	assert.Equal(t, http.StatusForbidden, rr.Code)
}
