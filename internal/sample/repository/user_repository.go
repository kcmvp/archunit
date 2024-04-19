package repository

import (
	"github.com/kcmvp/archunit/internal/sample/model"
)

type UserRepository struct {
}

func (u *UserRepository) findUserByName() model.User {
	return model.User{}
}
