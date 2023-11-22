package service

import (
	"github.com/kcmvp/archunit/sample/noimport/service"
	"github.com/kcmvp/archunit/sample/repository"
)

type UserService struct {
	userRepository  repository.UserRepository
	externalService service.ExternalService
}

func (receiver UserService) Login() {

}
