package cmd

import (
	"fmt"
	"log"

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
		"root:password@tcp(localhost:3306)/casbin?charset=utf8mb4&parseTime=True&loc=Local",
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

// 实现各子命令的具体逻辑（示例代码）
func (c *DatabaseCmd) initDatabase(cmd *cobra.Command, args []string) {
	e, err := c.getEnforcer()
	if err != nil {
		log.Fatal(err)
	}

	// 从CSV加载策略
	if err := e.LoadPolicy(); err != nil {
		log.Fatalf("Failed to load policy: %v", err)
	}

	// 这里可以添加从CSV文件导入策略的逻辑
	fmt.Println("Database initialized successfully")
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
	db := c.getGormDB()

	// 清空所有策略
	if err := db.Exec("TRUNCATE TABLE casbin_rule").Error; err != nil {
		log.Fatalf("Failed to truncate table: %v", err)
	}

	fmt.Println("Database reset successfully")
}
