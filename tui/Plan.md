## Goal
Build FlowTask — a personal, offline-first GTD-inspired task manager.

## Phases
1. ~~Phase 1: TUI (Desktop)~~ ✅ Done
2. ~~Phase 2: Android (Mobile)~~ ✅ Done
3. Phase 3: Sync — Supabase REST API — **Current**

## Current Step
Step 3: Go TUI sync client — direct Supabase REST API sync.

### Architecture (Revised)
```
┌──────────┐         ┌──────────────┐
│ Go TUI   │──REST──▶│              │
│ (SQLite) │◀──REST──│   Supabase   │
└──────────┘         │   (Postgres) │
                     │              │
┌──────────┐         │              │
│ Android  │──REST──▶│              │
│ (SQLite) │◀──REST──│              │
└──────────┘         └──────────────┘
```

### Key Decisions
- **No PowerSync** — no Go SDK available. Both clients use Supabase REST API directly.
- **No Supabase Auth** — single-user personal project.
- **Poll-based sync** — periodic pull from Supabase filtered by `updated_at`.
- **Last-write-wins** — conflict resolution via `updated_at` timestamp comparison.
- **Offline-first** — writes go to local SQLite first, queued for Supabase sync on reconnect.

### Supabase REST API
- URL: https://ykksgiyweklxbrfoomwa.supabase.co
- Key: sb_publishable_8JsM8svXjt1-yagX9M0n_w_jN7OWZBZ
- Tables confirmed: tasks, notes, inbox ✅

### Step 3: Go TUI Sync Client — **In Progress**
- Add `updated_at` column migration (003)
- Create Supabase HTTP client in Go
- On startup: full pull from Supabase, merge into local SQLite
- On local CRUD: queue sync operation to Supabase
- Periodic background sync (every 5-10s)
- Sync status indicator in status bar

### Step 4: Android Sync Client
- Add HTTP client (Ktor or OkHttp)
- Same sync logic as Go TUI
- Sync status indicator

### Step 5: Testing
- Test: create on TUI → appears on Android
- Test: create on Android → appears on TUI
- Test: offline editing on both platforms
- Test: conflict resolution (edit same task on both)
