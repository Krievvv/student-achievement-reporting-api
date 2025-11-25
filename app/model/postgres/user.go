package postgres

import (
	"time"
)

type User struct {
	ID           string    `json:"id"`
	Username     string    `json:"username"`
	Email        string    `json:"email"`
	PasswordHash string    `json:"-"` // Tidak dikirim ke JSON response
	FullName     string    `json:"full_name"`
	RoleID       string    `json:"role_id"`
	RoleName     string    `json:"role_name,omitempty"` // Untuk join query
	IsActive     bool      `json:"is_active"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// Request Body untuk Login
type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// Request untuk Seeding (Opsional)
type SeedRequest struct {
	RoleID string `json:"role_id"`
}