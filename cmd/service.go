package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

func init() {
	runStartServiceCmd.AddCommand(
		runUserServiceCmd,
		runProductServiceCmd,
		runTransactionServiceCmd,
		runOrderServiceCmd,
	)
}

var runStartServiceCmd = &cobra.Command{
	Use:   "start",
	Short: "Start specific microservice",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Error: Please specify the module that you want to start")
		fmt.Println()
		fmt.Println("Usage:")
		fmt.Println("    start [module]")
		fmt.Println()
		fmt.Println("Available Modules:")
		fmt.Println("  user          Run the user service")
		fmt.Println("  product       Run the product service")
		fmt.Println("  transaction   Run the transaction service")
		fmt.Println("  order         Run the order service")
		fmt.Println()
		fmt.Println(`Use " start --help" for more information about the start command.`)
		fmt.Println()
		os.Exit(1)
	},
}
