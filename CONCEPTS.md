# Tusk Domain Concepts & Architecture Glossary

This document serves as the canonical domain taxonomy and technical vocabulary for the Tusk project.

---

## 1. Domain Entities & Hierarchies

### Task
The fundamental unit of work within Tusk. A task possesses a title, optional markdown description, state, priority, tags, due date, timestamps, and an optional reference to a parent task.

### Root Task
A top-level task that has no parent (`parent_id == nil`). Root tasks represent major workstreams, epics, or standalone tasks.

### Subtask
A task that references another task as its parent (`parent_id != nil`). Subtasks can themselves act as parents. The current core supports up to 10 levels, counting a root as depth 1.

### Task Tree
The acyclic directed tree structure formed by a root task and all of its recursive child subtasks. Cycles (e.g., a task acting as an ancestor of itself) are strictly forbidden and rejected at the domain validation boundary.

### Rollup Progress
An integer percentage ($0\% - 100\%$) representing the completion status of a task:
- For a leaf task (no subtasks): $100\%$ if `done`, otherwise explicitly assigned manual progress ($0\%$–$99\%$, default $0\%$ on creation).
- For a parent task: the mathematical average of its immediate child subtasks' progress values:
  $$\text{Progress} = \left\lfloor \frac{\sum_{i=1}^{N} \text{subtask}_i.\text{Progress}}{N} \right\rfloor$$

### Display Order
The deterministic sort order applied when displaying tasks in the CLI or TUI:
1. Priority (Urgent > High > Medium > Low)
2. Due Date (Earliest to Latest, missing dates last)
3. Creation Date (Chronological)
4. ID (Ascending deterministic tie-breaker)

Apply the same order to siblings in a tree. Pinned/active ordering is reserved terminology; the current product contract has no persisted pin capability.

### Task Event
A timeline entry recording the kind, changed field names, sequence, and time of a task mutation. Task events do not store old notes or reconstruct task state. Deleting a task deletes its events.

---

## 2. Terminal UI (TUI) Concepts

### Elm Architecture
The unidirectional data flow pattern governing the Bubble Tea interface:
- **Model**: Application state owned by the UI message loop. Commands receive copied inputs rather than accessing the model asynchronously.
- **Update**: Handles incoming messages (`tea.Msg`), changes model state and prepares the next frame, returning the model and asynchronous commands (`tea.Cmd`).
- **View**: Returns the prepared frame without changing state, querying time, performing I/O or invoking child components.

### Active Panel
The currently focused area in the two-panel TUI:
- **TaskList Panel**: The left panel displaying the hierarchical tree of tasks.
- **TaskDetails Panel**: The right panel displaying metadata, notes, and progress of the selected task.
- **Modal View**: A centered modal overlay intercepting input for forms (creation, editing, deletion confirmation).

### Keymap Scope
The context-aware mapping of physical keystrokes to domain actions:
- **Global Keymap**: `Ctrl+C` cancels and exits from every context.
- **Browse Keymap**: `?` opens help and `Tab` switches list/details focus. In the task list, `j/k` moves selection; `x` completes a subtree or reopens the selected task; `a` creates a root task.
- **Form Keymap**: `Tab` moves between fields and buttons. `Enter` advances single-line fields, inserts a newline in notes or activates a focused button. `Ctrl+S` saves; `Esc` cancels, asking before discarding a dirty draft. Text such as `q`, `d` and `?` never invokes browse actions.

---

## 3. Storage & Infrastructure Concepts

### Storage Engine
The persistence boundary that saves and retrieves tasks and their task events independently of business orchestration. A write groups related changes so they succeed together; an unconfirmed outcome requires reading stored state before deciding what to repeat.

### Zero-Friction Startup
The design principle ensuring Tusk runs immediately upon binary execution without requiring external database services, background daemon processes, environment configurations, or interactive authentication screens.
