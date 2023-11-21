package service

import "github.com/kcmvp/archunit/sample/repository"

type UserService struct {
	userRepository repository.UserRepository
}

func (receiver UserService) Login() {

}
