package main

import (
	"fmt"
	"os"

	"burnclub/router"

	"github.com/joho/godotenv"
)

var ProjectId string

func goDotEnvVariable(key string) string {
	err := godotenv.Load(".env")
	if err != nil {
		fmt.Println("Error loading .env file")
	}
	return os.Getenv(key)
}

func main() {
	server := router.SetupServer()
	server.Router.Run(":3000")
	server.Database.Close()
}
