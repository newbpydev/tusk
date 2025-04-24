# Tusk: Your Tasks, Tamed with Go

A terminal-based task management application written in Go, designed for developers and power users who want a fast, extensible, and feature-rich TODO experience. Tusk combines PostgreSQL persistence, concurrent file backups, and a rich terminal UI to keep your workflow smooth, both now in CLI form and future REST/API extensions.

---

## 🗺️ Roadmap & Project Timeline

This section outlines what’s been completed, what’s in progress, and upcoming milestones.

### ✅ Completed

- **Scaffolding & Configuration** _(2025-04-23)_

  - Project structure with `cmd/`, `internal/`, `db/`, `migrations/`, and config.
  - GitHub repo initialized, `.env` support via `joho/godotenv`.
  - `.gitignore`, `.env.example`, and base README created.

- **Database Integration** _(2025-04-24)_

  - PostgreSQL schema defined with `users` and `tasks` tables, supporting recursive subtasks.
  - Migrations implemented using `golang-migrate` with triggers and constraints.
  - Type-safe DB layer generated via `sqlc`, using `pgx` and `pgtype`.
  - Smoke tests in `cmd/cli/main.go` for `CreateUser` and `CreateTask`.

### ▶️ In Progress

- **Step 3: Domain Models & Repository Layer**

  - Define Go structs for `User` and `Task`, including subtasks hierarchy and progress logic.
  - Implement repository wrappers that call the generated `sqlc` methods.
  - Auto-calculation of parent task completion based on subtasks.

### ⏳ Upcoming

- **Step 4: Service Layer & CLI Handlers**

  - Build service interfaces and business logic.
  - Wire CLI commands for `add`, `list`, `complete`, `delete`, and `reorder`.

- **Step 5: Terminal UI (TUI)**

  - Choose between `tview` or `tcell` for interactive Kanban and List views.
  - Keyboard navigation, highlighting, and toggling tasks.

- **Step 6: Kanban View & Advanced Features**

  - Group tasks by status in columns, enable drag-and-drop or key-based moves.
  - Persist column order and support custom sorting.

- **Step 7: User Authentication & Management**

  - Secure login prompt, session handling.
  - CLI commands for adding and removing users.

- **Step 8: File Backup & Concurrency**

  - Implement JSON/CSV backup that runs in parallel with DB writes via goroutines and channels.
  - Error handling and retry mechanisms.

- **Step 9: Testing & CI**

  - Unit tests for repository and service layers.
  - Integration tests with a test database.
  - GitHub Actions for automated testing and linting.

- **Step 10: REST API & Web App**

  - Expose the same business logic via a Go HTTP server.
  - Build a minimal web frontend (e.g. Next.js, SvelteKit) to consume the API.

---

## 🛠️ Getting Started

### Prerequisites

- Go 1.21+
- PostgreSQL 12+
- `sqlc` v1.XX (for type-safe queries)
- `golang-migrate` v4

### Setup

1. **Clone the repository**

   ```bash
   git clone https://github.com/YOUR_USERNAME/tusk.git
   cd tusk
   ```

2. **Configure environment**

   ```bash
   cp .env.example .env
   # Edit .env with your DATABASE_URL
   ```

3. **Run migrations**

   ```bash
   migrate -path db/migrations -database "$DATABASE_URL" up
   ```

4. **Generate DB code**

   ```bash
   sqlc generate
   ```

5. **Run the CLI smoke test**

   ```bash
   go run cmd/cli/main.go
   ```

---

## 🤝 Contributing

We welcome contributions! Please follow these guidelines:

1. Fork the repo and create a feature branch (`git checkout -b feature/foo`).
2. Run tests and ensure `sqlc generate` passes.
3. Follow the existing code style and include comments.
4. Open a Pull Request with a clear description of your changes.

---

## 📜 License

This project is licensed under the GNU Affero General Public License v3.0 or later. See the [LICENSE](LICENSE) file for details.

License header for source files:

```go
// Copyright (C) 2025 Juan Antonio Gomez Pena
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Affero General Public License for more details.
//
// You should have received a copy of the GNU Affero General Public License
// along with this program. If not, see <https://www.gnu.org/licenses/>.
```
