package route

import (
	"be_uas/app/service"
	"be_uas/middleware"
	"github.com/gofiber/fiber/v2"
)

func AuthRoutes(group fiber.Router, authS *service.AuthService) {
	auth := group.Group("/auth")
	auth.Post("/login", authS.Login)
	auth.Post("/refresh", middleware.AuthRequired(), authS.RefreshToken)
	auth.Post("/logout", authS.Logout)
	auth.Get("/profile", middleware.AuthRequired(), authS.GetProfile)
}