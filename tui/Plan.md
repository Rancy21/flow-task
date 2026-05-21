## Goal
Build FlowTask Phase 1 — a working Go TUI task manager with Bubble Tea.

## Phases
1. ~~Back-end infrastructure~~ ✅ Done
2. UI screens (Today, Week, Index, Notes, Editor) — **Current**
3. Entrypoint & integration testing

## Architecture Recap
- Bubble Tea + Lip Gloss + Bubbles
- SQLite via `modernc.org/sqlite` (WAL mode, FK ON)
- Models → Repos → UI (no DB in UI layer)
- Root App owns tab routing; children emit messages

## Current Step
Inbox refactor complete — new InboxItem model, inbox table, inbox editor.
Testing the Inbox capture flow.
