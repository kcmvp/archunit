package controller

import (
	"github.com/kcmvp/archunit/internal/sample/service"
	"github.com/kcmvp/archunit/internal/sample/views"
)

type LoginController struct {
	userService service.UserService
}

func (l *LoginController) Login() bool {
	l.userService.Login()
	return true
}

func (l *LoginController) LoginHis() []view.UserView {

	return []view.UserView{}
}
