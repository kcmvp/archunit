// nolint
package service

import (
	"context"
	"github.com/kcmvp/archunit/internal/sample/model"
	"github.com/kcmvp/archunit/internal/sample/repository"
)

type Audit func(string, context.Context) []string

var auditLog Audit = func(s string, ctx context.Context) []string {
	return []string{}
}

func AuditCall(id string, ctx context.Context) []string {
	return []string{}
}

type NameService interface {
	FirstNameI() string
	LastNameI() string
}

type UserService struct {
	userRepository repository.UserRepository
}

func (receiver UserService) GetUserById(id string) (model.User, error) {
	panic("for test")
}

func (receiver UserService) GetUserByNameAndAddress(name, address string) (model.User, error) {
	panic("for test")
}

func (receiver UserService) SearchUsersByFirstName(firstName string) ([]model.User, error) {
	panic("for test")
}

type NameServiceImpl struct {
}

func (n NameServiceImpl) FirstNameI() string {
	//TODO implement me
	panic("implement me")
}

func (n NameServiceImpl) LastNameI() string {
	//TODO implement me
	panic("implement me")
}

var _ NameService = (*NameServiceImpl)(nil)
