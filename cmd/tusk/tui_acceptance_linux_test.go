package main

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"runtime"
	"strings"
	"syscall"
	"testing"
	"time"

	"github.com/charmbracelet/x/ansi"
	"github.com/muesli/termenv"
	"github.com/newbpydev/tusk/internal/cli"
	"github.com/newbpydev/tusk/internal/core"
	"github.com/newbpydev/tusk/internal/ports"
	"github.com/newbpydev/tusk/internal/service/dateparse"
	"github.com/newbpydev/tusk/internal/tui"
	"golang.org/x/sys/unix"
)

type acceptancePTY struct {
	master, slave *os.File
	before        *unix.Termios
	cmd           *exec.Cmd
	ctx           context.Context
	cancel        context.CancelFunc
	screen        bytes.Buffer
	started       time.Time
}

func startAcceptancePTY(t *testing.T, binary string, args, env []string) *acceptancePTY {
	t.Helper()
	fd, err := unix.Open("/dev/ptmx", unix.O_RDWR|unix.O_NOCTTY|unix.O_NONBLOCK, 0)
	if err != nil {
		t.Fatal(err)
	}
	p := &acceptancePTY{master: os.NewFile(uintptr(fd), "acceptance-master")}
	t.Cleanup(func() {
		if p.cmd != nil && p.cmd.Process != nil && p.cmd.ProcessState == nil {
			_ = p.cmd.Process.Kill()
			_ = p.cmd.Wait()
		}
		if p.cancel != nil {
			p.cancel()
		}
		p.master.Close()
		if p.slave != nil {
			p.slave.Close()
		}
	})
	if err = unix.IoctlSetPointerInt(fd, unix.TIOCSPTLCK, 0); err != nil {
		t.Fatal(err)
	}
	number, err := unix.IoctlGetInt(fd, unix.TIOCGPTN)
	if err != nil {
		t.Fatal(err)
	}
	p.slave, err = os.OpenFile(fmt.Sprintf("/dev/pts/%d", number), os.O_RDWR, 0)
	if err != nil {
		t.Fatal(err)
	}
	if err = unix.IoctlSetWinsize(fd, unix.TIOCSWINSZ, &unix.Winsize{Row: 24, Col: 80}); err != nil {
		t.Fatal(err)
	}
	p.before, err = unix.IoctlGetTermios(int(p.slave.Fd()), unix.TCGETS)
	if err != nil {
		t.Fatal(err)
	}
	p.ctx, p.cancel = context.WithTimeout(context.Background(), 15*time.Second)
	p.cmd = exec.CommandContext(p.ctx, binary, args...)
	p.cmd.Env = append([]string{"PATH=" + os.Getenv("PATH"), "TERM=xterm-kitty", "TUSK_TIMEZONE=UTC"}, env...)
	p.cmd.Stdin, p.cmd.Stdout, p.cmd.Stderr = p.slave, p.slave, p.slave
	p.cmd.SysProcAttr = &syscall.SysProcAttr{Setsid: true, Setctty: true, Ctty: 0}
	p.started = time.Now()
	if err = p.cmd.Start(); err != nil {
		t.Fatal(err)
	}
	return p
}

