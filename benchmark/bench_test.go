package main

import (
	"fmt"
	"runtime"
	"testing"
	"time"

	"github.com/casbin/casbin/v2"
)

// testcase definition
var testCases = []struct {
	name     string
	userSize int
}{
	{"small-1", 10},
	{"small-2", 20},
	{"small-3", 30},
	{"small-4", 40},
	{"small-5", 50},
	{"small-6", 60},
	{"small-7", 70},
	{"small-8", 80},
	{"small-9", 90},
	{"medium-1", 100},
	{"medium-2", 200},
	{"medium-3", 300},
	{"medium-4", 400},
	{"medium-5", 500},
	{"medium-6", 600},
	{"medium-7", 700},
	{"medium-8", 800},
	{"medium-9", 900},
	{"large-1", 1000},
	{"large-2", 2000},
	{"large-3", 3000},
	{"large-4", 4000},
	{"large-5", 5000},
	{"large-6", 6000},
	{"large-7", 7000},
	{"large-8", 8000},
	{"large-9", 9000},
}

// role allocation pool
var roles = []string{
	"role:operator",
	"role:maintenance_tech",
	"role:guest",
	"role:data_analyst",
}

func BenchmarkEnforcer(b *testing.B) {
	for _, tc := range testCases {
		b.Run(tc.name, func(b *testing.B) {
			// initialize mem statistics
			var memStatsBefore, memStatsAfter runtime.MemStats
			runtime.GC()
			runtime.ReadMemStats(&memStatsBefore)

			// create Enforcer instance
			e, err := casbin.NewEnforcer("conf/plc-rbac-model.conf", "conf/plc-role-policy.csv")
			if err != nil {
				b.Fatalf("Failed to create enforcer: %v", err)
			}

			// dynamically generate users and allocate role.
			for i := 0; i < tc.userSize; i++ {
				user := fmt.Sprintf("testuser-%d", i)
				role := roles[i%len(roles)] // allocate role in loop.
				_, _ = e.AddGroupingPolicy(user, role)
			}

			// warm up cache
			e.Enforce("testuser-0", "device_control", "w")

			// conduct test
			startTime := time.Now()
			for n := 0; n < b.N; n++ {
				user := fmt.Sprintf("testuser-%d", n%tc.userSize)
				_, _ = e.Enforce(user, "device_control", "w")
			}
			duration := time.Since(startTime)

			// compute mem
			runtime.GC()
			runtime.ReadMemStats(&memStatsAfter)
			memUsed := memStatsAfter.HeapAlloc - memStatsBefore.HeapAlloc

			// output
			b.ReportMetric(float64(duration.Nanoseconds())/float64(b.N), "ns/op")
			b.ReportMetric(float64(memUsed)/1024/1024, "MB_mem")
		})
	}
}
