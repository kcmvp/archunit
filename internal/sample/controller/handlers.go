// nolint
package controller

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type CustomizeHandler func(c *gin.Context) error

var SayHello gin.HandlerFunc = func(c *gin.Context) {}

func LoginHandler(ctx gin.Context) {

}

type EmbeddedGroup struct {
	gin.RouterGroup
}

type GroupWithNonEmbedded struct {
	group gin.RouterGroup
}

type MyRouterGroup struct{}

func (m MyRouterGroup) Use(handlerFunc ...gin.HandlerFunc) gin.IRoutes {
	// TODO implement me
	panic("implement me")
}

func (m MyRouterGroup) Handle(s string, s2 string, handlerFunc ...gin.HandlerFunc) gin.IRoutes {
	// TODO implement me
	panic("implement me")
}

func (m MyRouterGroup) Any(s string, handlerFunc ...gin.HandlerFunc) gin.IRoutes {
	// TODO implement me
	panic("implement me")
}

func (m MyRouterGroup) GET(s string, handlerFunc ...gin.HandlerFunc) gin.IRoutes {
	// TODO implement me
	panic("implement me")
}

func (m MyRouterGroup) POST(s string, handlerFunc ...gin.HandlerFunc) gin.IRoutes {
	// TODO implement me
	panic("implement me")
}

func (m MyRouterGroup) DELETE(s string, handlerFunc ...gin.HandlerFunc) gin.IRoutes {
	// TODO implement me
	panic("implement me")
}

func (m MyRouterGroup) PATCH(s string, handlerFunc ...gin.HandlerFunc) gin.IRoutes {
	// TODO implement me
	panic("implement me")
}

func (m MyRouterGroup) PUT(s string, handlerFunc ...gin.HandlerFunc) gin.IRoutes {
	// TODO implement me
	panic("implement me")
}

func (m MyRouterGroup) OPTIONS(s string, handlerFunc ...gin.HandlerFunc) gin.IRoutes {
	// TODO implement me
	panic("implement me")
}

func (m MyRouterGroup) HEAD(s string, handlerFunc ...gin.HandlerFunc) gin.IRoutes {
	// TODO implement me
	panic("implement me")
}

func (m MyRouterGroup) Match(strings []string, s string, handlerFunc ...gin.HandlerFunc) gin.IRoutes {
	// TODO implement me
	panic("implement me")
}

func (m MyRouterGroup) StaticFile(s string, s2 string) gin.IRoutes {
	// TODO implement me
	panic("implement me")
}

func (m MyRouterGroup) StaticFileFS(s string, s2 string, system http.FileSystem) gin.IRoutes {
	// TODO implement me
	panic("implement me")
}

func (m MyRouterGroup) Static(s string, s2 string) gin.IRoutes {
	// TODO implement me
	panic("implement me")
}

func (m MyRouterGroup) StaticFS(s string, system http.FileSystem) gin.IRoutes {
	// TODO implement me
	panic("implement me")
}

func (m MyRouterGroup) Group(s string, handlerFunc ...gin.HandlerFunc) *gin.RouterGroup {
	// TODO implement me
	panic("implement me")
}

var _ gin.IRouter = (*MyRouterGroup)(nil)
var _ gin.IRouter = (*EmbeddedGroup)(nil)
