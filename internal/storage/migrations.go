package storage

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"database/sql/driver"
	"errors"
	"fmt"
	"io/fs"
	"maps"
	"path"
	"regexp"
	"strconv"
	"strings"
	"time"

	assets "github.com/newbpydev/tusk/db"
	"modernc.org/sqlite"
)

type storageError string

func (e storageError) Error() string { return string(e) }

const (
	errIncompatibleSchema storageError = "storage: incompatible schema"
	errCorruptSchema      storageError = "storage: corrupt schema"
	errMigrationOutcome   storageError = "storage: migration outcome unknown"
	dateLayout                         = "2006-01-02T15:04:05.000000000Z"
)

type migration struct {
	version             int
	name, sql, checksum string
}

func loadMigrations(files fs.FS) ([]migration, error) {
	names, err := fs.Glob(files, "migrations/*")
	if err != nil {
		return nil, err
	}
	if len(names) == 0 {
		return nil, fmt.Errorf("%w: empty inventory", errIncompatibleSchema)
	}
	pattern := regexp.MustCompile(`^([0-9]{3})_[a-z][a-z0-9_]*\.sql$`)
	// Migration authors cannot escape the transaction or change file/journal state.
	forbidden := regexp.MustCompile(`(?i)\b(begin|commit|end|rollback|savepoint|release|attach|detach|vacuum|pragma)\b`)
	result := make([]migration, 0, len(names))
	for i, filename := range names {
		name := path.Base(filename)
		match := pattern.FindStringSubmatch(name)
		if match == nil {
			return nil, fmt.Errorf("%w: invalid migration filename", errIncompatibleSchema)
		}
		version, _ := strconv.Atoi(match[1])
		if version != i+1 {
			return nil, fmt.Errorf("%w: noncontiguous migration %s", errIncompatibleSchema, name)
		}
		data, err := fs.ReadFile(files, filename)
		if err != nil {
			return nil, err
		}
		if len(strings.TrimSpace(string(data))) == 0 || forbidden.Match(data) {
			return nil, fmt.Errorf("%w: prohibited migration %s", errIncompatibleSchema, name)
		}
		result = append(result, migration{version: version, name: name, sql: string(data), checksum: fmt.Sprintf("%x", sha256.Sum256(data))})
	}
	return result, nil
}

type schemaReader interface {
	QueryContext(context.Context, string, ...any) (*sql.Rows, error)
	QueryRowContext(context.Context, string, ...any) *sql.Row
}

func schemaObjects(ctx context.Context, q schemaReader) (map[string]string, error) {
	rows, err := q.QueryContext(ctx, "SELECT type, name, sql FROM sqlite_schema WHERE name NOT GLOB 'sqlite_*' ORDER BY name")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	objects := make(map[string]string)
	for rows.Next() {
		var kind, name, source string
		if err := rows.Scan(&kind, &name, &source); err != nil {
			return nil, err
		}
		objects[kind+":"+name] = source
	}
	return objects, rows.Err()
}

// expectedSchema uses SQLite's generated catalog for exact embedded SQL bytes.
// The canonical gate recompiles it in private memory and checks every prefix.
// Nonmatching inventories fall back to that same compiler at runtime.
func expectedSchema(ctx context.Context, inventory []migration) (map[string]string, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	data, err := fs.ReadFile(assets.Migrations(), "schema_catalog.json")
	if err == nil {
		if objects, ok := readSchemaCatalog(data, inventory); ok {
			return objects, nil
		}
	}
	return evaluateSchema(ctx, inventory)
}

// evaluateSchema lets SQLite parse canonical DDL instead of maintaining a
// second handwritten schema description.
func evaluateSchema(ctx context.Context, inventory []migration) (map[string]string, error) {
	c, err := sqlite.NewConnector(":memory:")
	if err != nil {
		return nil, err
	}
	db := sql.OpenDB(c)
	db.SetMaxOpenConns(1)
	defer db.Close()
	for _, m := range inventory {
		if _, err := db.ExecContext(ctx, m.sql); err != nil {
			return nil, schemaFailure(err, errIncompatibleSchema)
		}
	}
	return schemaObjects(ctx, db)
}

// inspectSchema performs read-only compatibility checks; the caller must use a
// read-only physical connection when inspecting an existing user file.
func inspectSchema(ctx context.Context, q schemaReader, inventory []migration) (int, error) {
	return (&schemaInspector{inventory: inventory}).inspect(ctx, q)
}

