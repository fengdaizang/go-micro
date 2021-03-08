package handler

import (
	"fmt"
	"time"

	"github.com/gin-gonic/gin"

	"tanghu.com/go-micro/common/starter/server/middleware"
)

func init() {
	if err := middleware.RegisterMiddleware(&allservices{}); err != nil {
		panic(err.Error())
	}
}

type allservices struct {
}

func (s *allservices) Name() string {
	return "allservices"
}

func (s *allservices) Sequence() int {
	return 0
}

func (s *allservices) Handler() gin.HandlerFunc {
	return func(c *gin.Context) {
		t := time.Now()
		fmt.Println("allservices请求之前")
		//请求之前

		c.Next()

		fmt.Println("allservices请求之后")
		//请求之后
		//计算整个请求过程耗时
		t2 := time.Since(t)
		fmt.Println(t2)
	}
}
