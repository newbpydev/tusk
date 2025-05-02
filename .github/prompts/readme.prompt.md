# README Update Generator

Generate comprehensive updates for all README files in the repository to reflect recent changes and maintain documentation quality.

## 1. Git History Analysis

- Scan the git history since the last README update (use `git log --since="LAST_UPDATE_DATE" --name-only --pretty=format:"%h %s"`)
- Focus on commits that introduce new features, fix significant bugs, change API behavior, or modify dependencies
- Identify commit messages that mention documentation or README updates

## 2. Code Change Documentation

For each meaningful code change or feature:

- Add a concise description that explains the purpose and benefit
- Document any new dependencies with version requirements
- Update installation instructions if needed (including new prerequisites)
- Note any breaking changes with clear migration steps
- Include relevant code examples showing proper usage
- Document new configuration options with default values and explanation
- Add terminal output examples for CLI/TUI features
- Update screenshots if UI components have changed

## 3. Structure Preservation

- Maintain existing README structure and formatting
- Follow the existing writing style for consistency
- Preserve custom formatting, badges, and header hierarchy
- Keep the same level of technical detail as existing documentation
- Use consistent terminology throughout all documentation

## 4. Section Currency

Ensure all sections remain current and accurate:

### Project Overview

- Update the project description if functionality has expanded
- Revise technology stack list with version numbers
- Update architectural diagrams if structure has changed

### Setup Requirements

- Verify all prerequisites are listed with correct versions
- Update database schema information if changed
- Document any new environment variables or configuration options

### Usage Instructions

- Update CLI command examples with new options
- Revise TUI keyboard shortcuts if changed
- Add examples of new functionality
- Update troubleshooting guides for common issues

### API Documentation

- Document any new endpoints with request/response examples
- Update parameter descriptions and validation rules
- Note any changes to authentication requirements
- Include curl examples for API endpoints

### Configuration Options

- List all configuration parameters with descriptions
- Note default values and acceptable ranges
- Explain how configuration affects behavior
- Document configuration file format changes

### Project Roadmap

- Move completed items from "In Progress" to "Completed" with dates
- Add new planned features to "Upcoming" section
- Update timeline estimates if project schedule has changed
- Note any deprioritized features

## 5. Technical Verification

- Verify all links and references remain valid
- Ensure command examples work with the current version
- Check that directory structure accurately reflects current state
- Validate that installation steps work on all supported platforms
- Test code examples to ensure they compile and work as expected

## 6. Version and Changelog

- Update version numbers consistently across all documentation
- Add detailed changelog entries for each significant change
- Use semantic versioning format (MAJOR.MINOR.PATCH)
- Link changelogs to relevant GitHub issues or PRs
- Note any dependency version bumps

## 7. README Files to Update

Pay special attention to these key README files:

- `/README.md` - Main project documentation
- `/cmd/README.md` - Command entry points documentation
- `/internal/README.md` - Internal architecture documentation
- `/internal/adapters/README.md` - Adapters documentation
- `/internal/core/README.md` - Core domain documentation
- `/internal/service/README.md` - Service layer documentation
- `/db/README.md` - Database structure and migrations

## General Guidelines

- Follow standard Markdown formatting
- Maintain technical accuracy above all else
- Ensure documentation is clear and accessible to both new and existing users
- Use consistent tone and style across all documents
- Include both high-level conceptual information and detailed technical specifics
- Document not just what changed but why it changed
- Prefer tables for structured data like configuration options
- Include callouts for important warnings or breaking changes
- Use collapsible sections for verbose content (like full examples)
- Add appropriate cross-references between related README files

The final documentation should be comprehensive, accurate, and provide a smooth onboarding experience for new users while serving as a reliable reference for existing users.
