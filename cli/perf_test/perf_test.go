package main

import (
	"fmt"
	"log"
	"runtime"
	"sync"
	"testing"
	"time"

	"github.com/casbin/casbin/v2"
	gormadapter "github.com/casbin/gorm-adapter/v3"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var (
    dbPool *gorm.DB
    once   sync.Once
)

// 性能测试配置
var dsn = "plc_casbiner:P1c_c45b1N@tcp(localhost:3306)/plc_casbin?charset=utf8mb4"

var testCases = []struct {
	name      string
	userSize  int
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
	// {"large-1", 1000},
	// {"large-2", 2000},
	// {"large-3", 3000},
	// {"large-4", 4000},
	// {"large-5", 5000},
	// {"large-6", 6000},
	// {"large-7", 7000},
	// {"large-8", 8000},
	// {"large-9", 9000},
}

func BenchmarkEnforcer(b *testing.B) {
    for _, tc := range testCases {
        b.Run(tc.name, func(b *testing.B) {
            // 初始化环境
            initDatabase()
            
            // 准备测试数据
            a, _ := gormadapter.NewAdapterByDB(dbPool)
            e, _ := casbin.NewEnforcer("../conf/plc-rbac-model.conf", a)
            
            // 批量添加策略
            policies := make([][]string, tc.userSize)
            for i := 0; i < tc.userSize; i++ {
                policies[i] = []string{fmt.Sprintf("user-%d", i), roles[i%4]}
            }
            e.AddGroupingPolicies(policies)

            // 预热缓存
            e.Enforce("user-0", "device_control", "w")

            // 重置计时器和内存基准
            b.ResetTimer()
            runtime.GC()
            var memBefore runtime.MemStats
            runtime.ReadMemStats(&memBefore)

            // 执行测试循环
			start := time.Now()
            for i := 0; i < b.N; i++ {
                user := fmt.Sprintf("user-%d", i%tc.userSize)
                _, _ = e.Enforce(user, "device_control", "w")
            }
            duration := time.Since(start)

			avgLatency := float64(duration.Nanoseconds()) / float64(b.N)

            // 计算内存消耗
            runtime.GC()
            var memAfter runtime.MemStats
            runtime.ReadMemStats(&memAfter)
            memUsed := memAfter.HeapAlloc - memBefore.HeapAlloc

            // 报告结果
			b.ReportMetric(avgLatency, "ns/op")
            b.ReportMetric(float64(memUsed)/1024/1024, "MB/op")
        })
    }
}

func initDatabase() {
    once.Do(func() {
        var err error
        dbPool, err = gorm.Open(mysql.Open(dsn), &gorm.Config{
            Logger: logger.Default.LogMode(logger.Silent),
        })
        if err != nil {
            panic(err)
        }

        // 配置连接池
        sqlDB, _ := dbPool.DB()
        sqlDB.SetMaxIdleConns(10)
        sqlDB.SetMaxOpenConns(50)
        sqlDB.SetConnMaxLifetime(time.Hour)
    })

    // 确保表结构存在
    a, _ := gormadapter.NewAdapterByDB(dbPool)
    e, _ := casbin.NewEnforcer("../conf/plc-rbac-model.conf", a)
    _ = e.LoadPolicy() // 触发表创建（如果不存在）

    // 安全清空数据（仅当表存在时）
    if dbPool.Migrator().HasTable("casbin_rule") {
        dbPool.Exec("TRUNCATE TABLE casbin_rule")
    } else {
        log.Println("casbin_rule table not found, skipping truncate")
    }

    // 重新加载初始策略
    _ = e.LoadPolicy()
}

var roles = []string{
	"role:operator",
	"role:maintenance_tech",
	"role:bus_engineer",
	"role:data_analyst",
}
