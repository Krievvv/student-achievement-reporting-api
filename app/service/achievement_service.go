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

func (s *AchievementService) GetAllAchievements(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(string)
	
	// Cari Student ID dari User ID
	studentID, err := s.RepoPG.GetStudentIDByUserID(userID)
	if err != nil {
		return c.Status(403).JSON(fiber.Map{"error": "User is not a student"})
	}

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

func (s *AchievementService) GetAchievementByID(c *fiber.Ctx) error {
	id := c.Params("id")
	
	ref, err := s.RepoPG.GetReferenceByID(id)
	if err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "Achievement not found"})
	}

	detail, err := s.RepoMongo.FindAchievementByID(context.Background(), ref.MongoAchievementID)
	
	return c.JSON(fiber.Map{
		"data": fiber.Map{
			"ref":    ref,
			"detail": detail,
		},
	})
}

func (s *AchievementService) GetAchievementHistory(c *fiber.Ctx) error {
	id := c.Params("id")
	ref, err := s.RepoPG.GetReferenceByID(id)
	if err != nil { 
		return c.Status(404).JSON(fiber.Map{"error": "Not found"}) 
	}

	var history []map[string]interface{}

	// 1. Created
	history = append(history, map[string]interface{}{
		"status":    "draft",
		"timestamp": ref.CreatedAt,
		"note":      "Achievement created",
	})

	// 2. Submitted
	if ref.SubmittedAt != nil { 
		history = append(history, map[string]interface{}{
			"status":    "submitted",
			"timestamp": ref.SubmittedAt,
			"note":      "Submitted for verification",
		})
	}
	
	// 3. Final Status
	if ref.Status == "verified" && ref.VerifiedAt != nil {
		history = append(history, map[string]interface{}{
			"status":    "verified",
			"timestamp": ref.VerifiedAt,
			"note":      "Verified by Lecturer",
		})
	} else if ref.Status == "rejected" && ref.VerifiedAt != nil {
		history = append(history, map[string]interface{}{
			"status":    "rejected",
			"timestamp": ref.VerifiedAt,
			"note":      "Rejected: " + getString(ref.RejectionNote),
		})
	}

	return c.JSON(fiber.Map{"data": history})
}

func (s *AchievementService) CreateAchievement(c *fiber.Ctx) error {
	var req modelMongo.Achievement
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid Input"})
	}

	userRole := c.Locals("role").(string)
	userID := c.Locals("user_id").(string)
	var studentID string
	var err error

	if userRole == "Mahasiswa" {
		studentID, err = s.RepoPG.GetStudentIDByUserID(userID)
		if err != nil {
			return c.Status(403).JSON(fiber.Map{"error": "User is not registered as a student"})
		}
	} else {
		if req.StudentID == "" {
			return c.Status(400).JSON(fiber.Map{"error": "Admin/Dosen must provide studentId in body"})
		}
		studentID = req.StudentID
	}

	if req.Attachments == nil {
		req.Attachments = []modelMongo.Attachment{}
	}

	if req.Points == 0 {
		req.Points = 10 
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

// Update Achievement
func (s *AchievementService) UpdateAchievement(c *fiber.Ctx) error {
	id := c.Params("id")
	ref, err := s.RepoPG.GetReferenceByID(id)
	if err != nil { return c.Status(404).JSON(fiber.Map{"error": "Not found"}) }

	if ref.Status != "draft" {
		return c.Status(400).JSON(fiber.Map{"error": "Cannot update. Status is not draft"})
	}

	var req modelMongo.Achievement
	if err := c.BodyParser(&req); err != nil { return c.Status(400).JSON(fiber.Map{"error": "Invalid Input"}) }

	req.UpdatedAt = time.Now()

	ctx := context.Background()
	if err := s.RepoMongo.UpdateAchievement(ctx, ref.MongoAchievementID, req); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to update content"})
	}

	return c.JSON(fiber.Map{"message": "Achievement updated"})
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

func (s *AchievementService) SubmitAchievement(c *fiber.Ctx) error {
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

func (s *AchievementService) UploadAttachment(c *fiber.Ctx) error {
	id := c.Params("id")
	ref, err := s.RepoPG.GetReferenceByID(id)
	if err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "Achievement not found"})
	}

	file, err := c.FormFile("file")
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "File is required"})
	}

	uniqueName := fmt.Sprintf("%s_%s", uuid.New().String(), file.Filename)
	filePath := fmt.Sprintf("./uploads/%s", uniqueName)
	
	if err := c.SaveFile(file, filePath); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to save file to server"})
	}

	attachment := modelMongo.Attachment{
		FileName:   file.Filename,
		FileURL:    "/uploads/" + uniqueName, 
		FileType:   file.Header.Get("Content-Type"),
		UploadedAt: time.Now(),
	}

	ctx := context.Background()
	if err := s.RepoMongo.AddAttachment(ctx, ref.MongoAchievementID, attachment); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to update database record", "details": err.Error()})
	}

	return c.JSON(fiber.Map{
		"message": "File uploaded and linked successfully",
		"data":    attachment,
	})
}

func (s *AchievementService) GetAdviseesAchievements(c *fiber.Ctx) error {
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

// Helper
func getString(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

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