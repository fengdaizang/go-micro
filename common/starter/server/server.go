package server

import (
	"errors"
	"fmt"
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


func checkRouters(routers []controller.Router) error {
	// check router path/method
	for i := 0; i < len(routers); i++ {
		for j := i + 1; j < len(routers); j++ {
			if routers[i].Path == routers[j].Path && routers[i].Method == routers[j].Method {
				return errors.New(fmt.Sprintf("routerPath[%s]'s method[%s] exists", routers[i].Path, routers[j].Method))
			}
		}
	}

	// check router middleware
	ms := middleware.Middlewares()
	for _, c := range routers {
		for _, mName := range c.Middlewares {
			if _, ok := ms[mName]; !ok {
				return errors.New(fmt.Sprintf("router[%s]'s middleware[%s] not exists", c.Path, mName))
			}
		}
	}

	return nil
}