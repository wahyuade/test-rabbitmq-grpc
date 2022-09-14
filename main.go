package main

import (
	"github.com/joho/godotenv"
	"wahyuade.com/simple-e-commerce/cmd"
)

func main() {
	godotenv.Load()
	cmd.Execute()
}
