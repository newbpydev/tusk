package main

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"strings"
	"syscall"
	"testing"
	"time"

	"github.com/charmbracelet/x/ansi"
	"golang.org/x/sys/unix"
)

func TestTUIProcess_TerminalLifecycle(t *testing.T) {
	if testing.Short() {
		t.Skip("real Linux child PTY")
	}
	binary := filepath.Join(t.TempDir(), "tusk")
	build := exec.Command("make", "build", "BUILD_OUTPUT="+binary)
	build.Dir = "../.."
	if out, err := build.CombinedOutput(); err != nil {
		t.Fatalf("%v %s", err, out)
	}
	for _, mode := range []string{"q", "resize", "ctrlc", "interrupt", "terminate", "hangup"} {
		t.Run(mode, func(t *testing.T) {
			fd, err := unix.Open("/dev/ptmx", unix.O_RDWR|unix.O_NOCTTY|unix.O_NONBLOCK, 0)
			if err != nil {
				t.Fatal(err)
			}
			master := os.NewFile(uintptr(fd), "tui-master")
			defer master.Close()
			if err = unix.IoctlSetPointerInt(fd, unix.TIOCSPTLCK, 0); err != nil {
				t.Fatal(err)
			}
			number, err := unix.IoctlGetInt(fd, unix.TIOCGPTN)
			if err != nil {
				t.Fatal(err)
			}
			slave, err := os.OpenFile(fmt.Sprintf("/dev/pts/%d", number), os.O_RDWR, 0)
			if err != nil {
				t.Fatal(err)
			}
			defer slave.Close()
			if err = unix.IoctlSetWinsize(fd, unix.TIOCSWINSZ, &unix.Winsize{Row: 24, Col: 80}); err != nil {
				t.Fatal(err)
			}
			before, err := unix.IoctlGetTermios(int(slave.Fd()), unix.TCGETS)
			if err != nil {
				t.Fatal(err)
			}
			ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
			defer cancel()
			cmd := exec.CommandContext(ctx, binary, "tui")
			cmd.Env = []string{"PATH=" + os.Getenv("PATH"), "TERM=xterm-kitty", "TUSK_TIMEZONE=UTC", "TUSK_DB_PATH=" + filepath.Join(t.TempDir(), "tasks.db")}
			cmd.Stdin, cmd.Stdout, cmd.Stderr = slave, slave, slave
			cmd.SysProcAttr = &syscall.SysProcAttr{Setsid: true, Setctty: true, Ctty: 0}
			if err = cmd.Start(); err != nil {
				t.Fatal(err)
			}
			defer func() {
				if cmd.ProcessState == nil {
					_ = cmd.Process.Kill()
					_ = cmd.Wait()
				}
			}()
			var screen bytes.Buffer
			var chunk [4096]byte
			for !bytes.Contains(screen.Bytes(), []byte("No tasks")) {
				if ctx.Err() != nil {
					t.Fatalf("readiness timeout: %q", screen.String())
				}
				if _, err = unix.Poll([]unix.PollFd{{Fd: int32(fd), Events: unix.POLLIN}}, 50); err != nil {
					if errors.Is(err, unix.EINTR) {
						continue
					}
					t.Fatal(err)
				}
				n, e := unix.Read(fd, chunk[:])
				if errors.Is(e, unix.EAGAIN) || errors.Is(e, unix.EINTR) {
					continue
				}
				if e != nil {
					t.Fatalf("%v screen=%q", e, screen.String())
				}
				screen.Write(chunk[:n])
			}
			if bytes.Contains(screen.Bytes(), []byte("\x1b]11;?")) || bytes.Contains(screen.Bytes(), []byte("\x1b]10;?")) {
				t.Fatal("unsolicited global terminal color discovery")
			}
			if mode == "resize" {
				for _, step := range []struct {
					cols, rows uint16
					want       string
				}{
					{79, 23, "Resize to 80×24; Ctrl+C quits"},
					{120, 40, "╭─ > Tasks · 0 tasks " + strings.Repeat("─", 26) + "╮"},
				} {
					if err = unix.IoctlSetWinsize(fd, unix.TIOCSWINSZ, &unix.Winsize{Row: step.rows, Col: step.cols}); err != nil {
						t.Fatal(err)
					}
					var resized bytes.Buffer
					deadline := time.Now().Add(3 * time.Second)
					for !strings.Contains(ansi.Strip(resized.String()), step.want) {
						if time.Now().After(deadline) {
							t.Fatalf("no %dx%d resize frame containing %q; output=%q", step.cols, step.rows, step.want, resized.String())
						}
						if _, err = unix.Poll([]unix.PollFd{{Fd: int32(fd), Events: unix.POLLIN}}, 50); err != nil {
							if errors.Is(err, unix.EINTR) {
								continue
							}
							t.Fatal(err)
						}
						n, e := unix.Read(fd, chunk[:])
						if errors.Is(e, unix.EAGAIN) || errors.Is(e, unix.EINTR) {
							continue
						}
						if e != nil {
							t.Fatal(e)
						}
						resized.Write(chunk[:n])
						screen.Write(chunk[:n])
					}
				}
			}
			switch mode {
			case "q", "resize":
				_, err = master.Write([]byte("q"))
			case "ctrlc":
				_, err = master.Write([]byte{3})
			case "interrupt":
				err = cmd.Process.Signal(syscall.SIGINT)
			case "hangup":
				err = cmd.Process.Signal(syscall.SIGHUP)
			case "terminate":
				err = cmd.Process.Signal(syscall.SIGTERM)
			}
			if err != nil {
				t.Fatal(err)
			}
			err = cmd.Wait()
			want := 0
			if mode != "q" && mode != "resize" {
				want = 1
			}
			if got := cmd.ProcessState.ExitCode(); got != want {
				t.Fatalf("exit=%d want=%d err=%v", got, want, err)
			}
			for {
				n, e := unix.Read(fd, chunk[:])
				if errors.Is(e, unix.EAGAIN) {
					break
				}
				if e != nil {
					t.Fatal(e)
				}
				screen.Write(chunk[:n])
			}
			after, err := unix.IoctlGetTermios(int(slave.Fd()), unix.TCGETS)
			if err != nil {
				t.Fatal(err)
			}
			if !reflect.DeepEqual(before, after) {
				t.Fatal("terminal flags not restored")
			}
			if !bytes.Contains(screen.Bytes(), []byte("\x1b[?1049l")) {
				t.Fatal("alternate screen not restored")
			}
		})
	}
}
