package main

import (
	"github.com/gin-gonic/gin"

	"tanghu.com/go-micro/common/starter"
	"tanghu.com/go-micro/common/starter/server"
)

func main() {
	r := gin.Default()
	webService := starter.StartAPIServer("evidence", "v1")

	server.InitGinServer(r)
	webService.Handle("/", r)

	webService.Run()
}
