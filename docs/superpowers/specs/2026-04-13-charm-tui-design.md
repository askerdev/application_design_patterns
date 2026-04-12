# Charm TUI Rewrite Design

**Date:** 2026-04-13  
**Status:** Approved  
**Scope:** Replace `gocui`-based TUI with idiomatic Bubble Tea + Lip Gloss + Bubbles + Huh

---

## Context

Current TUI uses `github.com/jroimartin/gocui` with an imperative callback model: views are
manually created/destroyed, keybindings are registered/deleted per-section, and multi-step
forms are implemented as chained `showPrompt` callbacks. This makes the code hard to follow
and does not compose well.

Charm's ecosystem (Bubble Tea + Lip Gloss + Bubbles + Huh) offers an Elm-style reactive
architecture that is idiomatic, composable, and well-supported.

---

## Architecture

### Root Model

Single `tea.Program` with a root `model` that owns a `screen` enum and one sub-model per
section. The root `Update` delegates to the active sub-model; the root `View` renders the
active sub-model's view. `tea.WindowSizeMsg` is handled at root and propagated to all
sub-models so they can adapt to terminal size.

```
model (root)
├── screen  enum: home | tasks | projects | notes | reminders | tags | pomodoro | stats
├── homeModel
├── tasksModel
├── projectsModel
├── notesModel
├── remindersModel
├── tagsModel
├── pomodoroModel
└── statsModel
```

All sub-models hold references to their required repositories via a shared `*repos` struct
passed at construction time. The current `*domain.User` is also passed at init and threaded
through.

---

## Navigation

- **Home screen** — full-screen `list.Model` of 7 menu items (Tasks, Projects, Notes,
  Reminders, Tags, Pomodoro, Stats)
- `enter` on home → transition to section screen
- `esc` / `backspace` from any section → return to home
- `q` from home → quit
- `ctrl+c` → global quit (handled at root)

---

## Sections

### Tasks, Notes, Reminders, Tags

Each section:
- Primary view: `list.Model` (from `github.com/charmbracelet/bubbles/list`)
- Each domain item implements `list.Item` interface (`Title() string`, `Description() string`,
  `FilterValue() string`)
- Actions on **selected item** — no more ID prompts:
  - `a` — open add form
  - `d` — delete selected item (with confirmation via status line)
  - `c` (tasks only) — mark selected task complete
- Add/edit forms use `github.com/charmbracelet/huh` with typed fields:
  - Tasks: text (content) + select (priority H/M/L) + select (project, optional) + select (tag, optional)
  - Notes: text (content) + select (project, optional)
  - Reminders: text (content) + text (datetime) + select (project, optional) + select (tag, optional)
  - Tags: text (name)
- Form renders full-screen; `esc` cancels, `enter` on last field submits

### Projects

Split layout: left `list.Model` (project list, ~35% width) + right `viewport.Model` (project
detail: name, description, associated tasks). Layout composed with `lipgloss.JoinHorizontal`.
Navigating the list updates the viewport in real time. `a` adds project, `d` deletes selected.

### Pomodoro

- Two modes: **history** (list of past sessions) and **active** (timer running)
- `a` from history mode → `huh.Form` for duration + project → start session
- Active mode displays:
  - State label (RUNNING / PAUSED)
  - `MM:SS` countdown
  - `progress.Model` bar showing time elapsed
  - Help line: `p` pause · `r` resume · `c` complete · `esc` cancel
- Timer driven by `tea.Tick(time.Second, tickMsg{})` — idiomatic, no goroutine
- On completion/cancel → return to history list, persist session via repo

### Stats

Read-only `viewport.Model`. Renders task counts (total/done/todo/in-progress) and per-project
ETA using the existing `taskmath.CalculateETA` math model. Scrollable with `j`/`k`.

---

## Forms (huh)

All multi-field input uses `github.com/charmbracelet/huh`. Project and tag pickers become
`huh.Select` fields populated from repos — eliminates raw ID entry. Forms are embedded as a
`*huh.Form` field in the section model. When a form is active, the section model's `Update`
delegates to the form; `View` renders the form full-screen. On completion the model reads
field values, writes to repo, refreshes list.

---

## Styling (lipgloss)

Single `styles.go` defines all `lipgloss.Style` variables:

- App color palette (primary, secondary, muted, error)
- `titleStyle` — section titles / list titles
- `statusBarStyle` — bottom help bar
- `selectedItemStyle` / `normalItemStyle` — list item renderer
- `activePane` / `inactivePane` — border styles for split view (projects)
- `timerStyle`, `progressStyle` — pomodoro display

`list.DefaultDelegate` is replaced with a custom delegate using the above styles.

---

## File Layout

```
internal/tui/
├── app.go          root model, Init/Update/View, screen routing, WindowSizeMsg
├── styles.go       all lipgloss styles and palette
├── home.go         homeModel (list.Model menu)
├── tasks.go        tasksModel (list + huh form)
├── projects.go     projectsModel (list + viewport split)
├── notes.go        notesModel (list + huh form)
├── reminders.go    remindersModel (list + huh form)
├── tags.go         tagsModel (list + huh form)
├── pomodoro.go     pomodoroModel (progress + tea.Tick)
└── stats.go        statsModel (viewport)
```

`input.go` and `pickers.go` are deleted — superseded by `huh`.

---

## Dependencies

**Added:**
```
github.com/charmbracelet/bubbletea
github.com/charmbracelet/bubbles
github.com/charmbracelet/lipgloss
github.com/charmbracelet/huh
```

**Removed:**
```
github.com/jroimartin/gocui
github.com/nsf/termbox-go   (gocui transitive dep)
```

All other dependencies (sqlite, domain, repository, patterns, math, telegram) remain unchanged.

---

## What Does NOT Change

- All domain types (`internal/domain/`)
- All repository implementations (`internal/repository/`)
- The state machine (`internal/patterns/state/`)
- The ETA math model (`internal/math/`)
- The Telegram notifier (`internal/telegram/`)
- `cmd/taskflow/main.go` entry point (only `tui.New(...)` call changes)

---

## Error Handling

Errors from repo calls surface in a status bar line at the bottom of each section view.
The status bar is a `lipgloss`-styled string in each sub-model, cleared on next action.

---

## Testing

- Domain, repository, pattern, and math packages: existing tests unchanged
- TUI: no unit tests for the view layer (terminal rendering is not unit-testable without
  significant harness); test coverage lives in the layers below
