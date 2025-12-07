package service

import (
	"be_uas/app/model/postgres"
	repoPG "be_uas/app/repository/postgres"
	"os"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

type AuthService struct {
	UserRepo repoPG.IUserRepo
}

func NewAuthService(userRepo repoPG.IUserRepo) *AuthService {
	return &AuthService{UserRepo: userRepo}
}

// Login Logic
func (s *AuthService) Login(c *fiber.Ctx) error {
	var req postgres.LoginRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"status": "fail", "message": "Invalid request body"})
	}

	// Ambil User dari Database
	user, err := s.UserRepo.GetByUsername(req.Username)
	if err != nil {
		return c.Status(401).JSON(fiber.Map{"status": "fail", "message": "Invalid username or password"})
	}

	// Verifikasi Password 
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		return c.Status(401).JSON(fiber.Map{"status": "fail", "message": "Invalid username or password"})
	}

	// Cek Status Aktif
	if !user.IsActive {
		return c.Status(403).JSON(fiber.Map{"status": "fail", "message": "User account is inactive"})
	}

	// Generate ACCESS TOKEN 24 Jam
	// Token ini digunakan untuk setiap request ke API
	claims := jwt.MapClaims{
		"user_id":  user.ID,
		"username": user.Username,
		"role":     user.RoleName,
		"role_id":  user.RoleID,
		"exp":      time.Now().Add(time.Hour * 24).Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	t, err := token.SignedString([]byte(os.Getenv("JWT_SECRET")))
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"status": "error", "message": "Failed to generate access token"})
	}

	// 6. Generate REFRESH TOKEN 7 Hari
	// Token ini disimpan di client untuk mendapatkan access token baru tanpa login ulang
	refreshClaims := jwt.MapClaims{
		"user_id": user.ID,
		"exp":     time.Now().Add(time.Hour * 24 * 7).Unix(),
	}
	refreshToken := jwt.NewWithClaims(jwt.SigningMethodHS256, refreshClaims)
	rt, err := refreshToken.SignedString([]byte(os.Getenv("JWT_SECRET")))
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"status": "error", "message": "Failed to generate refresh token"})
	}

	// Konstruksi Response Menggunakan DTO (Data Transfer Object)
	response := postgres.LoginResponse{
		Status: "success",
	}
	
	// Isi Data Token
	response.Data.Token = t
	response.Data.RefreshToken = rt
	
	// Data User (Mapping dari Entity DB ke DTO Response)
	response.Data.User = postgres.UserDetail{
		ID:       user.ID,
		Username: user.Username,
		FullName: user.FullName,
		Role:     user.RoleName,
		// Contoh permission statis (bisa dikembangkan ambil dari DB)
		Permissions: []string{"achievement:create", "achievement:read"}, 
	}

	// Return JSON Response
	return c.JSON(response)
}

// Seed Admin
func (s *AuthService) SeedAdmin(c *fiber.Ctx) error {
	var req postgres.SeedRequest
	if err := c.BodyParser(&req); err != nil || req.RoleID == "" {
		return c.Status(400).JSON(fiber.Map{"error": "Role ID is required in body"})
	}

	// Password default: admin123
	bytes, _ := bcrypt.GenerateFromPassword([]byte("admin123"), 14)
	
	admin := postgres.User{
		Username:     "admin_super",
		Email:        "admin@univ.ac.id",
		PasswordHash: string(bytes),
		FullName:     "Super Admin",
		RoleID:       req.RoleID, 
		IsActive:     true,
		CreatedAt:    time.Now(),
	}

	if err := s.UserRepo.CreateUser(admin); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(fiber.Map{"message": "Admin seeded. Password: admin123"})
}

func (s *AuthService) RefreshToken(c *fiber.Ctx) error {
    userID := c.Locals("user_id").(string)
    user, err := s.UserRepo.GetUserByID(userID) 
    if err != nil {
        return c.Status(401).JSON(fiber.Map{"error": "Please login again"})
    }

    // Generate New Token (Logic sama dengan Login)
    claims := jwt.MapClaims{
        "user_id":   user.ID,
        "username":  user.Username,
        "role":      user.RoleName,
        "role_id":   user.RoleID,
        "exp":       time.Now().Add(time.Hour * 24).Unix(),
    }
    token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
    t, _ := token.SignedString([]byte(os.Getenv("JWT_SECRET")))

    return c.JSON(fiber.Map{"token": t})
}

func (s *AuthService) Logout(c *fiber.Ctx) error {
    return c.JSON(fiber.Map{"message": "Successfully logged out. Please remove token from client storage."})
}