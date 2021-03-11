package controller

import (
	"errors"
	"fmt"
	"sort"

	"github.com/gin-gonic/gin"

	"tanghu.com/go-micro/common/starter/server/middleware"
)

const (
	GET    = "GET"
	POST   = "POST"
	PUT    = "PUT"
	PATCH  = "PATCH"
	DELETE = "DELETE"
)

type Router struct {
	Path        string
	Method      string
	Middlewares []string
	Handler     gin.HandlerFunc
}

type Ctrl interface {
	Name() string
	Routers() []Router
	Middlewares() []string
}

type CtrlList []Ctrl

func (c CtrlList) Len() int           { return len(c) }
func (c CtrlList) Less(i, j int) bool { return c[i].Name() < c[j].Name() }
func (c CtrlList) Swap(i, j int)      { c[i], c[j] = c[j], c[i] }

var (
	errDuplicateCtrl = errors.New("Controller exists")
	ctrls            CtrlList
)

func RegisterCtrl(ct Ctrl) error {
	for _, c := range ctrls {
		if c.Name() == ct.Name() {
			return errDuplicateCtrl
		}
	}

	if err := checkRouters(ct.Routers()); err != nil {
		return err
	}

	ctrls = append(ctrls, ct)

	sort.Sort(ctrls)

	return nil
}

func Ctrls() []Ctrl {
	return ctrls
}

func checkRouters(routers []Router) error {
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
