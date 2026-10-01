package main

import (
	"fmt"
	"strings"
	"time"
)

// Host acceptance supplements the retained reference assessment. It never
// changes case percentiles, reference limits, samples or target-miss counts.
type hostAcceptance struct {
	Profile string     `json:"profile"`
	Passed  bool       `json:"passed"`
	Cases   []hostCase `json:"cases"`
}
type hostCase struct {
	Name       string  `json:"name"`
	MeanP90NS  float64 `json:"mean_p90_ns"`
	LimitNS    int64   `json:"limit_ns"`
	P95LimitNS int64   `json:"p95_limit_ns"`
	Passed     bool    `json:"passed"`
}

func assessHost(r report) hostAcceptance {
	a := hostAcceptance{Profile: "ryzen-4500u-balanced-v1", Passed: len(r.Cases) == 84 && r.Manifest["os"] == "linux" && r.Manifest["arch"] == "amd64" && strings.Contains(r.Manifest["cpu"], "AMD Ryzen 5 4500U") && r.Manifest["power_profile"] == "balanced"}
	expected := make(map[string]int64, 28)
	for _, bad := range []bool{false, true} {
		for _, arg := range []string{"--help", "--version"} {
			expected[fmt.Sprintf("%s/invalid-config=%v", arg, bad)] = int64(5 * time.Millisecond)
		}
	}
	for _, size := range []int{0, 100, 1000} {
		for _, command := range []string{"list", "tree", "stats", "history"} {
			for _, json := range []bool{false, true} {
				expected[fmt.Sprintf("%d/%s/json=%v", size, command, json)] = int64(15 * time.Millisecond)
			}
		}
	}
	type group struct {
		index int
		runs  [3]bool
		sum   int64
		valid bool
	}
	groups := make(map[string]*group, 28)
	for _, original := range r.Cases {
		c := original
		summarize(&c)
		limit, known := expected[c.Name]
		if !known || c.LimitNS != limit {
			a.Passed = false
			continue
		}
		g := groups[c.Name]
		if g == nil {
			hostLimit := int64(18 * time.Millisecond)
			if limit == int64(5*time.Millisecond) {
				hostLimit = limit
			}
			g = &group{index: len(a.Cases), valid: true}
			groups[c.Name] = g
			a.Cases = append(a.Cases, hostCase{Name: c.Name, LimitNS: hostLimit, P95LimitNS: c.P95LimitNS})
		}
		if c.Run < 1 || c.Run > 3 || g.runs[c.Run-1] {
			g.valid = false
		} else {
			g.runs[c.Run-1] = true
		}
		if limit == int64(15*time.Millisecond) {
			a.Cases[g.index].P95LimitNS = int64(22 * time.Millisecond)
		}
		g.sum += c.P90
		g.valid = g.valid && len(c.Warmups) == 5 && len(c.Samples) == 100 && c.Error == "" && c.P95 < a.Cases[g.index].P95LimitNS && c.P99 < c.P99LimitNS && c.Max < c.MaxLimitNS
		for _, sample := range c.Samples {
			g.valid = g.valid && sample > 0
		}
	}
	a.Passed = a.Passed && len(groups) == 28
	for _, g := range groups {
		c := &a.Cases[g.index]
		c.MeanP90NS = float64(g.sum) / 3
		c.Passed = g.valid && g.runs == [3]bool{true, true, true} && g.sum < 3*c.LimitNS
		a.Passed = a.Passed && c.Passed
	}
	return a
}
