package repository

import "github.com/kcmvp/archunit/internal/sample/model"

const (
	Mast  = "1"
	Slave = "2"
)

type UserRepository struct {
}

func (u UserRepository) FindUser() model.User {
	panic("implement me")
}
