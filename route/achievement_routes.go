package route

import (
	"be_uas/app/service"
	"be_uas/middleware"
	"github.com/gofiber/fiber/v2"
)

func AchievementRoutes(group fiber.Router, achS *service.AchievementService) {
	ach := group.Group("/achievements", middleware.AuthRequired())

	// Public/Shared Access (Filtered by logic in service)
	ach.Get("/", achS.GetAllAchievements)
	ach.Get("/:id", achS.GetAchievementByID)
	ach.Get("/:id/history", achS.GetAchievementHistory)

	// Mahasiswa Actions
	ach.Post("/", achS.CreateAchievement)
	ach.Put("/:id", achS.UpdateAchievement)
	ach.Delete("/:id", achS.DeleteAchievement)
	ach.Post("/:id/submit", achS.SubmitAchievement)
	ach.Post("/:id/attachments", achS.UploadAttachment)

	// Dosen Wali Actions
	ach.Get("/advisees", middleware.PermissionCheck("Dosen Wali"), achS.GetAdviseesAchievements)
	ach.Post("/:id/verify", middleware.PermissionCheck("Dosen Wali"), achS.VerifyAchievement)
	ach.Post("/:id/reject", middleware.PermissionCheck("Dosen Wali"), achS.RejectAchievement)
}