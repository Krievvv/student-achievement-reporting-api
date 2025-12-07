package postgres

import (
	"time"
)

type User struct {
	ID           string    `json:"id"`
	Username     string    `json:"username"`
	Email        string    `json:"email"`
	PasswordHash string    `json:"password_hash"` // Tidak dikirim ke JSON response
	FullName     string    `json:"full_name"`
	RoleID       string    `json:"role_id"`
	RoleName     string    `json:"role_name,omitempty"` // Untuk join query
	IsActive     bool      `json:"is_active"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
	Permissions  []string  `json:"permissions"`
}

// Request Body untuk Login
type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type SeedRequest struct {
	RoleID string `json:"role_id"`
}

// DTO: Struct khusus untuk Response Login
type LoginResponse struct {
	Status string `json:"status"`
	Data   struct {
		Token        string   `json:"token"`
		RefreshToken string   `json:"refreshToken"`
		User         UserDetail `json:"user"` // Nested struct
	} `json:"data"`
}

// DTO: Detail User yang aman untuk dikirim
type UserDetail struct {
	ID          string   `json:"id"`
	Username    string   `json:"username"`
	FullName    string   `json:"fullName"`
	Role        string   `json:"role"`
	Permissions []string `json:"permissions"`
}