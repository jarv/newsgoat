# Database Migrations

This directory contains SQL migration files that are automatically applied when newsgoat starts.

## Migration File Format

Migration files must follow this naming convention:

```
XXXXXX_description.sql
```

Where:
- `XXXXXX` is a 6-digit version number (e.g., `000001`, `000002`)
- `description` is a brief description of what the migration does
- Files must have a `.sql` extension

## How Migrations Work

1. Migrations are embedded into the binary at build time
2. On startup, newsgoat checks which migrations have been applied
3. Any new migrations are applied in version order
4. Applied migrations are recorded in the `schema_migrations` table
5. Migrations are idempotent - they won't be applied twice

## Creating a New Migration

To add a new migration:

1. Create a new file in this directory with the next version number:
   ```
   sql/migrations/000003_add_new_feature.sql
   ```

2. Write your SQL statements:
   ```sql
   -- Add a new column
   ALTER TABLE feeds ADD COLUMN last_fetch_duration INTEGER DEFAULT 0;

   -- Create a new index
   CREATE INDEX IF NOT EXISTS idx_feeds_last_updated ON feeds(last_updated);
   ```

3. Build and run newsgoat - the migration will be applied automatically

## Migration Tips

- Use `IF NOT EXISTS` for CREATE statements to ensure idempotency
- Use `IF EXISTS` for DROP statements
- Keep migrations small and focused on one change
- Test migrations on a copy of your database first
- Migrations run in a transaction - if any statement fails, the entire migration is rolled back

## Existing Migrations

- `000001_create_schema_migrations.sql` - Creates the schema_migrations tracking table
- `000002_initial_schema.sql` - Creates the initial database schema (feeds, items, read_status, log_messages, settings)
