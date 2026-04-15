# Phase 5 — Team Management Page

**Status:** Complete
**Date:** 2026-04-13
**Plan:** `docs/executing/org-overview-implementation.md`

---

## What changed

The `OrgTeamPage.vue` shell from Phase 3 is now a fully functional team
management page where org owners and admins can view members, invite new
ones by email, change roles, and remove members.

---

## Modified files

### `pages/org/OrgTeamPage.vue`

Replaced placeholder shell with full implementation. The page has three main
sections:

**1. Invite section** (admin+ only)

- Email input + role selector + "Send invite" button
- Role options: Member, Admin, Owner (Owner only visible to owners)
- Success/error feedback banners below the form
- Role description cards explaining what each role can do
- Hidden entirely for members (read-only access)

**2. Member list**

Each member row shows:
- **Avatar** — initial letter in a rounded circle
- **Email** — with "(you)" label for the current user
- **Joined date** — when the membership was created
- **Role badge/selector**:
  - Owners see a `<select>` dropdown on other members to change roles inline
  - Non-owners (or viewing themselves) see a read-only badge
  - Badge colours: green for owner, amber for admin, neutral for member
- **Remove button** — visible to admins+ on other members (not on self)

Current user's row is highlighted with a subtle accent background.

**3. Remove confirmation modal**

- Teleported to body with backdrop blur
- Shows the member's email and warns about project access loss
- Cancel / Remove (red) buttons with loading state
- Cleans up on cancel or after successful removal

**Permission gating:**

The page reads `appStore.organization?.role` (set by the Phase 2 `GET /api/org`
response which now includes `role`) to determine what the caller can do:

| Caller role | Can see list | Can invite | Can change roles | Can remove |
|-------------|:---:|:---:|:---:|:---:|
| Member | Yes | No | No | No |
| Admin | Yes | Yes (member/admin) | No | Yes (members only) |
| Owner | Yes | Yes (any role) | Yes | Yes (anyone except sole self) |

All authorization is also enforced server-side (Phase 2 handlers) — the UI
gating is defence-in-depth so users don't see controls they can't use.

---

## Data flow

```
OrgTeamPage mounts
  → listOrgMembers() → GET /api/org/members
  → Response: OrgMember[] (user_id, email, role, created_at)

User submits invite form
  → inviteOrgMember(email, role) → POST /api/org/members/invite
  → On success: show green banner, clear form, refetch list
  → On error: show red banner with server message

Owner changes a role via dropdown
  → updateMemberRole(userId, newRole) → PUT /api/org/members/{userId}/role
  → On success: update local member.role reactively (no refetch needed)

Admin clicks Remove → confirm modal
  → removeOrgMember(userId) → DELETE /api/org/members/{userId}
  → On success: close modal, refetch list
```

---

## Verification

| Check | Result |
|-------|--------|
| `vue-tsc --noEmit` | Clean |
| `vite build` | Clean (1.72s) |
| `vitest run` | 52/52 tests pass |

---

## What's next

Phase 6 builds the Organisation Settings page (`/org/settings`) — editable
name/slug, org ID, and the danger zone (delete org).
