# Adapters

This directory contains adapter implementations that connect the core application to external systems, following the hexagonal architecture pattern.

## Structure

```plaintext
adapters/
├── auth/              # Authentication adapter
├── backup/            # File backup adapter
├── db/                # Database adapter and repository implementations
│   ├── db.go          # Database connection management
│   ├── task_repo.go   # Task repository implementation
│   ├── user_repo.go   # User repository implementation
│   └── sqlc/          # Generated SQL code from sqlc
│       ├── db.go      # Database interface
│       ├── models.go  # SQL data models
│       └── queries.sql.go # Generated query methods
└── tui/               # Terminal UI adapter
    └── bubbletea/     # BubbleTea implementation of the TUI
        ├── app/       # Application container
        │   ├── app.go   # Application initialization and coordination
        │   ├── model.go # Core model definition and state management
        │   ├── update.go # Main update logic for handling messages
        │   ├── view.go  # Main view rendering
        │   ├── views.go # Utility functions for different views
        │   └── tasks.go # Task-specific functionality
        ├── components/# Reusable UI components
        ├── hooks/     # Custom hook functionality for UI
        ├── messages/  # Message definitions for Bubble Tea
        ├── styles/    # UI styling definitions
        └── views/     # Individual screens/views
```

## Purpose

Adapters serve as the bridge between the application's core logic and the outside world. They:

1. Implement the interfaces defined in the `ports` directory
2. Handle the technical details of interacting with external systems
3. Convert between domain models and external representations
4. Isolate the core application from dependency changes

## Key Adapters

### Database Adapter (`db/`)

The database adapter provides:

- Connection management with PostgreSQL using `pgx`
- Repository implementations for tasks and users
- Type-safe SQL queries using generated code from `sqlc`
- Transaction management

### Terminal UI Adapter (`tui/`)

The Terminal UI adapter using BubbleTea:

- Implements a modern terminal user interface
- Handles user interactions and keyboard shortcuts
- Provides views for task management, details, and creation
- Manages application state and transitions
- Supports subtask management with collapsible sections
- Offers filtering and sorting capabilities
- Styles UI components using Lip Gloss
- Follows a well-organized MVU (Model-View-Update) architecture

### Authentication Adapter (`auth/`)

The authentication adapter handles:

- Password hashing and verification
- Session management
- User identification and authorization

### Backup Adapter (`backup/`)

The backup adapter provides:

- File-based backup of task and user data
- Concurrent backup operations using goroutines
- Scheduled automated backups
- Error handling and retry mechanisms

## Development Guidelines

When working with adapters:

1. Keep adapter implementations isolated from each other
2. Don't leak implementation details to the core domain
3. Handle errors appropriately and convert to domain errors when crossing boundaries
4. Use dependency injection for external libraries
5. Write integration tests that verify adapter behavior against real systems

### Adding a New Adapter

To add a new adapter (e.g., for a new database type or UI):

1. Ensure the corresponding port interface exists in `ports/`
2. Create a new directory under `adapters/`
3. Implement the interface defined in the port
4. Update the application configuration to use the new adapter
5. Write tests that verify the adapter works correctly