func (p *acceptancePTY) read(t *testing.T) {
	t.Helper()
	fd := int(p.master.Fd())
	_, err := unix.Poll([]unix.PollFd{{Fd: int32(fd), Events: unix.POLLIN}}, 10)
	if err != nil && !errors.Is(err, unix.EINTR) {
		t.Fatal(err)
	}
	var chunk [8192]byte
	n, err := unix.Read(fd, chunk[:])
	if errors.Is(err, unix.EAGAIN) || errors.Is(err, unix.EINTR) {
		return
	}
	if err != nil {
		t.Fatal(err)
	}
	p.screen.Write(chunk[:n])
}
func (p *acceptancePTY) until(t *testing.T, want string) {
	t.Helper()
	for !strings.Contains(ansi.Strip(p.screen.String()), want) {
		if p.ctx.Err() != nil {
			t.Fatalf("PTY waiting for %q: %q", want, p.screen.String())
		}
		p.read(t)
	}
}
func (p *acceptancePTY) send(t *testing.T, text string) {
	t.Helper()
	if _, err := p.master.WriteString(text); err != nil {
		t.Fatal(err)
	}
}
func (p *acceptancePTY) marker(t *testing.T, path string) {
	t.Helper()
	for {
		if _, err := os.Stat(path); err == nil {
			return
		}
		if p.ctx.Err() != nil {
			t.Fatal("missing fixture barrier", path)
		}
		p.read(t)
	}
}
func (p *acceptancePTY) finish(t *testing.T, want int) {
	t.Helper()
	done := make(chan error, 1)
	go func() { done <- p.cmd.Wait() }()
	for {
		select {
		case <-done:
			goto exited
		default:
			p.read(t)
		}
	}
exited:
	for {
		var data [8192]byte
		n, err := unix.Read(int(p.master.Fd()), data[:])
		if errors.Is(err, unix.EAGAIN) {
			break
		}
		if err != nil {
			t.Fatal(err)
		}
		if n == 0 {
			break
		}
		p.screen.Write(data[:n])
	}
	if p.cmd.ProcessState.ExitCode() != want {
		t.Fatalf("exit %d want %d: %q", p.cmd.ProcessState.ExitCode(), want, p.screen.String())
	}
	after, err := unix.IoctlGetTermios(int(p.slave.Fd()), unix.TCGETS)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(p.before, after) {
		t.Fatal("terminal modes not restored")
	}
	for _, seq := range []string{"\x1b[?1049l", "\x1b[?25h", "\x1b[?2004l"} {
		if !bytes.Contains(p.screen.Bytes(), []byte(seq)) {
			t.Fatalf("missing restoration %q", seq)
		}
	}
	p.cancel()
	p.master.Close()
	p.slave.Close()
}

type acceptanceService struct {
	ports.TaskService
	mode, barrier string
	epoch         int
}

func (s acceptanceService) CreateTask(ctx context.Context, c ports.CreateTaskCommand) (*core.Task, error) {
	task, err := s.TaskService.CreateTask(ctx, c)
	if err != nil {
		return nil, err
	}
	if err = os.WriteFile(s.barrier, []byte(task.ID), 0600); err != nil {
		return nil, err
	}
	if s.mode != "recover-cancel" {
		<-ctx.Done()
	}
	if s.mode == "unknown-cancel" || s.mode == "recover-cancel" {
		return nil, ports.NewTransactionError("injected-after-commit", ports.ErrStorage)
	}
	return task, nil
}
func (s acceptanceService) GetTaskTree(ctx context.Context, id string) ([]*core.TaskNode, error) {
	if s.mode == "recover-cancel" && s.epoch > 1 {
		if err := os.WriteFile(s.barrier+"-recovery", nil, 0600); err != nil {
			return nil, err
		}
		<-ctx.Done()
		return nil, ctx.Err()
	}
	return s.TaskService.GetTaskTree(ctx, id)
}

type acceptanceOutput struct {
	*os.File
	failed bool
}

func (w *acceptanceOutput) Write(data []byte) (int, error) {
	if !w.failed && bytes.Contains(data, []byte("No tasks")) {
		w.failed = true
		return 0, io.ErrClosedPipe
	}
	return w.File.Write(data)
}

