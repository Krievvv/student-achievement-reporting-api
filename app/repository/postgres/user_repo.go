package postgres

import (
	"be_uas/app/model/postgres"
	"database/sql"
	"errors"
)

type IUserRepo interface {
	GetByUsername(username string) (*postgres.User, error)
	CreateUser(user postgres.User) error
	GetAllUsers() ([]postgres.User, error)
	GetUserByID(id string) (*postgres.User, error)
	UpdateUser(user postgres.User) error
	DeleteUser(id string) error
	GetRoleIDByName(roleName string) (string, error)
	UpdateUserRole(userID, roleID string) error 
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

func (r *UserRepo) GetAllUsers() ([]postgres.User, error) {
	query := `
		SELECT u.id, u.username, u.email, u.full_name, r.name as role_name, u.is_active, u.created_at
		FROM users u
		JOIN roles r ON u.role_id = r.id
		ORDER BY u.created_at DESC
	`
	rows, err := r.DB.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []postgres.User
	for rows.Next() {
		var u postgres.User
		// Scan sesuai urutan query
		if err := rows.Scan(&u.ID, &u.Username, &u.Email, &u.FullName, &u.RoleName, &u.IsActive, &u.CreatedAt); err != nil {
			return nil, err
		}
		users = append(users, u)
	}
	return users, nil
}

func (r *UserRepo) GetUserByID(id string) (*postgres.User, error) {
	query := `
        SELECT u.id, u.username, u.email, u.full_name, u.role_id, u.is_active, r.name as role_name
        FROM users u
        JOIN roles r ON u.role_id = r.id
        WHERE u.id = $1
    `
    user := &postgres.User{}
    err := r.DB.QueryRow(query, id).Scan(
        &user.ID, &user.Username, &user.Email, &user.FullName, 
        &user.RoleID, &user.IsActive, &user.RoleName, // Scan RoleName
    )
    return user, err
}

func (r *UserRepo) UpdateUser(user postgres.User) error {
	query := `UPDATE users SET full_name = $1, is_active = $2, updated_at = NOW() WHERE id = $3`
	_, err := r.DB.Exec(query, user.FullName, user.IsActive, user.ID)
	return err
}

func (r *UserRepo) DeleteUser(id string) error {
	query := `DELETE FROM users WHERE id = $1`
	_, err := r.DB.Exec(query, id)
	return err
}

func (r *UserRepo) GetRoleIDByName(roleName string) (string, error) {
	var id string
	err := r.DB.QueryRow("SELECT id FROM roles WHERE name = $1", roleName).Scan(&id)
	return id, err
}

func (r *UserRepo) UpdateUserRole(userID, roleID string) error {
    _, err := r.DB.Exec("UPDATE users SET role_id = $1, updated_at = NOW() WHERE id = $2", roleID, userID)
    return err
}