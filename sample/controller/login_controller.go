package controller

import (
	"github.com/kcmvp/archunit/sample/service"
	"github.com/kcmvp/archunit/sample/views"
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
