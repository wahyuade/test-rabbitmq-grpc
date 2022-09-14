package cmd

import (
	"bufio"
	"os"

	"github.com/spf13/cobra"
	"wahyuade.com/simple-e-commerce/services"
)

var runTransactionServiceCmd = &cobra.Command{
	Use:   "transaction",
	Short: "Run the transaction service",
	Run: func(cmd *cobra.Command, args []string) {
		service := services.New()
		service.Bootstrap(services.TransactionService{})

		reader := bufio.NewReader(os.Stdin)
		reader.ReadString('\n')
	},
}
