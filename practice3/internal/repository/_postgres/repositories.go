package postgres

import (
	"practice3/internal/repository"
	"practice3/internal/repository/_postgres/users"
)

func NewRepositories(d *Dialect) *repository.Repositories {
	return &repository.Repositories{
		User: users.NewUserRepository(d.DB),
	}
}