# Command Entry Points

This directory contains the main entry points for different interfaces of the Tusk application.

## Structure

```plaintext
cmd/
├── api/              # REST API server entry point (future)
│   └── main.go       # API server initialization and setup
└── cli/              # Command-line interface entry point
    └── main.go       # CLI initialization and execution
```

## CLI Entry Point (`cli/main.go`)

The CLI entry point is the main access point for the command-line interface of Tusk. It:

1. Loads configuration from environment variables and .env file
2. Sets up logging with appropriate options
3. Establishes database connections
4. Initializes and executes CLI commands

### Key Components

- Configuration loading via `config.Load()`
- Logging initialization via `logging.InitWithOptions()`
- Database connection via `db.Connect()`
- Command execution via `cli.Execute()`

## API Entry Point (`api/main.go`)

_Note: The API server is planned for future implementation._

This will serve as the entry point for the REST API server that will expose Tusk's functionality over HTTP. The API will allow:

- Task management via RESTful endpoints
- User authentication and authorization
- Cross-platform access to Tusk functionality

## Development Guidelines

When working with entry points:

1. Keep main functions clean and focused on initialization
2. Delegate business logic to the internal packages
3. Handle startup errors appropriately with good logging
4. Ensure proper resource cleanup on exit

## How to Build

Build the CLI application:

```bash
go build -o tusk ./cmd/cli
```

Future API server (when implemented):

```bash
go build -o tusk-api ./cmd/api
```
