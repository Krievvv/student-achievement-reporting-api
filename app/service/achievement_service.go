package service

import (
	"context"
	"fmt"
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

// ==========================================
// METODE TAMBAHAN (YANG SEBELUMNYA ERROR)
// ==========================================

// 1. Get My Achievements (List Prestasi Mahasiswa Login)
func (s *AchievementService) GetMyAchievements(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(string)
	
	// Cari Student ID dari User ID
	studentID, err := s.RepoPG.GetStudentIDByUserID(userID)
	if err != nil {
		return c.Status(403).JSON(fiber.Map{"error": "User is not a student"})
	}

	// Gunakan Repo GetAchievementsByStudentID
	refs, err := s.RepoPG.GetAchievementsByStudentID(studentID)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to fetch achievements"})
	}

	// Merge dengan Data Mongo (Opsional: Jika ingin detail di list)
	// Untuk list biasanya cukup data reference, tapi jika ingin detail:
	var results []map[string]interface{}
	ctx := context.Background()

	for _, ref := range refs {
		detail, _ := s.RepoMongo.FindAchievementByID(ctx, ref.MongoAchievementID)
		item := map[string]interface{}{
			"id":         ref.ID,
			"status":     ref.Status,
			"created_at": ref.CreatedAt,
			"detail":     detail,
		}
		results = append(results, item)
	}

	if results == nil {
		results = []map[string]interface{}{}
	}

	return c.JSON(fiber.Map{"data": results})
}

// 2. Get Achievement Detail (Detail Satu Prestasi)
func (s *AchievementService) GetAchievementDetail(c *fiber.Ctx) error {
	id := c.Params("id")
	
	// Ambil Reference Postgres
	ref, err := s.RepoPG.GetReferenceByID(id)
	if err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "Achievement not found"})
	}

	// Ambil Detail Mongo
	detail, err := s.RepoMongo.FindAchievementByID(context.Background(), ref.MongoAchievementID)
	
	return c.JSON(fiber.Map{
		"data": fiber.Map{
			"ref":    ref,
			"detail": detail,
		},
	})
}

// 3. Upload Attachment
func (s *AchievementService) UploadAttachment(c *fiber.Ctx) error {
	// Ambil file dari form-data
	file, err := c.FormFile("file")
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "File is required"})
	}

	// Pastikan folder 'uploads' ada secara manual di root project
	// Simpan file ke server
	filePath := fmt.Sprintf("./uploads/%s_%s", uuid.New().String(), file.Filename)
	if err := c.SaveFile(file, filePath); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to save file"})
	}

	// NOTE: Di implementasi nyata, Anda harus mengupdate dokumen MongoDB
	// untuk menyimpan path file ini ke dalam array 'attachments'.
	// Untuk saat ini kita return path-nya saja.
	
	return c.JSON(fiber.Map{
		"message": "File uploaded successfully",
		"url":     filePath,
	})
}

// 4. Get Student Achievements (Admin/Dosen View by StudentID)
func (s *AchievementService) GetStudentAchievements(c *fiber.Ctx) error {
	studentID := c.Params("id")

	refs, err := s.RepoPG.GetAchievementsByStudentID(studentID)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to fetch achievements"})
	}

	var results []map[string]interface{}
	ctx := context.Background()
	for _, ref := range refs {
		detail, _ := s.RepoMongo.FindAchievementByID(ctx, ref.MongoAchievementID)
		item := map[string]interface{}{
			"id":         ref.ID,
			"status":     ref.Status,
			"created_at": ref.CreatedAt,
			"detail":     detail,
		}
		results = append(results, item)
	}
	
	if results == nil {
		results = []map[string]interface{}{}
	}

	return c.JSON(fiber.Map{"data": results})
}

// ==========================================
// METODE LAMA (CREATE, SUBMIT, VERIFY, ETC)
// ==========================================

func (s *AchievementService) CreateAchievement(c *fiber.Ctx) error {
	var req modelMongo.Achievement
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid Input"})
	}

	userID := c.Locals("user_id").(string)
	studentID, err := s.RepoPG.GetStudentIDByUserID(userID)
	if err != nil {
		return c.Status(403).JSON(fiber.Map{"error": "User is not registered as a student"})
	}

	req.StudentID = studentID
	req.CreatedAt = time.Now()
	req.UpdatedAt = time.Now()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	mongoID, err := s.RepoMongo.InsertAchievement(ctx, req)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to save details to MongoDB"})
	}

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

func (s *AchievementService) SubmitForVerification(c *fiber.Ctx) error {
	id := c.Params("id")
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

func (s *AchievementService) DeleteAchievement(c *fiber.Ctx) error {
	id := c.Params("id")
	ref, err := s.RepoPG.GetReferenceByID(id)
	if err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "Achievement not found"})
	}

	if ref.Status != "draft" {
		return c.Status(400).JSON(fiber.Map{"error": "Only draft achievements can be deleted"})
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := s.RepoMongo.SoftDeleteAchievement(ctx, ref.MongoAchievementID); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to delete details"})
	}

	if err := s.RepoPG.UpdateStatus(id, "deleted"); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to update status"})
	}

	return c.JSON(fiber.Map{"message": "Achievement deleted successfully"})
}

func (s *AchievementService) GetAdviseeAchievements(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(string)
	refs, err := s.RepoPG.GetAchievementsByAdvisorID(userID)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to fetch achievements"})
	}

	var results []map[string]interface{}
	ctx := context.Background()

	for _, ref := range refs {
		detail, _ := s.RepoMongo.FindAchievementByID(ctx, ref.MongoAchievementID)
		item := map[string]interface{}{
			"ref_id":     ref.ID,
			"status":     ref.Status,
			"created_at": ref.CreatedAt,
			"detail":     detail,
		}
		results = append(results, item)
	}

	return c.JSON(fiber.Map{"data": results})
}

func (s *AchievementService) VerifyAchievement(c *fiber.Ctx) error {
	id := c.Params("id")
	userID := c.Locals("user_id").(string)

	if err := s.RepoPG.UpdateVerification(id, "verified", userID, nil); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to verify achievement"})
	}

	return c.JSON(fiber.Map{"message": "Achievement verified successfully"})
}

func (s *AchievementService) RejectAchievement(c *fiber.Ctx) error {
	id := c.Params("id")
	userID := c.Locals("user_id").(string)

	type RejectReq struct {
		Note string `json:"note"`
	}
	var req RejectReq
	if err := c.BodyParser(&req); err != nil || req.Note == "" {
		return c.Status(400).JSON(fiber.Map{"error": "Rejection note is required"})
	}

	if err := s.RepoPG.UpdateVerification(id, "rejected", userID, &req.Note); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to reject achievement"})
	}

	return c.JSON(fiber.Map{"message": "Achievement rejected"})
}