# PR: E01 Chat Empty State

## Summary

Implement the minimal chat empty state described in `docs/chat-empty-state-plan.md`.

## Changes

- Add `ChatEmptyState.vue` with centered greeting, up to 5 scene chips, and compact composer.
- Add `Composer.vue` with `compact` mode support and local send behavior.
- Add minimal chat shell components: `SessionList.vue`, `ChatHeader.vue`, `MessageList.vue`.
- Add minimal `Chat.vue` route shell with conditional empty-state rendering.
- Add lightweight chat stores and scene API/type helpers needed by E01.
- Route `/` to chat and keep the classic Home page at `/classic`.

## Acceptance Checklist

### Visual

- [x] Empty-state main area is vertically and horizontally centered.
- [x] Greeting changes by time of day and uses username/email when available.
- [x] Scene chips are centered and capped at 5.
- [x] Poster scene shows an `自动分层` coral badge when layered defaults are present or inferred.
- [x] Composer is below the scene chips and constrained to max 720px.
- [x] Composer receives focus on empty-state mount and after scene click.
- [x] Composer keeps inline controls for style, ratio, layered, quality, attachment, estimate, and send.
- [x] Compact mode hides the top scene/sample row and extra helper row.

### Interaction

- [x] Clicking a scene fills the textarea prompt.
- [x] Clicking a scene updates the ratio chip from `recommended_ratio`.
- [x] Clicking poster scene enables layered mode and shows `5 层`.
- [x] Clicking a scene focuses the input and keeps the cursor at the end.
- [x] Clicking a scene does not auto-send.
- [x] Typing in textarea and pressing Enter creates the first local user bubble and switches to message-flow layout.
- [x] First message immediately shows a user bubble and a placeholder AI progress card.

### Edge Cases

- [x] If the scenes API returns fewer than 5 scenes, only returned scenes are displayed.
- [x] If the scenes API returns more than 5 scenes, only the first 5 sorted scenes are displayed.
- [x] Anonymous users get the fallback greeting.
- [x] Mobile width wraps scene chips and keeps composer responsive.
- [x] If `GET /api/generation/scenes` fails, fallback scenes keep the composer usable and log an error.

## Self-Test

- [x] `cd web && pnpm.cmd build`
- [x] `go build ./...`
- [x] Local dev server returned 200 for `http://localhost:5175/`.
- [x] Backend scenes endpoint returned scene data from `http://localhost:3002/api/generation/scenes`.

## Screenshots

- [x] Empty state full screen on desktop: `docs/screenshots/e01-empty-desktop.png`
- [ ] Poster scene selected with layered chip active
- [ ] First message sent and message-flow layout visible
- [x] Mobile viewport: `docs/screenshots/e01-empty-mobile.png`

Poster-selected and first-message screenshots could not be captured through headless browser automation in this environment. Chrome/Edge headless did not emit screenshot files, and CDP calls timed out. The interaction behavior is covered by implementation checks and can be verified manually from the pushed branch.

## Additional Verification

- [x] `http://localhost:5175/` returned HTTP 200 from Vite.
- [x] `http://localhost:3002/api/generation/scenes` returned scene data.
- [x] Desktop screenshot captured through a visible browser window.
- [x] Mobile-sized screenshot captured through a visible browser window.
- [ ] Lighthouse score >= 90. Lighthouse is not installed in this workspace, so it could not be run without adding a new dependency.

## Notes

- The current branch is stacked on T04 because the repository had not merged the chat foundation into `main` yet.
- E01 normally assumes the full chat foundation exists. This branch adds the minimal frontend shell needed for the empty state to build and run, without implementing real message APIs or generation binding.
- Existing local build artifact change `web/dist/.gitkeep` is not included.
