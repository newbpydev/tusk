package storage

import (
	"bufio"
	"bytes"
	"context"
	"database/sql"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"

	assets "github.com/newbpydev/tusk/db"
)

func TestMigrate_ConcurrentInitializers(t *testing.T) {
	if path := os.Getenv("TUSK_MIGRATION_HELPER"); path != "" {
		inv, err := loadMigrations(assets.Migrations())
		if err != nil {
			t.Fatal(err)
		}
		c, err := newConnector(path, false)
		if err != nil {
			t.Fatal(err)
		}
		db := sql.OpenDB(c)
		defer db.Close()
		db.SetMaxOpenConns(1)
		fmt.Println("ready")
		if _, err := bufio.NewReader(os.Stdin).ReadString('\n'); err != nil {
			t.Fatal(err)
		}
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if err := migrate(ctx, db, inv); err != nil {
			t.Fatal(err)
		}
		return
	}
	if testing.Short() {
		t.Skip("process concurrency requires full/race suite")
	}
	path := filepath.Join(t.TempDir(), "concurrent.db")
	executable, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	type child struct {
		cmd    *exec.Cmd
		input  io.WriteCloser
		output io.ReadCloser
		stderr *bytes.Buffer
	}
	children := make([]child, 0, 2)
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	for i := 0; i < 2; i++ {
		cmd := exec.CommandContext(ctx, executable, "-test.run=^TestMigrate_ConcurrentInitializers$")
		cmd.Env = append(os.Environ(), "TUSK_MIGRATION_HELPER="+path)
		in, err := cmd.StdinPipe()
		if err != nil {
			t.Fatal(err)
		}
		out, err := cmd.StdoutPipe()
		if err != nil {
			t.Fatal(err)
		}
		stderr := new(bytes.Buffer)
		cmd.Stderr = stderr
		if err := cmd.Start(); err != nil {
			t.Fatal(err)
		}
		t.Cleanup(func() {
			if cmd.ProcessState == nil {
				_ = cmd.Process.Kill()
				_ = cmd.Wait()
			}
		})
		scanner := bufio.NewScanner(out)
		if !scanner.Scan() || scanner.Text() != "ready" {
			t.Fatal("missing child handshake")
		}
		children = append(children, child{cmd, in, out, stderr})
	}
	for _, c := range children {
		if _, err := io.WriteString(c.input, "go\n"); err != nil {
			t.Fatal(err)
		}
		c.input.Close()
	}
	for _, c := range children {
		out, _ := io.ReadAll(c.output)
		if err := c.cmd.Wait(); err != nil {
			t.Fatalf("child failed: %v %s %s", err, out, c.stderr)
		}
	}
	db := compatibilityDB(t, path, false)
	inv, err := loadMigrations(assets.Migrations())
	if err != nil {
		t.Fatal(err)
	}
	if count, err := inspectSchema(context.Background(), db, inv); err != nil || count != 1 {
		t.Fatalf("readback=%d err=%v", count, err)
	}
}
