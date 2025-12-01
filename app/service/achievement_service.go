package service

import (
	"context"
	"time"

	modelMongo "be_uas/app/model/mongodb"
	modelPG "be_uas/app/model/postgres"
	repoMongo "be_uas/app/repository/mongodb"
	repoPG "be_uas/app/repository/postgres"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

type AchievementService struct {
	RepoPG    repoPG.IAchievementRepoPG
	RepoMongo repoMongo.IAchievementRepoMongo
}

func NewAchievementService(pg repoPG.IAchievementRepoPG, mongo repoMongo.IAchievementRepoMongo) *AchievementService {
	return &AchievementService{
		RepoPG:    pg,
		RepoMongo: mongo,
	}
}

// Create Achievement (Draft)
func (s *AchievementService) CreateAchievement(c *fiber.Ctx) error {
	var req modelMongo.Achievement
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid Input"})
	}

	// Ambil Student ID berdasarkan User yang Login
	userID := c.Locals("user_id").(string)
	studentID, err := s.RepoPG.GetStudentIDByUserID(userID)
	if err != nil {
		return c.Status(403).JSON(fiber.Map{"error": "User is not registered as a student"})
	}

	req.StudentID = studentID
	req.CreatedAt = time.Now()
	req.UpdatedAt = time.Now()

	// Simpan Detail ke Mongo
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	mongoID, err := s.RepoMongo.InsertAchievement(ctx, req)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to save details to MongoDB"})
	}

	// Simpan Referensi ke Postgres
	refID := uuid.New().String()
	ref := modelPG.AchievementReference{
		ID:                 refID,
		StudentID:          studentID,
		MongoAchievementID: mongoID,
		Status:             "draft",
		CreatedAt:          time.Now(),
		UpdatedAt:          time.Now(),
	}

	if err := s.RepoPG.CreateReference(ref); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to save reference to PostgreSQL"})
	}

	return c.Status(201).JSON(fiber.Map{
		"message": "Achievement created",
		"ref_id":  refID,
		"data":    req,
	})
}

// Submit for Verification
func (s *AchievementService) SubmitForVerification(c *fiber.Ctx) error {
	id := c.Params("id") // ID Referensi Postgres

	ref, err := s.RepoPG.GetReferenceByID(id)
	if err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "Achievement not found"})
	}

	if ref.Status != "draft" {
		return c.Status(400).JSON(fiber.Map{"error": "Only draft achievements can be submitted"})
	}

	if err := s.RepoPG.UpdateStatus(id, "submitted"); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to update status"})
	}

	return c.JSON(fiber.Map{"message": "Achievement submitted successfully"})
}

// Delete (Soft Delete)
func (s *AchievementService) DeleteAchievement(c *fiber.Ctx) error {
	id := c.Params("id")

	ref, err := s.RepoPG.GetReferenceByID(id)
	if err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "Achievement not found"})
	}

	if ref.Status != "draft" {
		return c.Status(400).JSON(fiber.Map{"error": "Only draft achievements can be deleted"})
	}

	// Soft Delete Mongo
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := s.RepoMongo.SoftDeleteAchievement(ctx, ref.MongoAchievementID); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to delete details"})
	}

	// Update Status Postgres
	if err := s.RepoPG.UpdateStatus(id, "deleted"); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to update status"})
	}

	return c.JSON(fiber.Map{"message": "Achievement deleted successfully"})
}

// View Prestasi Mahasiswa Bimbingan
func (s *AchievementService) GetAdviseeAchievements(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(string)

	// Ambil list referensi dari Postgres
	refs, err := s.RepoPG.GetAchievementsByAdvisorID(userID)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to fetch achievements"})
	}

	// Loop dan ambil detail dari MongoDB (Data Merging)
	var results []map[string]interface{}
	ctx := context.Background()

	for _, ref := range refs {
		detail, _ := s.RepoMongo.FindAchievementByID(ctx, ref.MongoAchievementID)
		
		// Gabungkan data untuk response
		item := map[string]interface{}{
			"ref_id":     ref.ID,
			"status":     ref.Status,
			"created_at": ref.CreatedAt,
			"detail":     detail, // Data dari Mongo
		}
		results = append(results, item)
	}

	return c.JSON(fiber.Map{"data": results})
}

// Approve Prestasi
func (s *AchievementService) VerifyAchievement(c *fiber.Ctx) error {
	id := c.Params("id") // Ref ID
	userID := c.Locals("user_id").(string)

	// Update status jadi 'verified'
	if err := s.RepoPG.UpdateVerification(id, "verified", userID, nil); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to verify achievement"})
	}

	return c.JSON(fiber.Map{"message": "Achievement verified successfully"})
}

// Reject Prestasi
func (s *AchievementService) RejectAchievement(c *fiber.Ctx) error {
	id := c.Params("id")
	userID := c.Locals("user_id").(string)

	// Parse body untuk catatan penolakan
	type RejectReq struct {
		Note string `json:"note"`
	}
	var req RejectReq
	if err := c.BodyParser(&req); err != nil || req.Note == "" {
		return c.Status(400).JSON(fiber.Map{"error": "Rejection note is required"})
	}

	// Update status jadi 'rejected'
	if err := s.RepoPG.UpdateVerification(id, "rejected", userID, &req.Note); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to reject achievement"})
	}

	return c.JSON(fiber.Map{"message": "Achievement rejected"})
}