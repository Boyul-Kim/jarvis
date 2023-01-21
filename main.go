package main

import (
	"burnclub/router"
)

var ProjectId string

func main() {
	server := router.SetupServer()
	server.Router.Run(":3000")
	server.Database.Close()
}
