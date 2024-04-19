package service

import (
	"github.com/kcmvp/archunit/internal/sample/repository"
)

type UserService struct {
	userRepository repository.UserRepository
}

func (receiver UserService) Login() {

}

func (receiver UserService) OathLogin() {

}
