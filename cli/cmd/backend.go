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

type BackendCmd struct {
	sub string // Subject (username)
	obj string // Object (resource)
	act string // Action (operation)
	dsn string // Database connection string
}

func NewBackendCmd() *cobra.Command {

	cmd := &BackendCmd{}

	backendCmd := &cobra.Command{
		Use:   "backend",
		Short: "Access control backend operations",
	}

	backendCmd.PersistentFlags().StringVarP(&cmd.dsn, "dsn", "",
		"plc_casbiner:P1c_c45b1N@tcp(localhost:3306)/plc_casbin?charset=utf8mb4&parseTime=True&loc=Local",
		"Database connection string",
	)

	backendCmd.AddCommand(&cobra.Command{
		Use: "request [username] [resource] [operation]",
		Short: "Check if a user can perform an operation on a resource",
		Args: cobra.ExactArgs(3),
		Run: cmd.Request,
	})

	return backendCmd
}

func (b *BackendCmd) Request(cmd *cobra.Command, args []string) {
	// 初始化带数据库的Enforcer
	a, err := gormadapter.NewAdapterByDB(b.getGormDB())
	if err != nil {
		log.Fatalf("failed to create adapter: %v", err)
	}

	e, err := casbin.NewEnforcer("conf/plc-rbac-model.conf", a)
	if err != nil {
		log.Fatalf("Failed to create enforcer: %v", err)
	}

	fmt.Printf("Checking if %s can %s %s\n", args[0], args[2], args[1])

	// 执行权限检查
	ok, err := e.Enforce(args[0], args[1], args[2])
	if err != nil {
		log.Fatalf("Enforce error: %v", err)
	}

	// 输出结果
	if ok {
		fmt.Printf("[ALLOW] %s can %s %s\n", args[0], args[2], args[1])
	} else {
		fmt.Printf("[DENY] %s cannot %s %s\n", args[0], args[2], args[1])
	}
}

func (b *BackendCmd) getGormDB() *gorm.DB {
	db, err := gorm.Open(mysql.Open(b.dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("Failed to connect database: %v", err)
	}
	return db
}
