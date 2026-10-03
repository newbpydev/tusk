package tui

import (
	"encoding/json"
	"reflect"
	"testing"
	"time"
)

func TestTUIMeasurement_SummaryRetainsSamplesAndRejectsMisses(t *testing.T) {
	r := tuiMeasurement{Warmups: []int64{1, 2, 3, 4, 5}, Samples: make([]int64, 100), Budget: true, Output: 1}
	for i := range r.Samples {
		r.Samples[i] = int64(100 - i)
	}
	before := append([]int64(nil), r.Samples...)
	summarizeTUI(&r)
	if !r.Passed || r.P95 != 95 || r.P99 != 99 || r.Max != 100 || !reflect.DeepEqual(before, r.Samples) {
		t.Fatal(r)
	}
	r.Samples[0] = 50_000_000
	summarizeTUI(&r)
	if r.Passed || r.Max != 50_000_000 || r.Misses != 1 {
		t.Fatal("dropped outlier", r)
	}
	r.Budget = false
	summarizeTUI(&r)
	if !r.Passed {
		t.Fatal("invented observation budget")
	}
	r.Samples[0] = 0
	summarizeTUI(&r)
	if r.Passed {
		t.Fatal("accepted invalid timing")
	}
	r.Samples = r.Samples[:99]
	summarizeTUI(&r)
	if r.Passed {
		t.Fatal("accepted partial run")
	}
	data, err := json.Marshal(r)
	if err != nil {
		t.Fatal(err)
	}
	var decoded tuiMeasurement
	if err := json.Unmarshal(data, &decoded); err != nil || !reflect.DeepEqual(r, decoded) {
		t.Fatal("lossy report", err)
	}
}

func TestTUIMeasurement_RejectsInvalidWarmupsAndEmptyWork(t *testing.T) {
	r := measureTUI("empty", 1, false, func() int { return 0 })
	if r.Passed {
		t.Fatal("empty work accepted")
	}
	r = measureTUI("valid", 1, false, func() int { time.Sleep(time.Millisecond); return 1 })
	if !r.Passed || r.P50 <= 0 {
		t.Fatal("missing complete statistics")
	}
	r.Warmups[0] = 0
	summarizeTUI(&r)
	if r.Passed {
		t.Fatal("invalid warmup accepted")
	}
}
