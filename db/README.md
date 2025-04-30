# Database Schema and Queries

This directory contains the database schema definitions, migrations, and SQL queries for the Tusk application.

## Structure

```plaintext
db/
├── migrations/           # SQL migration files
│   ├── 001_init.up.sql   # Initial schema creation
│   └── 001_init.down.sql # Initial schema rollback
└── queries.sql           # SQL queries for sqlc generation
```

## Migration System

Tusk uses `golang-migrate` for database schema migrations. Migrations are versioned SQL files that:

1. Allow incremental schema changes over time
2. Support rolling back changes when necessary
3. Ensure consistent database state across environments
4. Track the current schema version

### Migration Files

- **Up migrations** (e.g., `001_init.up.sql`): Apply changes to advance the schema
- **Down migrations** (e.g., `001_init.down.sql`): Revert changes to roll back the schema

### Key Schema Components

The initial schema (`001_init.up.sql`) defines:

1. **Users Table**: Stores user information and authentication details

   - User ID, username, password hash, display name
   - Creation and update timestamps
   - Account status and preferences

2. **Tasks Table**: Stores task information with recursive structure

   - Task ID, title, description, status
   - Due date, priority, completion percentage
   - Parent-child relationships for subtasks
   - Tags and categorization
   - Creation, update, and completion timestamps

3. **Database Triggers**:
   - Auto-update task parent completion percentages
   - Timestamp management for creation/updates
   - Consistency checks for task relationships

## Type-Safe Query Generation

The `queries.sql` file contains SQL queries that are processed by `sqlc` to generate type-safe Go code. These queries cover:

1. User operations (create, read, update, delete)
2. Task operations (create, read, update, delete)
3. Task hierarchy management
4. Task filtering and searching
5. Task status management

## Development Guidelines

When working with the database:

1. **Always use migrations**: Never modify the schema directly in production
2. **Test migrations**: Ensure both up and down migrations work correctly
3. **Backward compatibility**: Consider existing data when altering schemas
4. **Use transactions**: Ensure data integrity for multi-step operations
5. **Document constraints**: Comment any non-obvious constraints or triggers

### Adding a New Migration

To add a new migration:

1. Create a new migration file pair with an incremented version number:

   ```bash
   # Example for adding user settings table
   touch db/migrations/002_user_settings.up.sql
   touch db/migrations/002_user_settings.down.sql
   ```

2. In the up migration, add SQL to create new tables or modify existing ones
3. In the down migration, add SQL to revert those changes
4. Update `queries.sql` with any new queries needed
5. Run `sqlc generate` to update the generated code
6. Test the migration on a development database

### Running Migrations

Apply migrations:

```bash
migrate -path db/migrations -database "$DATABASE_URL" up
```

Roll back the most recent migration:

```bash
migrate -path db/migrations -database "$DATABASE_URL" down 1
```

Check migration status:

```bash
migrate -path db/migrations -database "$DATABASE_URL" version
```
