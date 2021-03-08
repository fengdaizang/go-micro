package middleware

import (
	"errors"

	"github.com/gin-gonic/gin"
)

type Middleware interface {
	Name() string
	Sequence() int
	Handler() gin.HandlerFunc
}

var (
	errDuplicateName = errors.New("Middleware name exists")
)

type MiddlewareList []Middleware

func (m MiddlewareList) Len() int           { return len(m) }
func (m MiddlewareList) Less(i, j int) bool { return m[i].Sequence() < m[j].Sequence() }
func (m MiddlewareList) Swap(i, j int)      { m[i], m[j] = m[j], m[i] }

var middlewareMap map[string]Middleware

func init() {
	middlewareMap = make(map[string]Middleware)
}

func RegisterMiddleware(m Middleware) error {
	if _, ok := middlewareMap[m.Name()]; ok {
		return errDuplicateName
	}
	middlewareMap[m.Name()] = m
	return nil
}

func Middlewares() map[string]Middleware {
	return middlewareMap
}
