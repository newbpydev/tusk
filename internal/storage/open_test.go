package storage

import (
	"context"
	"database/sql/driver"
	"errors"
	"os"
	"path/filepath"
	"runtime"
	"sync"
	"testing"
)

func TestOpen_LiteralFilename(t *testing.T) {
	names := []string{"spaces ü % # =.db"}
	if runtime.GOOS != "windows" {
		names = append(names, "?mode=memory.db", ":memory:", "file:other.db")
	}
	for _, name := range names {
		path := filepath.Join(t.TempDir(), name)
		r, err := Open(context.Background(), Options{Path: path})
		if err != nil {
			t.Fatal(err)
		}
		if err := r.Close(); err != nil {
			t.Fatal(err)
		}
		info, err := os.Stat(path)
		if err != nil || info.Size() == 0 {
			t.Fatalf("literal file missing: %v", err)
		}
	}
}

func TestOpen_RejectsUnsafeTarget(t *testing.T) {
	dir := t.TempDir()
	for _, path := range []string{dir, filepath.Join(dir, "bad\x00name"), filepath.Join(dir, "bad\xffname")} {
		if r, err := Open(context.Background(), Options{Path: path}); err == nil {
			r.Close()
			t.Fatalf("accepted %q", path)
		}
	}
	if runtime.GOOS != "windows" {
		target := filepath.Join(dir, "target.db")
		if err := os.WriteFile(target, nil, 0600); err != nil {
			t.Fatal(err)
		}
		link := filepath.Join(dir, "link.db")
		if err := os.Symlink(target, link); err != nil {
			t.Fatal(err)
		}
		if r, err := Open(context.Background(), Options{Path: link}); err == nil {
			r.Close()
			t.Fatal("symlink accepted")
		}
	}
}

func TestOpen_ClosesPartialResources(t *testing.T) {
	for _, role := range []connectionRole{inspectionConnection, writerConnection, readerConnection} {
		path := filepath.Join(t.TempDir(), "failure.db")
		failure := errors.New("injected factory failure")
		factory := func(path string, r connectionRole) (driver.Connector, error) {
			if r == role {
				return nil, failure
			}
			return connectorFor(path, r)
		}
		if r, err := openAt(context.Background(), path, factory, prepareFile); r != nil || !errors.Is(err, failure) {
			t.Fatalf("failure=%v repo=%v", err, r)
		}
		r, err := Open(context.Background(), Options{Path: path})
		if err != nil {
			t.Fatalf("recovery: %v", err)
		}
		r.Close()
	}
	failure := errors.New("injected filesystem failure")
	if _, err := openAt(context.Background(), filepath.Join(t.TempDir(), "x.db"), connectorFor, func(string) error { return failure }); !errors.Is(err, failure) {
		t.Fatal(err)
	}
}

func TestOpen_ReplacementConnectionPragmas(t *testing.T) {
	r, err := Open(context.Background(), Options{Path: filepath.Join(t.TempDir(), "pools.db")})
	if err != nil {
		t.Fatal(err)
	}
	defer r.Close()
	if r.writer.Stats().MaxOpenConnections != 1 || r.reader.Stats().MaxOpenConnections != 4 {
		t.Fatal("incorrect pool bounds")
	}
	r.writer.SetMaxIdleConns(0)
	r.reader.SetMaxIdleConns(0)
	var only int
	if err := r.reader.QueryRow("PRAGMA query_only").Scan(&only); err != nil || only != 1 {
		t.Fatalf("replacement reader: %d %v", only, err)
	}
	var timeout int
	if err := r.writer.QueryRow("PRAGMA busy_timeout").Scan(&timeout); err != nil || timeout != 5000 {
		t.Fatalf("replacement writer: %d %v", timeout, err)
	}
}

func TestOpen_CloseAdmissionAndIdempotency(t *testing.T) {
	r, err := Open(context.Background(), Options{Path: filepath.Join(t.TempDir(), "close.db")})
	if err != nil {
		t.Fatal(err)
	}
	if err := r.admit(context.Background()); err != nil {
		t.Fatal(err)
	}
	started := make(chan struct{})
	done := make(chan error, 1)
	go func() { close(started); done <- r.Close() }()
	<-started
	// Wait for the observable admission boundary, not an arbitrary sleep.
	for {
		r.mu.Lock()
		closed := r.closed
		r.mu.Unlock()
		if closed {
			break
		}
		runtime.Gosched()
	}
	if err := r.admit(context.Background()); !errors.Is(err, errClosedRepository) {
		t.Fatal(err)
	}
	select {
	case <-done:
		t.Fatal("Close did not wait for admitted work")
	default:
	}
	r.active.Done()
	if err := <-done; err != nil {
		t.Fatal(err)
	}
	var wg sync.WaitGroup
	for i := 0; i < 4; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if err := r.Close(); err != nil {
				t.Error(err)
			}
		}()
	}
	wg.Wait()
}

func TestOpen_CurrentSchemaNoMigrationWrite(t *testing.T) {
	path := filepath.Join(t.TempDir(), "current.db")
	a, err := Open(context.Background(), Options{Path: path})
	if err != nil {
		t.Fatal(err)
	}
	defer a.Close()
	held, err := a.writer.Begin()
	if err != nil {
		t.Fatal(err)
	}
	defer held.Rollback()
	b, err := Open(context.Background(), Options{Path: path})
	if err != nil {
		t.Fatalf("current open unnecessarily needs writer: %v", err)
	}
	b.Close()
}