func TestTUIProcess_FaultChild(t *testing.T) {
	mode := os.Getenv("TUSK_TEST_TUI_CHILD")
	if mode == "" {
		t.Skip("child-only fault launcher")
	}
	ctx, stop := processContext()
	runner := func(ctx context.Context, cfg cli.Config) (cli.TUIResult, error) {
		factory := tuiFactory(cfg)
		epoch := 0
		open := func(ctx context.Context) (ports.TaskService, func() error, error) {
			svc, close, err := factory(ctx)
			epoch++
			if err != nil {
				return nil, nil, err
			}
			return acceptanceService{svc, mode, os.Getenv("TUSK_TEST_TUI_BARRIER"), epoch}, close, nil
		}
		var output io.Writer = os.Stdout
		if mode == "output" {
			output = &acceptanceOutput{File: os.Stdout}
		}
		result, err := tui.Run(ctx, tui.RunOptions{Input: os.Stdin, Output: output, Open: open, Location: time.UTC, Profile: termenv.Ascii, DayBounds: dateparse.DayBounds, ParseDue: dateparse.ParseDue})
		return cli.TUIResult{HadCommittedChanges: result.HadCommittedChanges, OutcomeUnknown: result.OutcomeUnknown}, err
	}
	code := cli.Run(ctx, []string{"tui"}, cli.Options{Stdout: os.Stdout, Stderr: os.Stderr, Version: Version, Getenv: os.Getenv, Terminal: terminalFacts, RunTUI: runner})
	stop()
	os.Exit(code)
}

func TestTUIProcess_TerminalRestoredOnFailure(t *testing.T) {
	if testing.Short() {
		t.Skip("real child PTY fault paths")
	}
	for _, mode := range []string{"startup", "output", "committed-cancel", "unknown-cancel", "recover-cancel"} {
		t.Run(mode, func(t *testing.T) {
			dir := t.TempDir()
			db := filepath.Join(dir, "tasks.db")
			barrier := filepath.Join(dir, "committed")
			if mode == "startup" {
				db = dir
			}
			p := startAcceptancePTY(t, os.Args[0], []string{"-test.run=^TestTUIProcess_FaultChild$"}, []string{"TUSK_TEST_TUI_CHILD=" + mode, "TUSK_DB_PATH=" + db, "TUSK_TEST_TUI_BARRIER=" + barrier})
			switch mode {
			case "startup":
				p.until(t, "Could not load tasks")
				p.send(t, "\x03")
			case "output":
			default:
				p.until(t, "No tasks")
				p.send(t, "a")
				p.until(t, "Create task")
				p.send(t, "\x1b[200~Committed via paste\x1b[201~")
				p.until(t, "Committed via paste")
				p.send(t, "\x13")
				p.marker(t, barrier)
				if mode == "recover-cancel" {
					p.until(t, "Readback and recovery")
					p.send(t, "r")
					p.marker(t, barrier+"-recovery")
				}
				p.send(t, "\x03")
			}
			p.finish(t, 1)
			text := p.screen.String()
			if mode == "committed-cancel" && !strings.Contains(text, "earlier changes remain committed") {
				t.Fatal("lost committed receipt", text)
			}
			if (mode == "unknown-cancel" || mode == "recover-cancel") && !strings.Contains(text, "outcome unknown") {
				t.Fatal("lost uncertain receipt", text)
			}
			if mode != "startup" && mode != "output" {
				t.Setenv("TUSK_DB_PATH", db)
				svc, close, err := tuiFactory(cli.Config{Location: time.UTC})(context.Background())
				if err != nil {
					t.Fatal(err)
				}
				defer close()
				tasks, err := svc.ListTasks(context.Background(), ports.TaskQuery{All: true})
				if err != nil || len(tasks) != 1 || tasks[0].Title != "Committed via paste" {
					t.Fatal("commit/readback/paste mismatch", tasks, err)
				}
			}
		})
	}
}

