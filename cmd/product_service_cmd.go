package cmd

import (
	"bufio"
	"os"

	"github.com/spf13/cobra"
	"wahyuade.com/simple-e-commerce/services"
)

var runProductServiceCmd = &cobra.Command{
	Use:   "product",
	Short: "Run the product service",
	Run: func(cmd *cobra.Command, args []string) {
		service := services.New()
		service.Bootstrap(services.ProductService{})

		reader := bufio.NewReader(os.Stdin)
		reader.ReadString('\n')
	},
}
