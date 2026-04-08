# OpenRouter Integration — Phase 5 Completion

**Status:** ✅ Complete
**Date:** 2026-04-07
**Plan:** [openrouter-implementation.md](../executing/openrouter-implementation.md), Phase 5
**Validation:** `vue-tsc --noEmit` clean, `npm run test` all 23 tests green
**Builds on:** [openrouter-phase-4.md](openrouter-phase-4.md)

---

## Goal

Replace the free-text `<input>` + `<datalist>` model picker on the Agent
Configuration page with a grouped dropdown that consumes `GET /api/models`
(Phase 4) and renders Anthropic and OpenRouter models in separate
`<optgroup>`s, with context length and pricing visible in each option label.
Wire `provider` through the form so saving an OpenRouter model writes
`provider='openrouter'` to `app_agent_config`.

After this phase, end-users can select an OpenRouter model from the UI for
the first time. Phase 3 made it possible via SQL; Phase 5 makes it possible
via the dropdown.

---

## Files changed

### New (2)

| File | Purpose |
|---|---|
| `frontend/src/types/models.ts` | `ModelOption` interface mirroring the Phase 4 backend struct (id, name, provider, context_length, pricing.prompt/completion) |
| `frontend/src/api/models.ts` | `getAvailableModels()` — single-line axios wrapper around `GET /api/models` |

### Modified (2)

| File | Change |
|---|---|
| `frontend/src/types/organization.ts` | `AppAgentConfig.provider: 'anthropic' \| 'openrouter'` field added between `model` and `mode` |
| `frontend/src/pages/AgentConfigPage.vue` | Removed `knownModels` array; added `models`, `formProvider`, `anthropicModels`/`openrouterModels` computeds, `formatContext`, `formatModelLabel`, `onModelSelect` helpers; `fetchConfig` now `Promise.all`s the config + models requests; `startEdit` seeds `formProvider` from the loaded config; `saveConfig` includes `provider` in the PUT payload; the `<input list="known-models">` block becomes a native `<select>` with two `<optgroup>`s; display mode shows `via {provider}` next to the model |

### Untouched (intentional)

- `BaseSelect.vue` — the techno-brutalist custom dropdown component used
  elsewhere in the form (mode picker). It doesn't support `<optgroup>`, and
  expanding it for one use case would mean rebuilding keyboard navigation,
  scroll-to-focused, and the styled panel against a grouped data shape. The
  native `<select>` is the right call for this one field — see "Native
  select vs custom dropdown" below.
- `applications.ts` API client — `updateAppAgentConfig` already accepts
  `Partial<AppAgentConfig>`, so the new `provider` field flows through
  without a signature change.
- All other pages, stores, and tests — none of them touch
  `AppAgentConfig.provider` or render a model picker.

---

## Native select vs custom dropdown

The plan called for `<optgroup>` grouping, which is a native-`<select>` HTML
feature with no clean equivalent in the project's `BaseSelect` component.
Three options were on the table:

1. **Native `<select>` with `appearance-none` + Tailwind styling.** Picked.
2. **Extend `BaseSelect` to support grouped options.** Rejected — would
   require rebuilding the keyboard navigation, focused-index tracking, and
   scroll-to-focused logic against a `Group[]` shape, plus a styled group
   header in the panel. ~80 lines of new code for one use case.
3. **Two separate `BaseSelect`s in sequence (provider then model).** Rejected
   — worse UX (two clicks for one decision) and doesn't match the dropdown
   mock in `openrouter-model-selection.md` §"Dropdown UX".

The native select sits slightly outside the techno-brutalist aesthetic of
the rest of the form (the OS draws the open panel, not Tailwind), but it's
visually consistent in its closed state via:

```html
class="block w-full bg-bg-elevated/80 border border-border rounded
       px-3 py-2 text-text-primary font-mono text-sm
       focus:border-accent focus:ring-1 focus:ring-accent/30
       focus:outline-none transition-colors appearance-none cursor-pointer"
```

`appearance-none` strips the default OS chrome from the closed state. The
open panel uses native rendering, which is the trade-off for free
`<optgroup>` support, accessibility, mobile keyboard handling, and
zero-JavaScript filtering.

