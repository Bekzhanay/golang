package repository

import (
	"practice3/pkg/modules"
)

type UserRepository interface {
	GetUsers() ([]modules.User, error)
	GetUserByID(id int) (*modules.User, error)
	CreateUser(u modules.User) (int, error)
	UpdateUser(id int, u modules.User) error
	DeleteUser(id int) (int64, error)
}

type Repositories struct {
	User UserRepository
}