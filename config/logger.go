package config

import (
	"os"
	"github.com/gofiber/fiber/v2/middleware/logger"
)

// Setup Logger Middleware Config
func NewLoggerConfig() logger.Config {
	// Pastikan folder logs ada
	if _, err := os.Stat("./logs"); os.IsNotExist(err) {
		os.Mkdir("./logs", 0755)
	}

	// Buka file log (append mode)
	file, _ := os.OpenFile("./logs/app.log", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)

	return logger.Config{
		Output: file, // Log ke file, bukan console
		Format: "[${time}] ${status} - ${method} ${path}\n",
	}
}