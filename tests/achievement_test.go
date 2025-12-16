package tests

import (
	"be_uas/app/model/mongodb"
	"be_uas/app/model/postgres"
	"be_uas/app/service"
	"be_uas/tests/mocks"
	"bytes"
	"encoding/json"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// Helper untuk menyuntikkan data User login (seolah-olah middleware JWT sudah jalan)
func setupAppWithAuth(handler fiber.Handler) *fiber.App {
	app := fiber.New()
	// Middleware Mocking: Inject user_id dan role
	app.Use(func(c *fiber.Ctx) error {
		c.Locals("user_id", "user-123")
		c.Locals("role", "Mahasiswa")
		return c.Next()
	})
	return app
}

func TestCreateAchievement_Success(t *testing.T) {
	// 1. Setup Mocks
	mockPG := new(mocks.AchievementRepoPG)
	mockMongo := new(mocks.AchievementRepoMongo)

	// Inject kedua mock ke Service
	svc := service.NewAchievementService(mockPG, mockMongo)

	// 2. Expectation
	// Mock Mongo Insert
	mockPG.On("GetStudentIDByUserID", "user-123").Return("student-uuid-999", nil)
	mockMongo.On("InsertAchievement", mock.Anything, mock.Anything).Return("mongo-id-999", nil)
	// Mock Postgres Create Reference
	mockPG.On("CreateReference", mock.Anything).Return(nil)

	// 3. Request
	reqBody := mongodb.Achievement{
		Title:           "Juara 1 Lomba",
		AchievementType: "competition",
		Points:          100,
	}
	bodyBytes, _ := json.Marshal(reqBody)

	app := setupAppWithAuth(svc.CreateAchievement) 
	app.Post("/achievements", svc.CreateAchievement)

	req := httptest.NewRequest("POST", "/achievements", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")

	// 4. Execute
	resp, _ := app.Test(req)

	// 5. Assert (201 Created)
	assert.Equal(t, 201, resp.StatusCode)
}

func TestSubmitAchievement_Success(t *testing.T) {
	mockPG := new(mocks.AchievementRepoPG)
	mockMongo := new(mocks.AchievementRepoMongo)
	svc := service.NewAchievementService(mockPG, mockMongo)

	refID := "ref-uuid-abc"

	// 1. Expectation
	// Mock Get Data (Harus status 'draft')
	mockPG.On("GetReferenceByID", refID).Return(&postgres.AchievementReference{
		ID: refID, Status: "draft",
	}, nil)

	// Mock Update Status ke 'submitted'
	mockPG.On("UpdateStatus", refID, "submitted").Return(nil)

	// 2. Request
	app := setupAppWithAuth(svc.SubmitAchievement)
	app.Post("/achievements/:id/submit", svc.SubmitAchievement)

	req := httptest.NewRequest("POST", "/achievements/"+refID+"/submit", nil)

	// 3. Execute
	resp, _ := app.Test(req)

	// 4. Assert (200 OK)
	assert.Equal(t, 200, resp.StatusCode)
}

func setupAppWithDosenAuth(handler fiber.Handler) *fiber.App {
	app := fiber.New()
	app.Use(func(c *fiber.Ctx) error {
		c.Locals("user_id", "dosen-uuid-123") 
		c.Locals("role", "Dosen Wali")        
		return c.Next()
	})
	return app
}

func TestVerifyAchievement_Success(t *testing.T) {
	mockPG := new(mocks.AchievementRepoPG)
	mockMongo := new(mocks.AchievementRepoMongo)
	svc := service.NewAchievementService(mockPG, mockMongo)

	refID := "ref-verify-123"
	dosenID := "dosen-uuid-123"

	// 1. Expectation
	// Mock: Dosen memverifikasi (Status berubah jadi 'verified')
	// Parameter order: id, status, verifiedBy, rejectionNote
	mockPG.On("UpdateVerification", refID, "verified", dosenID, (*string)(nil)).Return(nil)

	// 2. Request
	app := setupAppWithDosenAuth(svc.VerifyAchievement)
	app.Post("/achievements/:id/verify", svc.VerifyAchievement)

	req := httptest.NewRequest("POST", "/achievements/"+refID+"/verify", nil)

	// 3. Execute
	resp, _ := app.Test(req)

	// 4. Assert
	assert.Equal(t, 200, resp.StatusCode)
}

func TestRejectAchievement_Success(t *testing.T) {
	mockPG := new(mocks.AchievementRepoPG)
	mockMongo := new(mocks.AchievementRepoMongo)
	svc := service.NewAchievementService(mockPG, mockMongo)

	refID := "ref-reject-123"
	dosenID := "dosen-uuid-123"
	catatan := "Data kurang lengkap"

	// 1. Expectation
	// Mock: Dosen menolak (Status berubah jadi 'rejected', ada catatan)
	// Parameter order: id, status, verifiedBy, rejectionNote
	mockPG.On("UpdateVerification", refID, "rejected", dosenID, &catatan).Return(nil)

	// 2. Request Body (Reject butuh alasan/note)
	reqBody := map[string]string{"note": catatan}
	bodyBytes, _ := json.Marshal(reqBody)

	app := setupAppWithDosenAuth(svc.RejectAchievement)
	app.Post("/achievements/:id/reject", svc.RejectAchievement)

	req := httptest.NewRequest("POST", "/achievements/"+refID+"/reject", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")

	// 3. Execute
	resp, _ := app.Test(req)

	// 4. Assert
	assert.Equal(t, 200, resp.StatusCode)
}
