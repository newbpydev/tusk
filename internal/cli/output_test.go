package cli

import (
	"bytes"
	"context"
	"errors"
	"io"
	"strings"
	"testing"

	"github.com/newbpydev/tusk/internal/ports"
)

func TestOutput_CommittedEncodeFailure(t *testing.T) {
	var out, errout bytes.Buffer
	closed := 0
	got := fixtureRun(context.Background(), []string{"probe", "title"}, Options{Stdout: &out, Stderr: &errout, OpenService: func(context.Context, Config) (ports.TaskService, func() error, error) {
		return fakeService{}, func() error { closed++; return nil }, nil
	}}, func(ports.TaskService) ([]byte, bool, error) {
		return nil, true, errors.New("PRIVATE conversion failure")
	})
	if got != 1 || out.Len() != 0 || closed != 1 || !strings.Contains(errout.String(), "committed") {
		t.Fatalf("%d %d %q", got, closed, errout.String())
	}
}
func TestOutput_WriterFailures(t *testing.T) {
	for _, limit := range []int{0, 3, 10} {
		for _, fail := range []bool{false, true} {
			w := &countWriter{limit: limit, fail: fail}
			err := writeAll(w, []byte("one document\n"))
			if err == nil || w.calls != 1 {
				t.Fatalf("%+v err %v", w, err)
			}
		}
	}
}

type countWriter struct {
	limit, calls int
	fail         bool
}

func (w *countWriter) Write(p []byte) (int, error) {
	w.calls++
	if w.fail {
		return min(w.limit, len(p)), io.ErrClosedPipe
	}
	return min(w.limit, len(p)), nil
}

func TestOutput_UnknownOutcomePrecedesCause(t *testing.T) {
	for _, cause := range []error{context.Canceled, ports.ErrConflict, ports.ErrBusy, ports.ErrCorrupt} {
		value := ports.NewTransactionError("PRIVATE", cause)
		for _, unknown := range []error{value, &value} {
			code, msg := diagnostic(errors.Join(cause, unknown), false)
			if code != 1 || !strings.Contains(msg, "outcome unknown") || strings.Contains(msg, "PRIVATE") {
				t.Fatalf("%d %q", code, msg)
			}
		}
	}
}

func TestOutput_LargeJSONAllocation(t *testing.T) {
	task := jsonFixture()
	task.Description = strings.Repeat("n", 1<<20)
	result := testing.Benchmark(func(b *testing.B) {
		for range b.N {
			code := fixtureRun(context.Background(), []string{"probe", "title"}, Options{Stdout: io.Discard, OpenService: func(context.Context, Config) (ports.TaskService, func() error, error) {
				return fakeService{}, func() error { return nil }, nil
			}}, func(ports.TaskService) ([]byte, bool, error) { data, err := encodeJSON(&task); return data, false, err })
			if code != 0 {
				b.Fatal(code)
			}
		}
	})
	if result.AllocedBytesPerOp() >= 2<<20 {
		t.Fatalf("large JSON output allocates %d bytes; budget <2MiB", result.AllocedBytesPerOp())
	}
}
