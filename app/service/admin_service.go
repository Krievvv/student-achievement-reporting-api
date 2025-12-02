package service

import (
	"context"
	repoMongo "be_uas/app/repository/mongodb"
	repoPG "be_uas/app/repository/postgres"
	"be_uas/app/model/postgres"
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"
	"golang.org/x/crypto/bcrypt"
)

type AdminService struct {
	RepoPG    repoPG.IUserRepo
	RepoAchPG repoPG.IAchievementRepoPG
	RepoMongo repoMongo.IAchievementRepoMongo
}

func NewAdminService(userRepo repoPG.IUserRepo, achRepo repoPG.IAchievementRepoPG, mongoRepo repoMongo.IAchievementRepoMongo) *AdminService {
	return &AdminService{
		RepoPG:    userRepo,
		RepoAchPG: achRepo,
		RepoMongo: mongoRepo,
	}
}

// Create User (Bisa set Role)
func (s *AdminService) CreateUser(c *fiber.Ctx) error {
	type CreateUserReq struct {
		Username string `json:"username"`
		Email    string `json:"email"`
		Password string `json:"password"`
		FullName string `json:"full_name"`
		RoleName string `json:"role_name"` // "Admin", "Mahasiswa", "Dosen Wali"
	}

	var req CreateUserReq
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid input"})
	}

	// Hash Password
	bytes, _ := bcrypt.GenerateFromPassword([]byte(req.Password), 10)

	// Cari Role ID
	roleID, err := s.RepoPG.GetRoleIDByName(req.RoleName)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid Role Name"})
	}

	user := postgres.User{
		Username:     req.Username,
		Email:        req.Email,
		PasswordHash: string(bytes),
		FullName:     req.FullName,
		RoleID:       roleID,
		IsActive:     true,
		CreatedAt:    time.Now(),
	}

	if err := s.RepoPG.CreateUser(user); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to create user"})
	}

	return c.Status(201).JSON(fiber.Map{"message": "User created successfully"})
}

// Get All Users
func (s *AdminService) ListUsers(c *fiber.Ctx) error {
	users, err := s.RepoPG.GetAllUsers()
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to fetch users"})
	}
	return c.JSON(fiber.Map{"data": users})
}

// Delete User
func (s *AdminService) DeleteUser(c *fiber.Ctx) error {
	id := c.Params("id")
	if err := s.RepoPG.DeleteUser(id); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to delete user"})
	}
	return c.JSON(fiber.Map{"message": "User deleted successfully"})
}

// View All Achievements (Global + Pagination)
func (s *AdminService) GetAllAchievements(c *fiber.Ctx) error {
	// Ambil Query params page & limit
	page, _ := strconv.Atoi(c.Query("page", "1"))
	limit, _ := strconv.Atoi(c.Query("limit", "10"))
	offset := (page - 1) * limit

	// Ambil Data dari Postgres
	refs, total, err := s.RepoAchPG.GetAllAchievements(limit, offset)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to fetch achievements"})
	}

	// Merge dengan MongoDB Detail
	var results []map[string]interface{}
	ctx := context.Background()

	for _, ref := range refs {
		detail, _ := s.RepoMongo.FindAchievementByID(ctx, ref.MongoAchievementID)
		item := map[string]interface{}{
			"id":         ref.ID,
			"student_id": ref.StudentID,
			"status":     ref.Status,
			"created_at": ref.CreatedAt,
			"detail":     detail,
		}
		results = append(results, item)
	}

	// Response dengan Metadata Pagination
	return c.JSON(fiber.Map{
		"data": results,
		"meta": fiber.Map{
			"page":       page,
			"limit":      limit,
			"total_data": total,
			"total_page": (total + limit - 1) / limit,
		},
	})
}