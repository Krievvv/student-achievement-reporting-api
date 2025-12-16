package service

import (
	"be_uas/app/model/postgres"
	repoPG "be_uas/app/repository/postgres"
	"be_uas/utils"
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

// Login godoc
// @Summary      Login User
// @Description  Masuk sebagai Admin, Dosen, atau Mahasiswa
// @Tags         Auth
// @Accept       json
// @Produce      json
// @Param        request body postgres.LoginRequest true "Login Credentials"
// @Success      200  {object} postgres.LoginResponse
// @Failure      400  {object} map[string]interface{}
// @Failure      401  {object} map[string]interface{}
// @Router       /auth/login [post]
func (s *AuthService) Login(c *fiber.Ctx) error {
	var req postgres.LoginRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"status": "fail", "message": "Invalid request body"})
	}

	// Cek User
	user, err := s.UserRepo.GetByUsername(req.Username)
	if err != nil {
		return c.Status(401).JSON(fiber.Map{"status": "fail", "message": "Invalid credentials"})
	}

	// Cek Password 
	if !utils.CheckPassword(user.PasswordHash, req.Password) {
		return c.Status(401).JSON(fiber.Map{"status": "fail", "message": "Invalid credentials"})
	}

	if !user.IsActive {
		return c.Status(403).JSON(fiber.Map{"status": "fail", "message": "User inactive"})
	}

	// Generate Tokens 
	t, rt, err := utils.GenerateTokens(user.ID, user.Username, user.RoleName, user.RoleID)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"status": "error", "message": "Failed to generate tokens"})
	}

	// Response
	response := postgres.LoginResponse{
		Status: "success",
	}
	response.Data.Token = t
	response.Data.RefreshToken = rt
	response.Data.User = postgres.UserDetail{
		ID:          user.ID,
		Username:    user.Username,
		FullName:    user.FullName,
		Role:        user.RoleName,
		Permissions: user.Permissions,
	}

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

// RefreshToken godoc
// @Summary      Refresh Access Token
// @Description  Dapatkan token baru menggunakan Refresh Token (Mekanisme via Header Authorization saat ini)
// @Tags         Auth
// @Security     BearerAuth
// @Produce      json
// @Success      200  {object} map[string]interface{}
// @Router       /auth/refresh [post]
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

// Logout godoc
// @Summary      Logout
// @Description  Logout user (Client side clear token)
// @Tags         Auth
// @Security     BearerAuth
// @Success      200  {object} map[string]interface{}
// @Router       /auth/logout [post]
func (s *AuthService) Logout(c *fiber.Ctx) error {
    return c.JSON(fiber.Map{"message": "Successfully logged out. Please remove token from client storage."})
}

// GetProfile godoc
// @Summary      Get User Profile
// @Description  Mendapatkan detail user yang sedang login beserta permissions
// @Tags         Auth
// @Security     BearerAuth
// @Produce      json
// @Success      200  {object} postgres.UserDetail
// @Router       /auth/profile [get]
func (s *AuthService) GetProfile(c *fiber.Ctx) error {
	// Ambil user_id dari token (diset oleh middleware)
	userID := c.Locals("user_id").(string)

	// Ambil data lengkap dari DB
	user, err := s.UserRepo.GetUserByID(userID)
	if err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "User not found"})
	}

	// Return data lengkap sesuai SRS (tanpa password hash)
	return c.JSON(fiber.Map{
		"status": "success",
		"data": fiber.Map{
			"user": postgres.UserDetail{
				ID:          user.ID,
				Username:    user.Username,
				FullName:    user.FullName,
				Role:        user.RoleName,
				Permissions: user.Permissions,
			},
		},
	})
}