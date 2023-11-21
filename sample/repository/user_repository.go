package repository

import (
	"github.com/kcmvp/archunit/sample/model"
)

type UserRepository struct {
}

func (u *UserRepository) findUserByName() model.User {
	return model.User{}
}
