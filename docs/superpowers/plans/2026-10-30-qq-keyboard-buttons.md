# QQ Keyboard Buttons Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Replace finite-choice QQ Bot prompts with official QQ Keyboard buttons while preserving text commands and a plain-text fallback.

**Architecture:** Keep existing handler return types and business sessions. Add a small presentation layer that maps known prompts to keyboard payloads, decode `INTERACTION_CREATE` into the existing `qq.GroupMessage`, and acknowledge every interaction through QQ's interaction callback endpoint. The gateway subscribes to both group messages and interaction events.

**Tech Stack:** Go, net/http, encoding/json, Gorilla WebSocket, httptest.

## Global Constraints

- Never place secrets in source, tests, commits, or output.
- Keep the QQ Bot service independent from new-api.
- Preserve text commands and automatically fall back to plain text when keyboard sending is rejected.
- Use QQ custom keyboard limits: at most five rows and five buttons per row; callback actions use `type: 1`, permission type `2`.
- Run `go test -race ./...`, `go vet ./...`, and `go build ./...` before completion.

### Task 1: Add QQ keyboard payloads and client sending

**Files:**
- Modify: `internal/qq/client.go`
- Test: `internal/qq/client_test.go`

**Interfaces:**
- Produce exported `Keyboard`, `KeyboardRow`, `KeyboardButton`, `RenderData`, `ButtonAction`, and `ButtonPermission` types.
- Produce `(*Client).SendGroupKeyboardMessage(ctx, groupOpenID, content string, keyboard Keyboard, messageID string) error`.

- [ ] Write a failing httptest that asserts `msg_type: 2`, markdown content, keyboard rows/buttons, and `msg_id` are sent.
- [ ] Run `go test ./internal/qq -run Keyboard -count=1` and observe failure because the method/types do not exist.
- [ ] Implement the payload structs and the method, reusing token acquisition and HTTP error handling from `SendGroupMessage`.
- [ ] Add a helper to construct callback buttons with unique ids, label/data, style `1`, action type `1`, permission type `2`, and compatibility text.
- [ ] Run the focused test and commit `feat: add QQ keyboard message client`.

### Task 2: Decode button interactions and acknowledge them

**Files:**
- Modify: `internal/qq/event.go`, `internal/qq/client.go`
- Test: `internal/qq/event_test.go`, `internal/qq/client_test.go`

**Interfaces:**
- Produce `DecodeInteraction(data []byte) (Interaction, error)` with event id, group/member ids, and button data.
- Produce `(*Client).AcknowledgeInteraction(ctx, interactionID string) error` using `PUT /interactions/{id}` and `{"code":0}`.

- [ ] Add a failing decoder test for the official `data.resolved.button_data`, `group_openid`, `group_member_openid`, and top-level `id` shape, plus rejection of missing identity/data.
- [ ] Add a failing client test asserting the PUT path, authorization, and code body.
- [ ] Run the focused tests to verify they fail for missing symbols.
- [ ] Implement strict decoding and acknowledgement with escaped interaction ids and existing response error conventions.
- [ ] Run focused tests and commit `feat: handle QQ keyboard interactions`.

### Task 3: Build keyboards for every finite-choice prompt

**Files:**
- Modify: `internal/handler/menu.go`, `internal/handler/site_router.go`, `internal/handler/binding.go`, `internal/handler/ai.go`
- Create: `internal/handler/keyboard.go`
- Test: `internal/handler/menu_test.go`, `internal/handler/site_router_test.go`, `internal/handler/binding_test.go`, `internal/handler/ai_test.go`

**Interfaces:**
- Produce `KeyboardForPrompt(content string) (qq.Keyboard, bool)` for main menu, site selection, yes/no replacement, AI group/model prompts, and the stop action.
- Keep all existing handler `Handle` signatures unchanged.

- [ ] Add failing tests for six main-menu buttons, FluxA/Charity site buttons, 是/否 buttons, dynamic AI group/model buttons, and a no-keyboard result for ordinary replies.
- [ ] Run focused handler tests and verify failure.
- [ ] Implement prompt parsing in `keyboard.go`; use deterministic button ids and preserve the current textual prompt as message content.
- [ ] Add small prompt markers or exported formatter helpers only where existing text cannot be parsed safely; do not alter business semantics.
- [ ] Run handler tests and commit `feat: render finite choices as QQ buttons`.

### Task 4: Wire keyboard send, interaction dispatch, fallback, and intents

**Files:**
- Modify: `cmd/bot/main.go`, `internal/handler/menu.go`
- Test: `cmd/bot/main_test.go` if needed, `internal/qq/gateway_test.go`

**Interfaces:**
- Route both text events and `INTERACTION_CREATE` through one dispatch function.
- Use intents `(1 << 25) | (1 << 26)`.

- [ ] Add a failing integration-level test or focused helper test proving keyboard prompts call the keyboard sender and unsupported keyboard errors retry as plain text.
- [ ] Run the test and observe failure.
- [ ] Add an internal sender interface/type assertion for keyboard-capable clients; send `SendGroupKeyboardMessage` when `KeyboardForPrompt` succeeds, otherwise use `SendGroupMessage`.
- [ ] For interactions, acknowledge first, convert `button_data` to `qq.GroupMessage`, then route through the same site router and handlers; use interaction id as the reply event reference when sending the result.
- [ ] Ensure menu and site prompts use buttons, and binding/AI finite-choice prompts use dynamic buttons; free-form email/code/chat messages remain text-only.
- [ ] Run all tests and commit `feat: wire QQ keyboard interactions`.

### Task 5: Verify and publish

**Files:**
- Modify: `README.md` only if configuration or platform setup must be documented.

- [ ] Run `gofmt -w` on modified Go files.
- [ ] Run `go test -race ./...`.
- [ ] Run `go vet ./...`.
- [ ] Run `go build ./...`.
- [ ] Push `release` so GitHub Actions publishes `ghcr.io/cychenhaibin/qq-bot:release`.
- [ ] In `/root/FluxA/qq-bot`, run `docker compose pull`, `docker compose up -d`, `docker compose ps`, and inspect the last 100 log lines without printing secrets.
