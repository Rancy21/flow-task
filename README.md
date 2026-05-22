<picture align="center">
  <source media="(prefers-color-scheme: dark)" srcset="https://readme-typing-svg.demolab.com?font=Fira+Code&weight=700&size=32&duration=3000&pause=500&color=3B82F6&center=true&vCenter=true&width=600&lines=FlowTask">
  <img src="https://readme-typing-svg.demolab.com?font=Fira+Code&weight=700&size=32&duration=3000&pause=500&color=3B82F6&center=true&vCenter=true&width=600&lines=FlowTask" alt="FlowTask">
</picture>

<p align="center">
  <strong>A personal, offline-first task manager built around <em>my</em> workflow.</strong><br>
  GTD-inspired. Opinionated. Mine.
</p>

<p align="center">
  <img src="https://img.shields.io/badge/Go-00ADD8?style=flat&logo=go&logoColor=white" alt="Go">
  <img src="https://img.shields.io/badge/Kotlin-7F52FF?style=flat&logo=kotlin&logoColor=white" alt="Kotlin">
  <img src="https://img.shields.io/badge/SQLite-003B57?style=flat&logo=sqlite&logoColor=white" alt="SQLite">
  <img src="https://img.shields.io/badge/Supabase-3ECF8E?style=flat&logo=supabase&logoColor=white" alt="Supabase">
  <img src="https://img.shields.io/badge/license-MIT-blue?style=flat" alt="License">
</p>

---

## Why FlowTask?

I built this for myself. No time-boxing, no collaboration features, no subscriptions — just a task manager that works the way I think.

- **Capture fast.** Every thought drops into the Inbox. Title + description. That's it.
- **Schedule later.** When I'm ready, I promote an inbox item to a real task with priority and a date.
- **Execute today.** The Today view shows only what I committed to doing right now.
- **Plan the week.** The Week view groups tasks by day so I can see what's ahead.

It runs everywhere I do: **terminal** on my laptop, **Android** on my phone. They stay in sync via Supabase.

---

## How It Works

```
┌──────────┐         ┌──────────────┐         ┌──────────┐
│  Go TUI  │──REST──▶│   Supabase   │◀──REST──│ Android  │
│ (SQLite) │◀────────│  (Postgres)  │────────▶│ (SQLite) │
└──────────┘         └──────────────┘         └──────────┘
```

Both apps own their own SQLite database. They push changes to Supabase and pull updates every 30 seconds. Offline edits sync automatically when you're back online. Conflicts resolve by timestamp (last write wins).

No auth. No user accounts. This is my data, on my devices, synced through my Supabase project.

---

## The Workflow

| Step | Where | What |
|------|-------|------|
| Capture | Inbox | Dump thoughts as lightweight items. Title + optional description. |
| Process | Editor | Manually promote an inbox item to a task. Add priority, schedule a date, attach reflection notes. |
| Execute | Today | See only tasks scheduled for today, sorted by priority (P1 → P3). Mark them done. |
| Review | Week | Scan the full week ahead, grouped by day. Reschedule as needed. |
| Reflect | Notes | Browse all reflection notes across every task. |

No automatic promotion. No AI. Just intentional manual processing.

---

## Tech Stack

### TUI (Desktop)
| Layer | Choice |
|-------|--------|
| Language | Go |
| TUI Framework | [Bubble Tea](https://github.com/charmbracelet/bubbletea) |
| Styling | [Lip Gloss](https://github.com/charmbracelet/lipgloss) |
| Database | SQLite via `modernc.org/sqlite` (pure Go, no CGo) |
| Sync | Supabase REST API |

### Android
| Layer | Choice |
|-------|--------|
| Language | Kotlin 2.1 |
| UI | Jetpack Compose + Material 3 |
| Database | Room + SQLite |
| DI | Koin |
| HTTP | OkHttp |
| Sync | Supabase REST API |

---

## Architecture

```
flowtask/
├── tui/                          # Go TUI (desktop)
│   ├── cmd/main.go               # Entrypoint
│   ├── internal/
│   │   ├── db/                   # SQLite, WAL, migrations
│   │   ├── model/                # Task, Note, InboxItem
│   │   ├── repository/           # TaskRepo, NoteRepo, InboxRepo
│   │   ├── sync/                 # Supabase REST client
│   │   └── ui/                   # Bubble Tea screens & editors
│   ├── migrations/               # 001_init, 002_inbox, 003_updated_at
│   └── go.mod
│
├── android/                      # Android app
│   └── app/src/main/java/com/rancy21/flowtask/
│       ├── data/
│       │   ├── entity/           # Room @Entity classes
│       │   ├── dao/              # Room @Dao interfaces
│       │   ├── database/         # FlowTaskDatabase
│       │   ├── repository/       # TaskRepo, NoteRepo, InboxRepo
│       │   └── sync/             # OkHttp sync client
│       └── ui/
│           ├── today/            # Today screen
│           ├── week/             # Week screen
│           ├── inbox/            # Inbox screen
│           ├── notes/            # Notes screen
│           ├── editor/           # Task editor
│           ├── inboxeditor/      # Inbox capture editor
│           ├── navigation/       # Tab routing
│           └── theme/            # Material 3 theme
│
└── supabase/
    └── migrations/               # Postgres schema (mirrors SQLite)
```

---

## Getting Started

### Prerequisites
- **Go** 1.21+
- **Android Studio** (for the Android app)
- A **Supabase** project (free tier)

### Run the TUI

```bash
cd tui
go run ./cmd/main.go
```

### Build Android

```bash
cd android
./gradlew installDebug
```

### Set Up Sync

1. Create a [Supabase](https://supabase.com) project
2. Run the migrations in `supabase/migrations/`
3. Copy your project URL and publishable key
4. Update `tui/internal/sync/sync.go` and `android/.../data/sync/SyncClient.kt` with your credentials

---

## Data Model

```
tasks
├── id (UUID)
├── title
├── description
├── priority (P1 | P2 | P3)
├── status (INBOX | SCHEDULED | DONE)
├── scheduled_date (YYYY-MM-DD, nullable)
├── created_at, completed_at, updated_at

notes
├── id (UUID)
├── task_id → tasks (CASCADE)
├── content
├── created_at, updated_at

inbox
├── id (UUID)
├── title
├── description
├── created_at, updated_at
```

All dates are ISO 8601 strings. No epoch millis. No timezones (dates only). Same schema across all three databases.

---

## License

MIT — do whatever you want. This is my personal tool; if it helps you, great.

---

<p align="center">
  <sub>Built by <a href="https://github.com/Rancy21">@Rancy21</a> for their own workflow.</sub>
</p>
