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

// FR-001: Login Logic
func (s *AuthService) Login(c *fiber.Ctx) error {
	var req postgres.LoginRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid request body"})
	}

	// 1. Ambil User dari DB
	user, err := s.UserRepo.GetByUsername(req.Username)
	if err != nil {
		return c.Status(401).JSON(fiber.Map{"error": "Invalid credentials"})
	}

	// 2. Cek Password Hash
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		return c.Status(401).JSON(fiber.Map{"error": "Invalid credentials"})
	}

	// 3. Cek Active Status
	if !user.IsActive {
		return c.Status(403).JSON(fiber.Map{"error": "User is inactive"})
	}

	// 4. Generate JWT Token
	claims := jwt.MapClaims{
		"user_id":   user.ID,
		"username":  user.Username,
		"role":      user.RoleName, // "Admin", "Mahasiswa", dll
		"role_id":   user.RoleID,
		"exp":       time.Now().Add(time.Hour * 24).Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	t, err := token.SignedString([]byte(os.Getenv("JWT_SECRET")))
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to generate token"})
	}

	return c.JSON(fiber.Map{
		"status": "success",
		"data": fiber.Map{
			"token": t,
			"user":  user,
		},
	})
}

// Helper: Seed Admin
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