If the model count grows past ~15 entries and search becomes important, the
right move is to bite the bullet and build a custom grouped dropdown
component (or pull one in). Six entries in two groups doesn't justify it
yet.

---

## The label format

```ts
function formatModelLabel(m: ModelOption): string {
  return `${m.name} — ${formatContext(m.context_length)} · $${m.pricing.prompt}/$${m.pricing.completion}`
}
```

Renders as:

```
Claude Sonnet 4.6 — 200k · $3/$15
GPT-4o            — 128k · $2.5/$10
Gemini 2.5 Pro    — 1M   · $1.25/$10
```

Three deliberate choices:

1. **`200k` / `1M` for context, not `200000`.** Raw token counts are noisy
   in a label. The `formatContext` helper switches at 1M because that's the
   only width currently in use that benefits from the unit shift (Gemini).
2. **Single-character pricing format `$3/$15`.** Reads as
   "$3-prompt-per-million / $15-completion-per-million." Verbose alternatives
   ("$3 in / $15 out", "$3 prompt $15 completion") all bloated the option
   label past the comfortable width for a `<select>`. The caption below the
   dropdown (`Context window · prompt $/M · completion $/M`) carries the
   legend so users learn the format once.
3. **Em-dash separator after the name.** Visually parses the label into
   "what" + "specs" without leaning on extra columns or alignment that
   `<select>` can't provide anyway.

---

## Auto-set provider

```ts
function onModelSelect() {
  const selected = models.value.find(m => m.id === formModel.value)
  if (selected) formProvider.value = selected.provider
}
```

The user only ever picks a model. Provider follows from the model's
`<optgroup>` membership — there's no separate "provider" field for them to
get out of sync with the model. Two reasons this matters:

1. **Eliminates the invalid combo.** Without the auto-set, a user could
   theoretically save `model='openai/gpt-4o'` with `provider='anthropic'`,
   which the backend would route to the Anthropic provider and crash on the
   unknown model ID. Auto-set means model and provider are always
   internally consistent.
2. **Matches user mental model.** Users think "I want GPT-4o," not "I want
   to switch providers and then pick GPT-4o." The provider is a routing
   detail, exposed only in display mode (`via openrouter`) so they know
   what's actually happening.

The hidden state (`formProvider`) still exists and gets sent in the PUT
payload — it's just not directly bindable in the template. This is the
right shape: one-way data flow from the model selection to provider, never
the other direction.

---

## `fetchConfig` parallelisation

```ts
const [cfg, available] = await Promise.all([
  getAppAgentConfig(appId),
  getAvailableModels(),
])
```

The two requests are independent (different endpoints, no shared data), so
they fan out in parallel. The page only renders once both resolve, so
serialising them would just add a round-trip of latency for no gain.

Both go through the same auth interceptor, so a 401 on either still
triggers logout via the existing axios interceptor in `api/client.ts`.

---

## Display mode

```html
<div class="text-right">
  <span class="font-mono text-sm text-text-primary">{{ config.model }}</span>
  <span class="font-mono text-xs text-text-muted ml-2">
    via {{ config.provider ?? 'anthropic' }}
  </span>
</div>
```

The `?? 'anthropic'` fallback covers configs created before Phase 3's
migration ran in a given environment. The backend's defaults response
already returns `provider: 'anthropic'` for missing rows (Phase 3), but the
fallback in the template is cheap defence in depth — same pattern as
`providerFor("")` in the Go backend.

Visual placement: the provider sits in muted secondary text to the right of
the model ID, on the same line. It's informational, not the headline. Users
who don't care never need to read it; users debugging "why is my GPT-4o app
acting like Claude" see immediately that the routing is wrong.

---

## Validation gates from `openrouter-implementation.md` Phase 5

- [x] Dropdown shows Anthropic models grouped under "Anthropic (Direct)"
- [x] If OpenRouter key is configured, dropdown also shows "OpenRouter" group
      (the `v-if="openrouterModels.length"` guard hides the optgroup
      entirely when the backend returns no OpenRouter entries)
