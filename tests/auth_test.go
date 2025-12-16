package tests

import (
	"be_uas/app/model/postgres"
	"be_uas/app/service"
	"be_uas/tests/mocks" // Import folder mocks yang baru dibuat
	"bytes"
	"encoding/json"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
	"golang.org/x/crypto/bcrypt"
)

func TestLogin_Success(t *testing.T) {
	// 1. Setup
	os.Setenv("JWT_SECRET", "test_secret")
	mockUserRepo := new(mocks.UserRepo)
	authService := service.NewAuthService(mockUserRepo) // Inject Mock

	// 2. Data Dummy
	password := "rahasia123"
	hashed, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	
	dummyUser := &postgres.User{
		ID:           "user-uuid-1",
		Username:     "mahasiswa_test",
		PasswordHash: string(hashed),
		RoleName:     "Mahasiswa",
		IsActive:     true,
	}

	// 3. Expectation (Mocking)
	// "Kalau ada yang minta user 'mahasiswa_test', kasih dummyUser ini ya"
	mockUserRepo.On("GetByUsername", "mahasiswa_test").Return(dummyUser, nil)

	// 4. Setup Fiber & Request
	app := fiber.New()
	app.Post("/auth/login", authService.Login)

	reqBody, _ := json.Marshal(map[string]string{
		"username": "mahasiswa_test",
		"password": password,
	})
	req := httptest.NewRequest("POST", "/auth/login", bytes.NewReader(reqBody))
	req.Header.Set("Content-Type", "application/json")

	// 5. Execute
	resp, _ := app.Test(req)

	// 6. Assert
	assert.Equal(t, 200, resp.StatusCode)
}

func TestLogin_WrongPassword(t *testing.T) {
	// 1. Setup
	mockUserRepo := new(mocks.UserRepo)
	authService := service.NewAuthService(mockUserRepo)

	// 2. Data Dummy (Password Asli: "rahasia123")
	password := "rahasia123"
	hashed, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	
	dummyUser := &postgres.User{
		ID:           "user-uuid-1",
		Username:     "mahasiswa_test",
		PasswordHash: string(hashed),
		IsActive:     true,
	}

	mockUserRepo.On("GetByUsername", "mahasiswa_test").Return(dummyUser, nil)

	// 3. Request (Password Input: "SALAH")
	app := fiber.New()
	app.Post("/auth/login", authService.Login)

	reqBody, _ := json.Marshal(map[string]string{
		"username": "mahasiswa_test",
		"password": "password_salah",
	})
	req := httptest.NewRequest("POST", "/auth/login", bytes.NewReader(reqBody))
	req.Header.Set("Content-Type", "application/json")

	// 4. Execute
	resp, _ := app.Test(req)

	// 5. Assert (Harus Error 401)
	assert.Equal(t, 401, resp.StatusCode)
}