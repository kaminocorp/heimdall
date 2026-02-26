On the Connections page, where we list connections, let's make two modifications:

1. There should be an edit button that allows me to modify Connections and resave them (currently I can only delete and add).
2. Editing and resaving should re-trigger the test/check.
3. There should be a "ping" button which manually activates/runs the connectivity check.

---

## Implementation Plan

The backend `PUT /connections/:id` handler and the frontend API client (`updateConnection`) and store action (`updateConnection`) already exist and are fully wired. Only UI-layer changes are needed.

### 1. ConnectionCard — add Edit and Ping buttons

- Add an **edit** (pencil) button alongside the existing delete button in the hover-reveal action area. Emits `edit` with the connection object.
- Add a **ping** button (signal/refresh icon). Emits `test` with the connection ID. Visually sits next to edit/delete.
- Both buttons use the same hover-reveal pattern as the existing delete button.

### 2. ConnectionList — forward new events

- Forward the new `edit` and `test` events from child cards up to the page.

### 3. ConnectionForm — support edit mode

- Add an optional `initialValues?: Connection` prop.
- When provided, pre-populate all fields (name, type, direction, config) from the existing connection on mount.
- Lock the **type** dropdown when editing (changing type mid-edit would invalidate the config fields).
- Change the submit button label from "Create" to "Save" when editing.
- Emit the same `submit` event — the page handler decides whether to call create or update.

### 4. ConnectionsPage — wire edit and ping flows

- Add an `editingConnection` ref (`Connection | null`).
- **Edit flow:** On `edit` event, set `editingConnection` and show the form. On submit, call `store.updateConnection(id, payload)` instead of `createConnection`. After a successful update, trigger `store.testConnection(id)` (same as create). Clear `editingConnection` and hide form.
- **Ping flow:** On `test` event, call `store.testConnection(id)`. Show error banner if the test fails, same as the post-create test.
- The "New Connection" button clears `editingConnection` before showing the form, so the form opens in create mode.

### Files changed (4)

| # | File | Change |
|---|------|--------|
| 1 | `frontend/src/components/connections/ConnectionCard.vue` | Add edit + ping buttons, emit `edit` and `test` |
| 2 | `frontend/src/components/connections/ConnectionList.vue` | Forward `edit` and `test` events |
| 3 | `frontend/src/components/connections/ConnectionForm.vue` | Accept `initialValues` prop, pre-populate fields, lock type on edit, dynamic button label |
| 4 | `frontend/src/pages/ConnectionsPage.vue` | `editingConnection` ref, edit handler, ping handler, route submit to create vs update |

No backend changes required.

### Questions

None — requirements are clear. The existing backend and store layers cover everything needed.