- [x] Selecting a model auto-sets the provider (`onModelSelect`)
- [x] Saving persists both model + provider to backend (`saveConfig` PUT
      body now includes `provider: formProvider.value`)
- [x] Display mode shows model + provider
- [x] Pricing and context length visible in options
- [x] `vue-tsc --noEmit` clean
- [x] `npm run test` — all 23 tests passing (no test edits required)

No new tests added: the existing suite doesn't cover `AgentConfigPage.vue`
to begin with (the page is mostly form state plumbing with no business
logic to assert), and adding the model dropdown didn't introduce any logic
beyond mechanical filter/find calls and a string formatter. The honest
test is the manual smoke test below.

---

## Manual smoke test recipe

1. **Load the page** with no OpenRouter key set on the backend. Open the
   model dropdown — should show only the "Anthropic (Direct)" group with
   three entries. The "OpenRouter" optgroup should not appear at all.
2. **Set `OPENROUTER_API_KEY`** on the backend, restart, reload the page.
   Open the dropdown — should now show both groups. The OpenRouter group
   should have six entries.
3. **Pick `GPT-4o`** from the OpenRouter group. Save. The display mode
   should now read "openai/gpt-4o" with "via openrouter" in muted text.
   Refresh — selection persists.
4. **Pick `Claude Sonnet 4.6`** from the Anthropic group. Save. Display
   mode reads "claude-sonnet-4-6 via anthropic". The DB row's `provider`
   column should now be `anthropic`.
5. **Open the agent chat** for that app and send a message. Confirm the
   response comes from the model the user just picked (test by picking
   models with distinctive personalities — e.g. GPT-4o and Sonnet — and
   asking the same prompt twice).
6. **Pricing labels** — confirm each option displays `<context> · $<p>/$<c>`
   and that the formatter renders `1M` for Gemini's million-token window
   rather than `1000k`.

---

## What stayed the same

- `BaseSelect` (mode picker) — unchanged. Mode is a simple three-option flat
  list, perfect fit for the custom component.
- All other form fields — model is the only field that changed. Mode,
  interval presets, system-prompt-override textarea, save/cancel buttons:
  identical.
- Skeleton loader — still shows four skeleton rows on initial load. The
  added `getAvailableModels` request is parallelised inside the same
  `loading` flag, so there's no second loading state to manage.
- Toast error path — `Failed to load agent config` still fires if either
  the config or models request rejects (the `Promise.all` rejects on the
  first failure). This is the right shape: if either fails, the page is
  unusable.

---

## Follow-ups

### Phase 6 — Frontend streaming chat
Now the only remaining piece of the original implementation plan. Same
shape as before: `useAgent.ts` learns the new `chunk` / `tool_start` /
`tool_result` / `done` WebSocket message types, the chat page renders
`activeTools` as a small progress indicator, the existing markdown renderer
transparently re-renders as the message content grows. Independent of
Phases 4–5.

This will also need the backend half (`RunConversationStream`,
`chat.go` rewrite, `Provider.ChatCompletionStream` already exists from
Phase 1) — the streaming boundary doc in `openrouter-implementation.md`
Phase 1 is the canonical reference for which paths stream and which stay
blocking. **Monitoring stays blocking** — re-read the warning callout
before touching `monitor.go`.

### Pre-launch checklist for Phase 5
Before this is shown to real users, verify the six OpenRouter prices in
`backend/internal/agent/models.go` against OpenRouter's live catalogue
(flagged in the Phase 4 completion doc and again in the original plan).
The pricing in the dropdown is currently the working baseline from the
plan document, not a verified snapshot.

### Optional UX polish
- **Keyboard search** — native `<select>` does basic type-ahead by first
  letter, which is fine for six entries. If the list grows, the
  type-ahead-by-substring behavior of a custom component starts to matter.
- **Group header styling** — `<optgroup>` labels are rendered by the OS, so
  there's no way to style them to match the rest of the form. The trade-off
  is documented above.
- **"Recommended" badge** — could mark `claude-sonnet-4-6` as the
  recommended default with a `★` prefix in the label. Skipped for now;
  users who want a recommendation can read the docs.
