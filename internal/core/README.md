# Core Domain

This directory contains the heart of Tusk's business logic and domain models, free from external dependencies.

## Structure

```plaintext
core/
├── errors/            # Domain-specific error types
│   └── errors.go      # Common error definitions
├── task/              # Task domain model
│   └── model.go       # Core task entity and related types
└── user/              # User domain model
    └── model.go       # Core user entity and related types
```

## Domain Models

### Task Model (`task/model.go`)

The Task model is the central entity of Tusk, representing actionable items that users need to track. Key features:

- Title, description, and due date
- Priority levels (Low, Medium, High, Urgent)
- Completion status tracking
- Support for hierarchical subtasks
- Tags and categorization
- Progress calculation for parent tasks based on subtask completion

Task entities are pure Go structs with business logic methods that enforce domain rules, such as:

- Validation of task properties
- Status transitions
- Recalculation of progress percentages
- Parent-child relationship management

### User Model (`user/model.go`)

The User model represents individuals who interact with Tusk:

- Username and display name
- Password hash (not the password itself)
- User preferences
- Role-based permissions (for future multi-user scenarios)
- Authentication state

### Domain Errors (`errors/errors.go`)

Domain-specific error types that describe business rule violations:

- Validation errors
- Not found errors
- Permission errors
- Conflict errors

These error types help maintain a clear separation between domain errors and technical errors.

## Design Principles

The core domain follows these principles:

1. **Independence**: No dependencies on external packages, frameworks, or libraries
2. **Immutability**: Prefer immutable data structures when possible
3. **Validation**: Self-validating domain objects that enforce business rules
4. **Rich domain model**: Logic belongs in the domain, not in services
5. **Ubiquitous language**: Names and concepts match the problem domain

## Development Guidelines

When working with the core domain:

1. **Keep it pure**: Avoid dependencies on libraries, especially those related to I/O or external systems
2. **Focus on behavior**: Domain models should encapsulate behavior, not just data
3. **Test thoroughly**: Core domain logic should have comprehensive unit tests
4. **Use value objects**: For concepts like TaskPriority, Status, etc.
5. **Respect boundaries**: Domain models should not know about repositories or UI

### Adding New Domain Concepts

When introducing new concepts to the domain:

1. Identify whether it's a first-class entity (like Task or User) or a value object
2. Create appropriate structures with validation and business methods
3. Update existing entities if there are relationships
4. Add relevant domain error types
5. Write comprehensive tests for the new domain logic
