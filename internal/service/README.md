# Service Layer

This directory contains the service layer implementations that orchestrate domain models and ports to fulfill business use cases.

## Structure

```plaintext
service/
├── task/                  # Task service implementation
│   ├── async_wrapper.go   # Asynchronous task operations with background processing
│   ├── implementation.go  # Core task service implementation
│   ├── implementation_test.go # Tests for task service
│   └── service.go         # Task service interface definition
└── user/                  # User service implementation
    ├── implementation.go  # User service implementation
    └── service.go         # User service interface definition
```

## Purpose

The service layer acts as the application's use case orchestrator. It:

1. Coordinates interactions between domain models and external systems
2. Enforces transaction boundaries
3. Handles cross-cutting concerns like logging and error management
4. Provides a cohesive API for CLI and API adapters to use
5. Maintains separation between business logic and external dependencies

## Key Services

### Task Service (`task/`)

The task service provides operations for managing tasks:

- Creating and updating tasks
- Managing task hierarchies and dependencies
- Completing tasks and handling subtask propagation
- Searching and filtering tasks
- Asynchronous task processing for operations that can run in the background

#### Async Task Processing

The `async_wrapper.go` enables certain task operations to run asynchronously:

- Background file backups when tasks are modified
- Notification generation for due tasks
- Batch operations on multiple tasks
- Progress updates for long-running operations

### User Service (`user/`)

The user service handles user-related operations:

- User registration and account management
- Authentication and authorization
- User preference management
- Session handling

## Design Patterns

The service layer leverages several design patterns:

1. **Dependency Injection**: Services receive their dependencies through constructors
2. **Repository Pattern**: Services use repository interfaces to access persistent data
3. **Decorator Pattern**: For adding cross-cutting concerns like logging or async processing
4. **Strategy Pattern**: For implementing different approaches to common operations
5. **Factory Pattern**: For creating complex domain objects

## Development Guidelines

When working with services:

1. **Interface First**: Define service interfaces before implementation to clarify use cases
2. **Thin Layer**: Keep services thin, delegating business logic to domain models
3. **Port Dependencies**: Services should depend on ports (interfaces), not adapters (implementations)
4. **Error Handling**: Handle and transform errors appropriately for consumers
5. **Testing**: Use mocks for port dependencies to test service logic in isolation

### Adding a New Service

To add a new service:

1. Create a new directory under `service/`
2. Define the service interface in `service.go`
3. Implement the interface in `implementation.go`
4. Add tests in `_test.go` files
5. Consider whether async capabilities are needed
6. Update consumers to use the new service
