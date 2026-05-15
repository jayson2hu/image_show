# PR: T04 Conversation CRUD

## Summary

- Add authenticated Conversation CRUD APIs for Chat redesign T04.
- Add `Conversation` model migration and route registration.
- Add controller tests for list/create/detail/rename/delete, auth failure, ownership checks, search, cursor pagination, and soft delete.

## Scope

- `GET /api/conversations`
- `POST /api/conversations`
- `GET /api/conversations/:id`
- `PATCH /api/conversations/:id`
- `DELETE /api/conversations/:id`

## Acceptance Criteria

- [x] 5 APIs are available with expected statuses: 200 / 201 / 200 / 200 / 204.
- [x] Auth failure returns 401.
- [x] Cross-user access returns 404 without exposing existence.
- [x] List supports `?q=` fuzzy title search with `LIKE %xx%`.
- [x] List supports `?limit=20&cursor=<last_id>` cursor pagination.
- [x] Soft delete hides rows from list while preserving SQL row.
- [x] `controller/conversation_test.go` covers list, create, rename, delete, unauthorized 401, ownership 404, and missing 404.

## Self-Test Steps

- [x] `go build ./...`
- [x] `go test ./model/... -v`
- [x] `go test ./controller -run Conversation -v`
- [x] `go test ./...`
- [x] Manual API flow against local server on port 3001:
  - Login with default dev admin.
  - Create conversation: 201.
  - List conversations: returns created item.
  - Get conversation detail: 200.
  - Rename conversation: 200.
  - Delete conversation: 204.
  - Search after delete: deleted item hidden.

## Manual API Result

```text
created_status     : 201
created_id         : 1
list_count         : 1
detail_title       : T04 curl test
renamed_title      : T04 renamed
delete_status      : 204
after_delete_count : 0
```

## Notes

- T04 depends on T01. The current branch includes the minimal `Conversation` model and migration required for T04 because the local codebase did not yet contain T01 schema changes.
- This PR intentionally does not implement T05 messages, generation binding, share, export, or chat UI.
- Existing local build artifact change `web/dist/.gitkeep` is not included.
