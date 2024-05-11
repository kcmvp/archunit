// nolint
package service

import "github.com/kcmvp/archunit/internal/sample/model"

func (receiver *UserService) SearchUsersByLastName(firstName string) ([]model.User, error) {
	panic("for test")
}

type FullNameImpl struct {
}

func (f FullNameImpl) FirstNameI() string {
	//TODO implement me
	panic("implement me")
}

func (f FullNameImpl) LastNameI() string {
	//TODO implement me
	panic("implement me")
}

var _ NameService = (*FullNameImpl)(nil)
