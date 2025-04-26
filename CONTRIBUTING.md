# Contributing to Tusk

Thank you for your interest in contributing to Tusk! This document provides guidelines and instructions for contributing to this project.

## Code Style and Conventions

### Go Code Style

- Follow standard Go formatting (use `gofmt` or `go fmt`)
- Use meaningful variable and function names
- Group imports in the standard Go way:
  1. Standard library imports
  2. External dependencies
  3. Internal packages
- Follow the [Effective Go](https://golang.org/doc/effective_go) guidelines

### Comments and Documentation

- All exported functions, types, and constants must have godoc comments
- Start comments with the name of the thing being documented
- Use complete sentences with proper punctuation
- Explain "why" not just "what" in comments

### Error Handling

- Always check errors and provide context when returning them
- Use the early return pattern for error handling
- Wrap errors with additional context using `fmt.Errorf("doing X: %w", err)` or a similar pattern

### Testing

- Write table-driven tests where appropriate
- Aim for 80% or higher test coverage
- Tests should be independent and not rely on external systems when possible

## Project Structure

This project follows a hexagonal architecture:

- **Core Domain**: Contains pure business logic and domain models
- **Ports**: Define interfaces for the application to interact with external systems
- **Adapters**: Connect the application to external systems like databases and UI
- **Services**: Implement use cases by orchestrating domain models and ports

Please maintain this separation of concerns when contributing new code.

## Pull Request Process

1. Fork the repository and create a new branch for your feature
2. Ensure your code follows our style guidelines
3. Add or update tests as necessary
4. Update documentation to reflect any changes
5. Submit a pull request with a clear description of the changes

All pull requests should include:

- A clear description of what the PR does
- Any related issue numbers
- Screenshots for UI changes (if applicable)

## Commit Message Guidelines

- Use the present tense ("Add feature" not "Added feature")
- Use the imperative mood ("Move cursor to..." not "Moves cursor to...")
- Limit the first line to 72 characters or less
- Reference issues and pull requests after the first line

Example commit message:

```
Add task filtering by tag feature

- Implement tag-based filtering in the repository layer
- Add CLI command for filtering tasks by tag
- Update TUI to show tag filters

Fixes #123
```

## License

By contributing to this project, you agree that your contributions will be licensed under the project's [AGPL-3.0 License](LICENSE).