// One Open/migrate call owns this inspector. Reuse only the expected catalog
// compiled from embedded DDL; the user's catalog and ledger are read every time.
// It is never shared between repositories or concurrent operations.
type schemaInspector struct {
	inventory []migration
	expected  map[int]map[string]string
}

func (s *schemaInspector) inspect(ctx context.Context, q schemaReader) (int, error) {
	inventory := s.inventory
	var app int
	if err := q.QueryRowContext(ctx, "PRAGMA application_id").Scan(&app); err != nil {
		return 0, err
	}
	objects, err := schemaObjects(ctx, q)
	if err != nil {
		return 0, err
	}
	if app == 0 && len(objects) == 0 {
		return 0, nil
	}
	if app != 0x5455534B {
		return 0, errIncompatibleSchema
	}
	rows, err := q.QueryContext(ctx, "SELECT version, filename, checksum, applied_at FROM schema_migrations ORDER BY version")
	if err != nil {
		return 0, schemaFailure(err, errCorruptSchema)
	}
	count := 0
	for rows.Next() {
		var version int
		var name, checksum, applied string
		if err := rows.Scan(&version, &name, &checksum, &applied); err != nil {
			rows.Close()
			return 0, schemaFailure(err, errCorruptSchema)
		}
		if version > len(inventory) {
			rows.Close()
			return 0, errIncompatibleSchema
		}
		if version != count+1 {
			rows.Close()
			return 0, errCorruptSchema
		}
		m := inventory[count]
		if name != m.name || checksum != m.checksum {
			rows.Close()
			return 0, errIncompatibleSchema
		}
		timestamp, err := time.Parse(dateLayout, applied)
		if err != nil || timestamp.Format(dateLayout) != applied {
			rows.Close()
			return 0, errCorruptSchema
		}
		count++
	}
	err = errors.Join(rows.Err(), rows.Close())
	if err != nil {
		return 0, err
	}
	if count == 0 {
		return 0, errCorruptSchema
	}
	expected, ok := s.expected[count]
	if !ok {
		expected, err = expectedSchema(ctx, inventory[:count])
		if err != nil {
			return 0, err
		}
		if s.expected == nil {
			s.expected = make(map[int]map[string]string)
		}
		s.expected[count] = expected
	}
	if !maps.Equal(objects, expected) {
		return 0, errCorruptSchema
	}
	return count, nil
}

func migrate(ctx context.Context, db *sql.DB, inventory []migration) (err error) {
	return (&schemaInspector{inventory: inventory}).migrate(ctx, db)
}

func (s *schemaInspector) migrate(ctx context.Context, db *sql.DB) (err error) {
	inventory := s.inventory
	if len(inventory) == 0 {
		return errIncompatibleSchema
	}
	conn, err := db.Conn(ctx)
	if err != nil {
		return err
	}
	defer conn.Close()
	read, err := conn.BeginTx(ctx, &sql.TxOptions{ReadOnly: true})
	if err != nil {
		return err
	}
	count, inspectErr := s.inspect(ctx, read)
	if rollbackErr := read.Rollback(); rollbackErr != nil {
		discardConnection(conn)
		return errors.Join(inspectErr, errMigrationOutcome, storageCause(rollbackErr))
	}
	if inspectErr != nil || count == len(inventory) {
		return inspectErr
	}
	tx, err := conn.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() {
		if rollbackErr := tx.Rollback(); rollbackErr != nil && !errors.Is(rollbackErr, sql.ErrTxDone) {
			err = errors.Join(err, errMigrationOutcome, storageCause(rollbackErr))
			discardConnection(conn)
		}
	}()
	count, err = s.inspect(ctx, tx)
	if err != nil {
		return err
	}
	if count == 0 {
		if _, err := tx.ExecContext(ctx, "PRAGMA application_id=1414878027"); err != nil {
			return err
		}
	}
	for _, m := range inventory[count:] {
		if _, err := tx.ExecContext(ctx, m.sql); err != nil {
			return fmt.Errorf("migration %s: application failed: %w", m.name, storageCause(err))
		}
		if _, err := tx.ExecContext(ctx, "INSERT INTO schema_migrations(version,filename,checksum,applied_at) VALUES(?,?,?,?)", m.version, m.name, m.checksum, time.Now().UTC().Format(dateLayout)); err != nil {
			return fmt.Errorf("migration %s: ledger write failed: %w", m.name, storageCause(err))
		}
	}
	if err := tx.Commit(); err != nil {
		discardConnection(conn)
		return errors.Join(errMigrationOutcome, storageCause(err))
	}
	return nil
}

func discardConnection(conn *sql.Conn) {
	_ = conn.Raw(func(any) error { return driver.ErrBadConn })
}
