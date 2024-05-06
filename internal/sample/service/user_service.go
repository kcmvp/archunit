// nolint
package service

import (
	"github.com/kcmvp/archunit/internal/sample/model"
	"github.com/kcmvp/archunit/internal/sample/repository"
)

type UserService struct {
	userRepository repository.UserRepository
}

func (receiver UserService) GetUserById(id string) (model.User, error) {
	panic("for test")
}

func (receiver UserService) GetUserByNameAndAddress(name, address string) (model.User, error) {
	panic("for test")
}

func (receiver UserService) SearchUsersByFirsName(firstName string) ([]model.User, error) {
	panic("for test")
}
