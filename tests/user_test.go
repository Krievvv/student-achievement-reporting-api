package tests

import (
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

// Helper khusus untuk Admin
func setupAppWithAdminAuth(handler fiber.Handler) *fiber.App {
	app := fiber.New()
	app.Use(func(c *fiber.Ctx) error {
		c.Locals("user_id", "admin-uuid-001")
		c.Locals("role", "Admin")
		return c.Next()
	})
	return app
}

func TestCreateUser_Success(t *testing.T) {
	mockUser := new(mocks.UserRepo)
	mockPG := new(mocks.AchievementRepoPG)
	mockMongo := new(mocks.AchievementRepoMongo)

	// Inject ke Admin Service
	svc := service.NewAdminService(mockUser, mockPG, mockMongo)

	// 2. Expectation
	// Skenario: Admin ingin membuat user dengan role "Dosen Wali"
	// Mock 1: Service akan mencari ID Role berdasarkan nama "Dosen Wali"
	mockUser.On("GetRoleIDByName", "Dosen Wali").Return("role-uuid-dosen", nil)

	// Mock 2: Service menyimpan data user baru ke database
	mockUser.On("CreateUser", mock.Anything).Return(nil)

	// 3. Request
	reqBody := map[string]interface{}{
		"username":  "dosen_baru",
		"password":  "123456",
		"email":     "dosen@univ.ac.id",
		"full_name": "Budi Dosen",
		"role_name": "Dosen Wali", // Service akan convert ini jadi ID lewat GetRoleIDByName
	}
	bodyBytes, _ := json.Marshal(reqBody)

	// Setup App sebagai Admin
	app := setupAppWithAdminAuth(svc.CreateUser)
	app.Post("/users", svc.CreateUser)

	req := httptest.NewRequest("POST", "/users", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")

	// 4. Execute
	resp, _ := app.Test(req)

	// 5. Assert (Harapannya 201 Created)
	assert.Equal(t, 201, resp.StatusCode)
}