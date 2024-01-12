package module1

import (
	"github.com/kcmvp/archunit/sample/repository"
	v1 "github.com/kcmvp/archunit/sample/service/ext/v1"
)

type AppController struct {
	v1.LoginService
	repository.UserRepository
}

func (a *AppController) name() {

}
