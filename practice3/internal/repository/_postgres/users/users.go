package users

import (
	"database/sql"
	"errors"
	"fmt"

	"github.com/jmoiron/sqlx"

	"practice3/pkg/modules"
)

var ErrUserNotFound = errors.New("user not found")

type UserRepository struct {
	db *sqlx.DB
}

func NewUserRepository(db *sqlx.DB) *UserRepository {
	return &UserRepository{db: db}
}

func (r *UserRepository) GetUsers() ([]modules.User, error) {
	var users []modules.User
	err := r.db.Select(&users, `SELECT id, name, email, age, created_at FROM users ORDER BY id`)
	if err != nil {
		return nil, fmt.Errorf("select users: %w", err)
	}
	return users, nil
}

func (r *UserRepository) GetUserByID(id int) (*modules.User, error) {
	var u modules.User
	err := r.db.Get(&u, `SELECT id, name, email, age, created_at FROM users WHERE id = $1`, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrUserNotFound
		}
		return nil, fmt.Errorf("get user by id: %w", err)
	}
	return &u, nil
}

func (r *UserRepository) CreateUser(u modules.User) (int, error) {
	var id int
	err := r.db.QueryRow(
		`INSERT INTO users (name, email, age) VALUES ($1, $2, $3) RETURNING id`,
		u.Name, u.Email, u.Age,
	).Scan(&id)
	if err != nil {
		return 0, fmt.Errorf("insert user: %w", err)
	}
	return id, nil
}

func (r *UserRepository) UpdateUser(id int, u modules.User) error {
	res, err := r.db.Exec(
		`UPDATE users SET name = $1, email = $2, age = $3 WHERE id = $4`,
		u.Name, u.Email, u.Age, id,
	)
	if err != nil {
		return fmt.Errorf("update user: %w", err)
	}

	affected, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("rows affected: %w", err)
	}
	if affected == 0 {
		return ErrUserNotFound
	}
	return nil
}

func (r *UserRepository) DeleteUser(id int) (int64, error) {
	res, err := r.db.Exec(`DELETE FROM users WHERE id = $1`, id)
	if err != nil {
		return 0, fmt.Errorf("delete user: %w", err)
	}

	affected, err := res.RowsAffected()
	if err != nil {
		return 0, fmt.Errorf("rows affected: %w", err)
	}
	if affected == 0 {
		return 0, ErrUserNotFound
	}
	return affected, nil
}