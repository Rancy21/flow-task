## Phase 1: Core infrastructure (COMPLETE ✅)

### Tasks
- [x] Go module init (`github.com/Rancy21/flowtask`)
- [x] Data models: Task, Note, Priority, TaskStatus
- [x] DB layer: SQLite open, WAL/FK pragmas, migration runner, schema
- [x] `001_init.sql` migration (tasks + notes tables)
- [x] `TaskRepo`: GetToday, GetWeek, GetInbox, GetByID, Create, Update, MarkDone, Schedule, Delete
- [x] `NoteRepo`: GetAll, GetByTask, Create, Delete
- [x] `styles.go`: Palette, priority badges, layout styles, status bar, empty state, boxes
- [x] `app.go`: Root model, tab routing (1-4 / tab), key bindings, header/status bar

---

## Phase 2: UI Screens (COMPLETE ✅)

### Tasks
- [x] Create `internal/ui/today.go` — TodayModel
- [x] Create `internal/ui/week.go` — WeekModel
- [x] Create `internal/ui/index.go` — IndexModel (later refactored to Inbox)
- [x] Create `internal/ui/notes.go` — NotesModel
- [x] Create `internal/ui/editor.go` — EditorModel, EditorDoneMsg, EditorCancelMsg
- [x] Create `cmd/main.go` — entrypoint wiring DB + repos + Bubble Tea
- [x] Compile and fix issues
- [x] Run and manual test

---

## Phase 3: Inbox Refactor (COMPLETE ✅)

### Why
The Index tab was originally using the Task model with INBOX status. The user wanted:
- Inbox items are NOT tasks — lightweight capture only (title + description)
- Separate `inbox` table, `InboxItem` model, `InboxRepo`
- Dedicated inbox editor (no priority, date, notes)
- No auto-promote — user manually creates tasks from inbox content

### Changes
- [x] Add `InboxItem` model (`internal/model/inbox.go`)
- [x] Add `InboxRepo` (`internal/repository/inbox_repo.go`)
- [x] Add `002_inbox.sql` migration + embed in `db.go`
- [x] Rewrite `index.go` → InboxModel using InboxItem + InboxRepo
- [x] Create `inbox_editor.go` — simple title+description modal
- [x] Update `app.go`: TabInbox, inbox editor routing, OpenInboxEditorMsg, InboxEditorDoneMsg/CancelMsg
- [x] Update `cmd/main.go`: wire InboxRepo
- [x] Update `AGENTS.md`: InboxItem model, Inbox section, Inbox Editor section, conventions
- [x] Compile

---

## Phase 4: Polish & Git (NEXT)

### Tasks
- [ ] Test inbox capture/edit/delete flow
- [ ] `go mod tidy` 
- [ ] `git init` + initial commit
- [ ] Add `.gitignore` (verify contents)
