// nolint
package controller

import (
	"context"
	"fmt"
	"time"

	"github.com/kcmvp/archunit/internal/sample/service"
	_ "github.com/kcmvp/archunit/internal/sample/views"
)

type LoginController struct {
	userService service.UserService
}

type CustomizeHandler func(c context.Context) error

type AppContext struct {
	context.Context
}

func (a AppContext) Deadline() (deadline time.Time, ok bool) {
	// TODO implement me
	panic("implement me")
}

func (a AppContext) Done() <-chan struct{} {
	// TODO implement me
	panic("implement me")
}

func (a AppContext) Err() error {
	// TODO implement me
	panic("implement me")
}

func (a AppContext) Value(key any) any {
	// TODO implement me
	panic("implement me")
}

func LoginHandler() {
	fmt.Println("for testing")
}

var _ context.Context = (*AppContext)(nil)
