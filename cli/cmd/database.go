package cmd

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/casbin/casbin/v2"
	gormadapter "github.com/casbin/gorm-adapter/v3"
	"github.com/spf13/cobra"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

type DatabaseCmd struct {
	dsn string // 数据库连接字符串
}

func NewDatabaseCmd() *cobra.Command {
	cmd := &DatabaseCmd{}

	databaseCmd := &cobra.Command{
		Use:   "database",
		Short: "Manage database operations",
	}

	// 公共参数
	databaseCmd.PersistentFlags().StringVarP(
		&cmd.dsn,
		"dsn", "",
		"plc_casbiner:P1c_c45b1N@tcp(localhost:3306)/plc_casbin?charset=utf8mb4&parseTime=True&loc=Local",
		"Database connection string",
	)

	// 子命令
	databaseCmd.AddCommand(&cobra.Command{
		Use:   "init",
		Short: "Initialize database",
		Run:   cmd.initDatabase,
	})

	databaseCmd.AddCommand(&cobra.Command{
		Use:   "add [user] [role]",
		Short: "Add user-role mapping",
		Args:  cobra.ExactArgs(2),
		Run:   cmd.addUserRole,
	})

	databaseCmd.AddCommand(&cobra.Command{
		Use:   "reset",
		Short: "Reset database",
		Run:   cmd.resetDatabase,
	})

	return databaseCmd
}

func (c *DatabaseCmd) getEnforcer() (*casbin.Enforcer, error) {
	// 初始化GORM适配器
	a, err := gormadapter.NewAdapterByDB(c.getGormDB())
	if err != nil {
		return nil, fmt.Errorf("failed to create adapter: %w", err)
	}

	// 创建Enforcer时使用数据库适配器
	return casbin.NewEnforcer("conf/plc-rbac-model.conf", a)
}

func (c *DatabaseCmd) getGormDB() *gorm.DB {
	db, err := gorm.Open(mysql.Open(c.dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("Failed to connect database: %v", err)
	}
	return db
}

func (c *DatabaseCmd) initDatabase(cmd *cobra.Command, args []string) {
    e, err := c.getEnforcer()
    if err != nil {
        log.Fatal(err)
    }

    // 清空现有策略
    c.clear()

    // 从CSV加载策略
    if err := loadFlatPolicyFromCSV(e, "conf/plc-role-policy.csv"); err != nil {
        log.Fatal(err)
    }

    // 保存到数据库
    if err := e.SavePolicy(); err != nil {
        log.Fatalf("Failed to save policy: %v", err)
    }

    fmt.Println("Database initialized with CSV policies")
}

func (c *DatabaseCmd) addUserRole(cmd *cobra.Command, args []string) {
	e, err := c.getEnforcer()
	if err != nil {
		log.Fatal(err)
	}

	// 添加用户角色关系
	if _, err := e.AddGroupingPolicy(args[0], args[1]); err != nil {
		log.Fatalf("Failed to add grouping policy: %v", err)
	}

	if err := e.SavePolicy(); err != nil {
		log.Fatalf("Failed to save policy: %v", err)
	}

	fmt.Printf("Successfully added %s -> %s\n", args[0], args[1])
}

func (c *DatabaseCmd) resetDatabase(cmd *cobra.Command, args []string) {
    c.clear()
	fmt.Println("Database reset successfully")
}

func (c *DatabaseCmd) clear() {
    db := c.getGormDB()

    if db.Migrator().HasTable("casbin_rule") {
        db.Exec("TRUNCATE TABLE casbin_rule")
    } else {
        log.Println("casbin_rule table not found, skipping truncate")
    }
}


func loadFlatPolicyFromCSV(e *casbin.Enforcer, csvPath string) error {
    file, err := os.Open(csvPath)
    if err != nil {
        return fmt.Errorf("failed to open CSV file: %w", err)
    }
    defer file.Close()

    scanner := bufio.NewScanner(file)
    lineNum := 0

    for scanner.Scan() {
        lineNum++
        line := strings.TrimSpace(scanner.Text())
        
        // 跳过空行和注释
        if line == "" || strings.HasPrefix(line, "#") {
            continue
        }

        // 拆分CSV字段（兼容带/不带逗号空格）
        parts := strings.Split(line, ",")
        for i := range parts {
            parts[i] = strings.TrimSpace(parts[i])
        }

        if len(parts) < 2 {
            continue // 跳过无效行
        }

        switch parts[0] {
        case "p":
            if len(parts) < 4 {
                return fmt.Errorf("invalid policy format at line %d: %s", lineNum, line)
            }
            // 直接添加单条策略：p_type, v0, v1, v2
            if _, err := e.AddPolicy(parts[1], parts[2], parts[3]); err != nil {
                return fmt.Errorf("failed to add policy at line %d: %w", lineNum, err)
            }
        case "g":
            if len(parts) < 3 {
                return fmt.Errorf("invalid grouping format at line %d: %s", lineNum, line)
            }
            // 添加用户-角色关系
            if _, err := e.AddGroupingPolicy(parts[1], parts[2]); err != nil {
                return fmt.Errorf("failed to add grouping at line %d: %w", lineNum, err)
            }
        default:
            return fmt.Errorf("unknown policy type '%s' at line %d", parts[0], lineNum)
        }
    }

    return scanner.Err()
}