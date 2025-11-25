package postgres

import (
	"be_uas/app/model/postgres"
	"database/sql"
	"errors"
)

type IUserRepo interface {
	GetByUsername(username string) (*postgres.User, error)
	CreateUser(user postgres.User) error
}

type UserRepo struct {
	DB *sql.DB
}

func NewUserRepo(db *sql.DB) IUserRepo {
	return &UserRepo{DB: db}
}

// GetByUsername: Digunakan untuk Login
func (r *UserRepo) GetByUsername(username string) (*postgres.User, error) {
	query := `
		SELECT u.id, u.username, u.email, u.password_hash, u.full_name, u.role_id, u.is_active, r.name as role_name
		FROM users u
		JOIN roles r ON u.role_id = r.id
		WHERE u.username = $1
	`
	user := &postgres.User{}
	err := r.DB.QueryRow(query, username).Scan(
		&user.ID, &user.Username, &user.Email, &user.PasswordHash,
		&user.FullName, &user.RoleID, &user.IsActive, &user.RoleName,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.New("user not found")
		}
		return nil, err
	}
	return user, nil
}

// CreateUser: Digunakan untuk Seeding Admin
func (r *UserRepo) CreateUser(user postgres.User) error {
	query := `
		INSERT INTO users (username, email, password_hash, full_name, role_id, is_active, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`
	_, err := r.DB.Exec(query, user.Username, user.Email, user.PasswordHash, user.FullName, user.RoleID, user.IsActive, user.CreatedAt)
	return err
}