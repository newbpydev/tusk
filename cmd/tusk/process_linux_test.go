package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"syscall"
	"testing"
	"time"

	"golang.org/x/sys/unix"
)

func TestProcess_RealTerminalSignals(t *testing.T) {
	if testing.Short() {
		t.Skip("real Linux PTY")
	}
	binary := filepath.Join(t.TempDir(), "tusk")
	build := exec.Command("make", "build", "BUILD_OUTPUT="+binary)
	build.Dir = "../.."
	if out, err := build.CombinedOutput(); err != nil {
		t.Fatalf("%v %s", err, out)
	}
	for _, mode := range []string{"yes", "no", "eof", "interrupt", "terminate"} {
		t.Run(mode, func(t *testing.T) {
			home := t.TempDir()
			env := []string{"PATH=" + os.Getenv("PATH"), "HOME=" + home, "TUSK_DB_PATH=" + filepath.Join(home, "db"), "TUSK_TIMEZONE=UTC", "TERM=xterm-kitty"}
			add := exec.Command(binary, "add", "Terminal confirmation 界 👩‍💻", "--json")
			add.Env = env
			data, err := add.Output()
			if err != nil {
				t.Fatal(err)
			}
			var task struct {
				ID string `json:"id"`
			}
			if err = json.Unmarshal(data, &task); err != nil {
				t.Fatal(err)
			}
			fd, err := unix.Open("/dev/ptmx", unix.O_RDWR|unix.O_NOCTTY|unix.O_NONBLOCK, 0)
			if err != nil {
				t.Fatal(err)
			}
			master := os.NewFile(uintptr(fd), "pty-master")
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
			if err = unix.IoctlSetWinsize(fd, unix.TIOCSWINSZ, &unix.Winsize{Row: 40, Col: 80}); err != nil {
				t.Fatal(err)
			}
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			cmd := exec.CommandContext(ctx, binary, "delete", task.ID)
			cmd.Env = env
			cmd.Stdin = slave
			cmd.Stdout = slave
			cmd.Stderr = slave
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
			var chunk [1024]byte
			for !bytes.Contains(screen.Bytes(), []byte("[y/N]")) {
				if ctx.Err() != nil {
					_ = cmd.Process.Kill()
					_ = cmd.Wait()
					t.Fatal("prompt timeout")
				}
				fds := []unix.PollFd{{Fd: int32(fd), Events: unix.POLLIN}}
				if _, err = unix.Poll(fds, 50); err != nil {
					t.Fatal(err)
				}
				n, e := unix.Read(fd, chunk[:])
				if errors.Is(e, unix.EAGAIN) {
					continue
				}
				if e != nil {
					t.Fatal(e)
				}
				screen.Write(chunk[:n])
			}
			switch mode {
			case "yes":
				_, err = master.Write([]byte("yes\n"))
			case "no":
				_, err = master.Write([]byte("\n"))
			case "eof":
				_, err = master.Write([]byte{4})
			case "interrupt":
				err = cmd.Process.Signal(syscall.SIGINT)
			case "terminate":
				err = cmd.Process.Signal(syscall.SIGTERM)
			}
			if err != nil {
				t.Fatal(err)
			}
			err = cmd.Wait()
			want := 0
			if mode == "interrupt" || mode == "terminate" {
				want = 1
			}
			got := 0
			if err != nil {
				var exit *exec.ExitError
				if !errors.As(err, &exit) {
					t.Fatal(err)
				}
				got = exit.ExitCode()
			}
			if got != want {
				t.Fatalf("exit %d want %d", got, want)
			}
			list := exec.Command(binary, "list", "--all", "--json")
			list.Env = env
			data, err = list.Output()
			if err != nil {
				t.Fatal(err)
			}
			var tasks []any
			if err = json.Unmarshal(data, &tasks); err != nil {
				t.Fatal(err)
			}
			wantCount := 1
			if mode == "yes" {
				wantCount = 0
			}
			if len(tasks) != wantCount {
				t.Fatalf("remaining %d", len(tasks))
			}
		})
	}
}
