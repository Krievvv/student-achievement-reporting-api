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

// GetAllAchievements godoc
// @Summary      Get My Achievements
// @Description  Mahasiswa melihat daftar prestasi miliknya sendiri (Gabungan data PostgreSQL & MongoDB)
// @Tags         Achievements
// @Security     BearerAuth
// @Produce      json
// @Success      200  {object} map[string]interface{} "Data berisi array prestasi: id, status, detail (dari mongo)"
// @Failure      403  {object} map[string]interface{} "Error: User is not a student"
// @Failure      500  {object} map[string]interface{} "Error: Failed to fetch achievements"
// @Router       /achievements [get]
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

// GetAchievementByID godoc
// @Summary      Get Achievement Detail
// @Description  Melihat detail lengkap prestasi (Menggabungkan data Reference dari Postgres dan Detail Konten dari MongoDB)
// @Tags         Achievements
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        id   path      string  true  "Achievement Reference ID (UUID)"
// @Success      200  {object}  map[string]interface{} "Structure: {data: {ref: ReferenceObj, detail: MongoObj}}"
// @Failure      404  {object}  map[string]interface{}
// @Failure      500  {object}  map[string]interface{}
// @Router       /achievements/{id} [get]
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

// GetAchievementHistory godoc
// @Summary      Get Status History
// @Description  Melihat timeline riwayat perubahan status prestasi (Draft -> Submitted -> Verified/Rejected)
// @Tags         Achievements
// @Security     BearerAuth
// @Param        id   path      string  true  "Achievement Ref ID"
// @Produce      json
// @Success      200  {object} map[string]interface{} "Response format: {data: [{status, timestamp, note}]}"
// @Failure      404  {object} map[string]interface{}
// @Router       /achievements/{id}/history [get]
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

// CreateAchievement godoc
// @Summary      Create Achievement (Draft)
// @Description  Mahasiswa input prestasi, atau Admin inputkan untuk mahasiswa
// @Tags         Achievements
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        body body mongodb.Achievement true "Achievement Data"
// @Success      201  {object} map[string]interface{}
// @Router       /achievements [post]
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

// UpdateAchievement godoc
// @Summary      Update Achievement (Draft Only)
// @Description  Mengubah data konten prestasi. Hanya diizinkan jika status prestasi masih 'draft'.
// @Tags         Achievements
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        id    path      string               true  "Achievement Ref ID (UUID)"
// @Param        body  body      mongodb.Achievement  true  "Data Update (Title, Description, Details, etc)"
// @Success      200   {object}  map[string]interface{}  "message: Achievement updated"
// @Failure      400   {object}  map[string]interface{}  "Error: Invalid Input atau Status bukan Draft"
// @Failure      401   {object}  map[string]interface{}  "Error: Unauthorized"
// @Failure      404   {object}  map[string]interface{}  "Error: Not found"
// @Failure      500   {object}  map[string]interface{}  "Error: Internal Server Error"
// @Router       /achievements/{id} [put]
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

// DeleteAchievement godoc
// @Summary      Delete Draft Achievement
// @Description  Menghapus prestasi secara soft-delete. Hanya bisa dilakukan jika status masih 'draft'.
// @Tags         Achievements
// @Security     BearerAuth
// @Param        id   path      string  true  "Achievement Ref ID"
// @Produce      json
// @Success      200  {object} map[string]interface{}
// @Failure      400  {object} map[string]interface{} "Error: Only draft can be deleted"
// @Failure      401  {object} map[string]interface{} "Error: Unauthorized"
// @Failure      404  {object} map[string]interface{} "Error: Not found"
// @Failure      500  {object} map[string]interface{} "Error: Server failure"
// @Router       /achievements/{id} [delete]
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

// SubmitAchievement godoc
// @Summary      Submit to Advisor
// @Description  Mengubah status prestasi dari 'draft' menjadi 'submitted' untuk diverifikasi Dosen Wali.
// @Tags         Achievements
// @Security     BearerAuth
// @Param        id   path      string  true  "Achievement Ref ID"
// @Produce      json
// @Success      200  {object} map[string]interface{}
// @Failure      400  {object} map[string]interface{}
// @Failure      404  {object} map[string]interface{}
// @Failure      500  {object} map[string]interface{}
// @Router       /achievements/{id}/submit [post]
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

// UploadAttachment godoc
// @Summary      Upload File Bukti
// @Description  Upload gambar/pdf untuk prestasi
// @Tags         Achievements
// @Security     BearerAuth
// @Accept       multipart/form-data
// @Produce      json
// @Param        id   path      string  true  "Achievement Ref ID"
// @Param        file formData  file    true  "File Bukti"
// @Success      200  {object} map[string]interface{}
// @Router       /achievements/{id}/attachments [post]
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

// GetAdviseesAchievements godoc
// @Summary      Get Advisee Achievements (Dosen Wali)
// @Description  Dosen Wali melihat daftar prestasi yang diajukan oleh mahasiswa bimbingannya untuk diverifikasi.
// @Tags         Achievements (Dosen)
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Success      200  {object} map[string]interface{} "Response format: { data: [ {ref_id, status, detail...} ] }"
// @Failure      500  {object} map[string]interface{}
// @Router       /achievements/advisees [get]
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

// VerifyAchievement godoc
// @Summary      Verify Achievement (Dosen)
// @Description  Dosen Wali menyetujui prestasi mahasiswa bimbingannya (Status berubah menjadi 'verified')
// @Tags         Achievements (Dosen)
// @Security     BearerAuth
// @Param        id   path      string  true  "Achievement Ref ID"
// @Produce      json
// @Success      200  {object} map[string]interface{}
// @Failure      500  {object} map[string]interface{}
// @Router       /achievements/{id}/verify [post]
func (s *AchievementService) VerifyAchievement(c *fiber.Ctx) error {
	id := c.Params("id")
	userID := c.Locals("user_id").(string)

	if err := s.RepoPG.UpdateVerification(id, "verified", userID, nil); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to verify achievement"})
	}

	return c.JSON(fiber.Map{"message": "Achievement verified successfully"})
}

// RejectAchievement godoc
// @Summary      Reject Achievement (Dosen)
// @Description  Dosen Wali menolak prestasi mahasiswa dengan memberikan catatan (Note) wajib.
// @Tags         Achievements (Dosen)
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        id    path      string             true  "Achievement Ref ID (UUID)"
// @Param        body  body      map[string]string  true  "Payload: { 'note': 'Alasan penolakan...' }"
// @Success      200   {object}  map[string]interface{}
// @Failure      400   {object}  map[string]interface{}
// @Failure      500   {object}  map[string]interface{}
// @Router       /achievements/{id}/reject [post]
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

// GetStudentAchievements godoc
// @Summary      Get Student's Achievements
// @Description  Melihat daftar prestasi lengkap (Gabungan data PostgreSQL & MongoDB) milik mahasiswa tertentu berdasarkan Student ID.
// @Tags         Academic
// @Security     BearerAuth
// @Param        id   path      string  true  "Student UUID"
// @Produce      json
// @Success      200  {object} map[string]interface{} "Format: { data: [{ id, status, created_at, detail: {...} }] }"
// @Failure      500  {object} map[string]interface{}
// @Router       /students/{id}/achievements [get]
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