package server

import (
	"sort"

	"github.com/gin-gonic/gin"

	"tanghu.com/go-micro/common/starter/server/controller"
	"tanghu.com/go-micro/common/starter/server/middleware"
	_ "tanghu.com/go-micro/common/starter/server/middleware/handler"
)

func InitGinServer(r *gin.Engine) {
	// controller
	cs := controller.Ctrls()
	for _, c := range cs {
		g := r.Group(c.Name(), getHandlerFuncList(c.Middlewares())...)
		for _, router := range c.Routers() {
			g.Handle(router.Method, router.Path, append(getHandlerFuncList(router.Middlewares), router.Handler)...)
		}
	}
}

// get middlewares's handlerFuncs
func getHandlerFuncList(names []string) []gin.HandlerFunc {
	var handlerFuncList []gin.HandlerFunc
	var middlewares []middleware.Middleware

	mm := middleware.Middlewares()
	for _, name := range names {
		middlewares = append(middlewares, mm[name])
	}
	middlewares = orderMiddleware(middlewares)

	for _, m := range middlewares {
		handlerFuncList = append(handlerFuncList, m.Handler())
	}
	return handlerFuncList
}

func orderMiddleware(ms []middleware.Middleware) []middleware.Middleware {
	res := middleware.MiddlewareList(ms)
	sort.Sort(res)
	return res
}
