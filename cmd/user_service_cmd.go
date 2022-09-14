package cmd

import (
	"bufio"
	"os"

	"github.com/spf13/cobra"
	"wahyuade.com/simple-e-commerce/services"
)

var runUserServiceCmd = &cobra.Command{
	Use:   "user",
	Short: "Run the user service",
	Run: func(cmd *cobra.Command, args []string) {
		service := services.New()
		service.Bootstrap(services.UserService{})

		reader := bufio.NewReader(os.Stdin)
		reader.ReadString('\n')
	},
}
