package controller

import (
	"github.com/kcmvp/archunit/internal/sample/service"
	_ "github.com/kcmvp/archunit/internal/sample/views"
)

type LoginController struct {
	userService service.UserService
}
