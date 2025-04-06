package main

import (
	"log"
	"os"

	cmd "github.com/nihilKnight/casbin-section/cli/cmd"
	"github.com/spf13/cobra"
)

func main() {
	rootCmd := &cobra.Command{
		Use:   "plc-casbin-cli",
		Short: "Casbin-based PLC access control system",
	}

	rootCmd.AddCommand(cmd.NewDatabaseCmd())
	rootCmd.AddCommand(cmd.NewRequestCmd())

	if err := rootCmd.Execute(); err != nil {
		log.Fatal(err)
		os.Exit(1)
	}
}
