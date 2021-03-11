package server

import (
	"sort"

	"github.com/gin-gonic/gin"

	"tanghu.com/go-micro/common/starter/server/controller"
	"tanghu.com/go-micro/common/starter/server/middleware"
)

func InitGinServer(r *gin.Engine) {
	// controller
	ctrls := controller.Ctrls()
	for _, ctrl := range ctrls {
		g := r.Group(ctrl.Name(), getHandlerFuncList(ctrl.Middlewares())...)
		for _, router := range ctrl.Routers() {
			g.Handle(router.Method, router.Path, append(getHandlerFuncList(router.Middlewares), router.Handler)...)
		}
	}
}

// get middlewares's handlerFuncs
func getHandlerFuncList(names []string) []gin.HandlerFunc {
	var handlerFuncList []gin.HandlerFunc
	var middlewares []middleware.Middleware

	allMiddlewares := middleware.Middlewares()
	for _, name := range names {
		middlewares = append(middlewares, allMiddlewares[name])
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
