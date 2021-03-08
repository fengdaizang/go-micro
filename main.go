package main

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
	"tanghu.com/go-micro/common/starter"
	"tanghu.com/go-micro/common/starter/server"
)

func main() {
	r := gin.Default()
	webService := starter.StartAPIServer("evidence", "v1")
	server.InitGinServer(r)
	webService.Handle("/", r)

	fmt.Println(viper.GetString("fabric.config_file"))

	webService.Run()
}
