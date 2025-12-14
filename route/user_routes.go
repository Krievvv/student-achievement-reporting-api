package route

import (
	"be_uas/app/service"
	"be_uas/middleware"
	"github.com/gofiber/fiber/v2"
)

func UserRoutes(group fiber.Router, adminS *service.AdminService) {
	users := group.Group("/users", middleware.AuthRequired(), middleware.PermissionCheck("Admin"))
	users.Get("/", adminS.ListUsers)        
	users.Post("/", adminS.CreateUser)
	users.Get("/:id", adminS.GetUserDetail) 
	users.Put("/:id", adminS.UpdateUser)
	users.Put("/:id/role", adminS.UpdateRole) 
	users.Delete("/:id", adminS.DeleteUser)
}