func TestTUIStartupMeasurements(t *testing.T) {
	output := os.Getenv("TUSK_TUI_STARTUP_OUTPUT")
	if output == "" {
		t.Skip("make bench-tui startup observation")
	}
	binary, err := measurementBinaryPath()
	if err != nil {
		t.Fatal(err)
	}
	db := filepath.Join(releaseDirectory(t), "tasks.db")
	seed := exec.Command(binary, "add", "Startup fixture")
	seed.Env = []string{"PATH=" + os.Getenv("PATH"), "TUSK_DB_PATH=" + db, "TUSK_TIMEZONE=UTC"}
	if out, err := seed.CombinedOutput(); err != nil {
		t.Fatal(err, string(out))
	}
	type startupRun struct {
		Run     int     `json:"run"`
		Warmups []int64 `json:"warmup_ns"`
		Samples []int64 `json:"samples_ns"`
		RSS     []int64 `json:"child_max_rss_kib"`
	}
	var runs []startupRun
	for run := 1; run <= 3; run++ {
		r := startupRun{Run: run}
		for i := range 105 {
			p := startAcceptancePTY(t, binary, []string{"tui"}, []string{"TUSK_DB_PATH=" + db})
			p.until(t, "Startup fixture")
			elapsed := time.Since(p.started).Nanoseconds()
			p.send(t, "q")
			p.finish(t, 0)
			if i < 5 {
				r.Warmups = append(r.Warmups, elapsed)
			} else {
				r.Samples = append(r.Samples, elapsed)
				r.RSS = append(r.RSS, p.cmd.ProcessState.SysUsage().(*syscall.Rusage).Maxrss)
			}
		}
		runs = append(runs, r)
	}
	data, err := os.ReadFile(binary)
	if err != nil {
		t.Fatal(err)
	}
	report := struct {
		Go, BinarySHA256, Policy string
		Runs                     []startupRun
	}{runtime.Version(), fmt.Sprintf("%x", sha256.Sum256(data)), "Linux child PTY launch to fixture text in populated frame; setup outside timing; no startup SLA; all samples retained; restoration checked each child", runs}
	data, err = json.MarshalIndent(report, "", "  ")
	if err != nil {
		t.Fatal(err)
	}
	if err = os.WriteFile(output, append(data, '\n'), 0644); err != nil {
		t.Fatal(err)
	}
}

// The injected child proves the in-flight race deterministically. This separate
// production executable check proves real writes retain their receipt on SIGINT.
func TestTUIProcess_CommittedBeforeInterrupt(t *testing.T) {
	if testing.Short() {
		t.Skip("production executable and real child PTY")
	}
	dir := t.TempDir()
	binary := filepath.Join(dir, "tusk")
	build := exec.Command("make", "build", "BUILD_OUTPUT="+binary)
	build.Dir = "../.."
	if out, err := build.CombinedOutput(); err != nil {
		t.Fatalf("%v: %s", err, out)
	}
	db := filepath.Join(dir, "tasks.db")
	p := startAcceptancePTY(t, binary, []string{"tui"}, []string{"TUSK_DB_PATH=" + db})
	p.until(t, "No tasks")
	p.send(t, "a")
	p.until(t, "Create task")
	p.send(t, "\x1b[200~Production commit 界\x1b[201~")
	p.until(t, "Production commit 界")
	p.send(t, "\x13")
	p.until(t, "Saved")
	if err := p.cmd.Process.Signal(syscall.SIGINT); err != nil {
		t.Fatal(err)
	}
	p.finish(t, 1)
	if !strings.Contains(p.screen.String(), "earlier changes remain committed") {
		t.Fatal("production process lost committed receipt", p.screen.String())
	}
	readback := exec.Command(binary, "list", "--all", "--json")
	readback.Env = []string{"PATH=" + os.Getenv("PATH"), "TUSK_DB_PATH=" + db, "TUSK_TIMEZONE=UTC"}
	out, err := readback.Output()
	if err != nil {
		t.Fatal(err)
	}
	var tasks []struct{ Title string }
	if err := json.Unmarshal(out, &tasks); err != nil {
		t.Fatal(err)
	}
	if len(tasks) != 1 || tasks[0].Title != "Production commit 界" {
		t.Fatalf("production readback mismatch: %s", out)
	}
}
