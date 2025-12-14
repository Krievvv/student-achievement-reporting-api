package route

import (
	"be_uas/app/service"
	"be_uas/middleware"
	"github.com/gofiber/fiber/v2"
)

func AcademicRoutes(group fiber.Router, acadS *service.AcademicService, achS *service.AchievementService) {
	academic := group.Group("/", middleware.AuthRequired())

	// Students
	academic.Get("/students", acadS.GetAllStudents)
	academic.Get("/students/:id", acadS.GetStudentByID)
	academic.Get("/students/:id/achievements", achS.GetStudentAchievements)

	academic.Put("/students/:id/advisor", middleware.PermissionCheck("Admin"), acadS.UpdateStudentAdvisor)

	academic.Get("/lecturers", acadS.GetAllLecturers)
	academic.Get("/lecturers/:id/advisees", acadS.GetLecturerAdvisees)
}