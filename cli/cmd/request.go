package cmd

import (
	"fmt"
	"log"

	"github.com/casbin/casbin/v2"
	gormadapter "github.com/casbin/gorm-adapter/v3"
	"github.com/spf13/cobra"
)

func NewRequestCmd() *cobra.Command {
	var (
		sub, obj, act string
		dsn           string
	)

	cmd := &cobra.Command{
		Use:   "request",
		Short: "Check access request",
		Run: func(cmd *cobra.Command, args []string) {
			// 初始化带数据库的Enforcer
			a, err := gormadapter.NewAdapter("mysql", dsn)
			if err != nil {
				log.Fatalf("Failed to create adapter: %v", err)
			}

			e, err := casbin.NewEnforcer("conf/plc-rbac-model.conf", a)
			if err != nil {
				log.Fatalf("Failed to create enforcer: %v", err)
			}

			// 执行权限检查
			ok, err := e.Enforce(sub, obj, act)
			if err != nil {
				log.Fatalf("Enforce error: %v", err)
			}

			// 输出结果
			if ok {
				fmt.Printf("[ALLOW] %s can %s %s\n", sub, act, obj)
			} else {
				fmt.Printf("[DENY] %s cannot %s %s\n", sub, act, obj)
			}
		},
	}

	cmd.Flags().StringVarP(&sub, "sub", "s", "", "Subject (username)")
	cmd.Flags().StringVarP(&obj, "obj", "o", "", "Object (resource)")
	cmd.Flags().StringVarP(&act, "act", "a", "", "Action (operation)")
	cmd.Flags().StringVar(&dsn, "dsn", 
		"root:password@tcp(localhost:3306)/casbin?charset=utf8mb4&parseTime=True&loc=Local",
		"Database connection string")

	return cmd
}
