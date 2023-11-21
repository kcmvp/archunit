package module1

import "github.com/kcmvp/archunit/sample/repository"

type AppController struct {
	repo repository.UserRepository
}

func (a *AppController) name() {

}
