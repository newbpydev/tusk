# Internal Application Code

This directory contains the internal implementation of the Tusk application, following a hexagonal architecture pattern (also known as ports and adapters).

## Structure

```plaintext
internal/
├── adapters/           # External system adapters (implementation of ports)
├── app/                # Application services orchestration
├── cli/                # CLI command implementations
├── config/             # Configuration loading and parsing
├── core/               # Core domain models and business logic
├── ports/              # Interface definitions
├── service/            # Service layer implementation
└── util/               # Shared utilities
```

## Architectural Overview

Tusk follows a hexagonal architecture which separates the application into:

1. **Core Domain** (`core/`) - Contains pure business logic and domain models
2. **Ports** (`ports/`) - Define interfaces for the application to interact with external systems
3. **Adapters** (`adapters/`) - Connect the application to external systems like databases and UI
4. **Services** (`service/`) - Implement use cases by orchestrating domain models and ports

This architecture allows:

- Separation of concerns
- Testability of business logic in isolation
- Flexibility to swap implementations (e.g., different database providers)
- Parallel development of different components

## Key Components

### Core Domain

The heart of the application, containing:

- Task models with support for subtasks and priorities
- User models with authentication information
- Domain-specific errors and validation logic

### Ports

Interface definitions separated by direction:

- **Input ports**: Use cases that can be triggered by external actors
- **Output ports**: Interfaces that the application uses to interact with external systems

### Adapters

Implementations of ports for specific technologies:

- Database adapters for PostgreSQL
- Terminal UI adapter using BubbleTea
- Authentication adapter for user validation
- Backup adapter for file operations

### Services

Business logic that coordinates between domain models and ports:

- Task service for managing tasks
- User service for managing users and authentication

## Development Guidelines

When working in the internal package:

1. Maintain strict boundaries between layers:

   - Core domain should not depend on adapters
   - Services should only depend on ports, not adapters

2. Keep the core domain pure:

   - No external dependencies
   - No database or UI concerns
   - Focus on business logic only

3. Use dependency injection:

   - Pass dependencies through constructors
   - Avoid global state

4. Write tests for each layer:
   - Use mocks for ports in service tests
   - Test adapters against real systems when possible
