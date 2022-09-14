package cmd

import (
	"bufio"
	"os"

	"github.com/spf13/cobra"
	"wahyuade.com/simple-e-commerce/services"
)

var runOrderServiceCmd = &cobra.Command{
	Use:   "order",
	Short: "Run the order service",
	Run: func(cmd *cobra.Command, args []string) {
		service := services.New()
		service.Bootstrap(services.OrderService{})

		reader := bufio.NewReader(os.Stdin)
		reader.ReadString('\n')
	},
}
