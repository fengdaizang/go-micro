package controller

import (
	"errors"
	"sort"

	"github.com/gin-gonic/gin"
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
	ctrls = append(ctrls, ct)

	sort.Sort(ctrls)

	return nil
}

func Ctrls() []Ctrl {
	return ctrls
}
