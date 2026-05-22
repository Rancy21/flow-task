## Phase 5: Android App (COMPLETE ✅)

- [x] Project scaffold (Gradle 8.11.1, AGP 8.7.2, Kotlin 2.1.0)
- [x] Room entities, DAOs, database, repositories
- [x] Koin DI, theme, 4-tab navigation
- [x] All screens: Today, Week, Inbox, Notes
- [x] Task Editor with notes, Inbox Editor
- [x] All bugfixes (FAB, stale state, loading bar, keyboard, XML, icons, .gitignore)
- [x] Commit & push (54 files)

---

## Phase 6: Sync (Supabase REST API)

### Why
- Local-only is fine for Phase 1-2, but user needs TUI ↔ Android sync
- PowerSync rejected: no Go SDK. Supabase REST API is simpler and works for both.
- Same Supabase project, same tables, direct REST API calls
- Last-write-wins conflict resolution via `updated_at` column

### Plan
1. Add `updated_at` column to all local SQLite + Supabase tables (done on Supabase) ✅
2. Go TUI: Supabase HTTP client + sync logic
3. Android: HTTP client + sync logic
4. Testing: cross-platform CRUD + offline + conflicts

### Step 1: Supabase Setup (COMPLETE ✅)
- [x] Create Supabase project
- [x] Create Postgres tables (tasks, notes, inbox) with `updated_at`
- [x] Verify REST API accessibility with publishable key
- [x] Push migrations via `supabase db push`

### Step 2: Go TUI Sync Client — Pull & Push (COMPLETE ✅)
- [x] Add `updated_at` column migration (003_add_updated_at.sql)
- [x] Update Task, Note, InboxItem models with UpdatedAt field
- [x] Update all repositories: SELECT queries, INSERT, UPDATE with updated_at
- [x] Add Upsert() methods to all repos for sync pulls
- [x] Create `internal/sync/` package with Supabase HTTP client
- [x] Pull: GET /rest/v1/tasks, /notes, /inbox → upsert locally
- [x] Push: POST/PATCH tasks, notes, inbox items to Supabase
- [x] Delete: DELETE from Supabase on local delete
- [x] Wire sync into main.go: startup pull + 30s background poll
- [x] Wire sync push into task editor + inbox editor
- [x] Fix: notes not pushing to Supabase after creation
- [x] Verified: pull, push, delete work against live Supabase

### Step 3: Android Sync Client (COMPLETE ✅)
- [x] Add OkHttp dependency
- [x] Add `updatedAt` to all Room entities (TaskEntity, NoteEntity, InboxEntity)
- [x] Room migration 1→2: add `updated_at` columns
- [x] Create `SyncClient` with pull/push/delete via Supabase REST API
- [x] Add upsert methods to all repositories (check-then-insert-or-update)
- [x] Wire sync push into TaskEditorViewModel (task + notes)
- [x] Wire sync push into InboxEditorViewModel
- [x] Startup sync pull + 30s background sync in FlowTaskApp
- [x] Fix: add INTERNET permission to AndroidManifest
- [x] Build verified ✅
- [x] Tested: pull, push working on device

### Step 4: Testing
- [ ] Test: create task on TUI → appears on Android
- [ ] Test: create task on Android → appears on TUI
- [ ] Test: edit task on TUI → updated on Android
- [ ] Test: delete task on Android → removed from TUI
- [ ] Test: offline edit on Android → syncs when back online
- [ ] Test: conflict (edit same task on both while offline) → last-write-wins
- [ ] Test: notes sync correctly
- [ ] Test: inbox items sync correctly
