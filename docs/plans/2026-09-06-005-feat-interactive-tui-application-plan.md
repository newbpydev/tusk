---
artifact_contract: ce-unified-plan/v1
artifact_readiness: implementation-ready
product_contract_source: 2026-09-06-001-feat-tusk-modern-task-system-plan.md
---

# Feature Plan 005: Interactive TUI Application

## Goal Capsule

Implement the keyboard-driven Terminal User Interface (TUI) in `internal/tui/` using Charm's Bubble Tea, Lipgloss, and Bubbles. Deliver a multi-panel productivity interface with zero layout shifts, zero render-phase state mutations, vim navigation, live search, and modal creation/edit forms.

---

## Technical Design & Scope

### 1. Pure Elm Architecture (`internal/tui/`)
- Root `Model`:
  ```go
  type Model struct {
      taskList    TaskListModel
      taskDetails TaskDetailsModel
      formModal   FormModalModel
      helpBar     HelpBarModel
      activePanel PanelType
      width       int
      height      int
      service     ports.TaskService
  }
  ```
- **Inviolable Invariant**: `View() string` is 100% read-only and idempotent. No pointers are mutated, no caches populated, and no child styles initialized during `View()`.
- State transitions, cursor movements, and modal toggles occur exclusively within `Update(tea.Msg)`.

### 2. UI Panels & Layout
- **Task Tree / List Panel**: Displays tasks with selection cursor (`>`), status icons, priority colors, and subtask nesting indents.
- **Detail Viewport**: Displays full description, formatted due dates, tag badges, and subtask progress bar.
- **Help Bar**: Dynamic bottom bar displaying available keybindings for the active panel.
- **Modal Overlay**: Centered modal for task creation, editing, and delete confirmations using `lipgloss.Place`.

### 3. Keymap & Interaction
- `j`/`k`, `Down`/`Up`: Move cursor.
- `h`/`l`, `Left`/`Right`: Collapse/expand subtasks or switch panels.
- `Space` / `x`: Toggle task completion.
- `a`: Trigger create task modal.
- `e`: Trigger edit task modal.
- `d`: Delete task with confirmation dialog.
- `/`: Activate search filter.
- `?`: Toggle full help screen.
- `q`: Quit application.

---

## Verification Scenarios

1. Synthetic message pump tests: send sequences of `tea.KeyMsg` (`j`, `k`, `x`, `a`) to `Update()` and assert correct model state changes.
2. Window resize tests: send `tea.WindowSizeMsg` across 80x24, 120x40, and 200x60 dimensions and assert panels fit without terminal overflow or line wrapping breaks.
3. Pure `View()` test: call `View()` 100 times in a loop and assert returned string is identical and model hash is unchanged.
4. Modal overlay tests: assert modal renders centered and traps keyboard focus while active.
