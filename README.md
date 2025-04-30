# Tusk: Your Tasks, Tamed with Go

A terminal-based task management application written in Go, designed for developers and power users who want a fast, extensible, and feature-rich TODO experience. Tusk combines PostgreSQL persistence, concurrent file backups, and a rich terminal UI to keep your workflow smooth, both now in CLI form and future REST/API extensions.

## 🧰 Technology Stack

- **Backend**: Go 1.24
- **Database**: PostgreSQL 12+
- **SQL Tools**:
  - `sqlc` - Type-safe SQL query generation
  - `golang-migrate` - Database migration management
  - `pgx` - PostgreSQL driver for Go
- **CLI Framework**: Cobra - Rich command-line interface with subcommands
- **Terminal UI**:
  - `bubbletea` - MVC framework for terminal applications
  - `lipgloss` - Styling for terminal UI components
- **Configuration**: `godotenv` - Environment variable management
- **Authentication**: `golang.org/x/crypto` - Password hashing and verification
- **Version Control**: Git

## 📂 Project Structure

```plaintext
tusk-go/
├── cmd/                    # Entry points for different app interfaces
│   ├── api/                # REST API server entry point (future)
│   └── cli/                # Command-line interface entry point
├── configs/                # Configuration files and templates
├── db/                     # Database definitions
│   ├── migrations/         # SQL migration files
│   └── queries.sql         # SQL queries for sqlc generation
├── internal/               # Internal application code (not exported)
│   ├── adapters/           # External system adapters (hexagonal architecture)
│   │   ├── auth/           # Authentication adapter
│   │   ├── backup/         # File backup adapter
│   │   ├── db/             # Database adapter and repository implementations
│   │   │   └── sqlc/       # Generated SQL code
│   │   └── tui/            # Terminal UI adapter
│   │       └── bubbletea/  # BubbleTea implementation of the TUI
│   ├── app/                # Application services orchestration
│   ├── cli/                # CLI command implementations
│   ├── config/             # Configuration loading and parsing
│   ├── core/               # Core domain models and business logic
│   │   ├── errors/         # Domain-specific errors
│   │   ├── task/           # Task domain model
│   │   └── user/           # User domain model
│   ├── ports/              # Interface definitions (hexagonal architecture)
│   │   ├── input/          # Input ports (use cases)
│   │   └── output/         # Output ports (repositories)
│   ├── service/            # Service layer implementation
│   │   ├── task/           # Task service implementation
│   │   └── user/           # User service implementation
│   └── util/               # Shared utilities
├── migrations/             # Application-level migrations
├── test/                   # Integration and E2E tests
├── go.mod                  # Go module definition
├── go.sum                  # Go module checksums
├── LICENSE                 # AGPL-3.0 license file
├── README.md               # This file
└── sqlc.yaml               # sqlc configuration
```

The project follows a hexagonal architecture (also known as ports and adapters) to maintain a clean separation of concerns:

- **Core Domain**: Contains pure business logic and domain models
- **Ports**: Define interfaces for the application to interact with external systems
- **Adapters**: Connect the application to external systems like databases and UI
- **Services**: Implement use cases by orchestrating domain models and ports

This architecture allows us to easily:

- Switch between different database providers
- Add new user interfaces (CLI, TUI, Web) without changing business logic
- Test business logic in isolation from external dependencies

---

## 🗺️ Roadmap & Project Timeline

This section outlines what's been completed, what's in progress, and upcoming milestones.

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

- **Domain Models & Repository Layer** _(2025-04-25)_

  - Define Go structs for `User` and `Task`, including subtasks hierarchy and progress logic.
  - Implement repository wrappers that call the generated `sqlc` methods.
  - Auto-calculation of parent task completion based on subtasks.

- **Service Layer & CLI Handlers** _(2025-04-25)_

  - Built service interfaces and business logic for tasks and users.
  - Implemented CLI commands for `add`, `list`, `complete`, `delete`, and `reorder`.
  - Added Cobra command structure for intuitive CLI experience.

- **Initial Terminal UI (TUI)** _(2025-04-26)_

  - Implemented basic `bubbletea` framework with list view of tasks.
  - Added keyboard navigation and task status toggling.
  - Created task deletion functionality with confirmation prompt.
  - Designed basic styling with `lipgloss` for improved visual hierarchy.

- **Advanced Terminal UI Features** _(2025-04-28)_

  - Implemented detail view for tasks with scrollable content.
  - Added task creation form with field validation.
  - Implemented time-based task categorization (overdue, today, upcoming).
  - Added status messages and loading indicators for better user feedback.
  - Enhanced UI with priority color-coding and status indicators.

### ▶️ In Progress

- **TUI Enhancements**

  - Adding subtask management within the TUI.
  - Working on task filtering and sorting options.
  - Developing comprehensive keyboard shortcuts and help menu.
  - Implementing edit mode for existing tasks.

### ⏳ Upcoming

- **Kanban View & Advanced Features**

  - Group tasks by status in columns, enable drag-and-drop or key-based moves.
  - Persist column order and support custom sorting.
  - View task dependencies and relationships visually.

- **User Authentication & Management**

  - Secure login prompt, session handling.
  - CLI commands for adding and removing users.
  - Role-based permissions for team task management.

- **File Backup & Concurrency**

  - Implement JSON/CSV backup that runs in parallel with DB writes via goroutines and channels.
  - Error handling and retry mechanisms.
  - Automated scheduled backups with configurable intervals.

- **Testing & CI**

  - Unit tests for repository and service layers.
  - Integration tests with a test database.
  - GitHub Actions for automated testing and linting.
  - Code coverage reports and quality metrics.

- **REST API & Web App**

  - Expose the same business logic via a Go HTTP server.
  - Build a minimal web frontend (e.g. Next.js, SvelteKit) to consume the API.
  - Add JWT authentication for secure API access.

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

5. **Build and run the CLI**

   ```bash
   go build -o tusk ./cmd/cli
   ./tusk help
   ```

6. **Run the TUI interface**

   ```bash
   ./tusk tui
   ```

### TUI Key Commands

- **Navigation**

  - `↑`/`↓` or `k`/`j` - Navigate between tasks
  - `Page Up`/`Page Down` or `Ctrl+b`/`Ctrl+f` - Scroll by page
  - `Home`/`End` or `g`/`G` - Jump to top/bottom
  - `Tab` - Cycle between panels (left to right)
  - `Left`/`Right` or `h`/`l` - Navigate between panels
  - `1`/`2`/`3` - Toggle panel visibility

- **Task Management**

  - `Space` or `c` - Toggle task completion status
  - `Enter` - View task details
  - `n` - Create new task
  - `d` - Delete selected task
  - `e` - Edit selected task
  - `r` - Refresh task list

- **Application Controls**
  - `Esc` - Return to previous view
  - `q` - Quit the application

---

## 🤝 Contributing

We welcome contributions! Please follow these guidelines:

1. Fork the repo and create a feature branch (`git checkout -b feature/foo`).
2. Run tests and ensure `sqlc generate` passes.
3. Follow the existing code style and include comments.
4. Open a Pull Request with a clear description of your changes.

### Development Environment

For best results, we recommend:

- VS Code with Go extension
- PostgreSQL running locally or in Docker
- Go 1.24 or newer

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
