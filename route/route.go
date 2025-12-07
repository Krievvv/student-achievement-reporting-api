package route

import (
    "be_uas/app/service"
    "be_uas/middleware"
    "github.com/gofiber/fiber/v2"
)

func SetupRoutes(app *fiber.App, 
    authService *service.AuthService, 
    achieveService *service.AchievementService, 
    adminService *service.AdminService, 
    reportService *service.ReportService,
    academicService *service.AcademicService) {

    api := app.Group("/api/v1")

    // AUTH
    auth := api.Group("/auth")
    auth.Post("/login", authService.Login)
    auth.Post("/seed", authService.SeedAdmin)
    auth.Post("/refresh", middleware.AuthRequired(), authService.RefreshToken) 
    auth.Post("/logout", authService.Logout) 
    auth.Get("/profile", middleware.AuthRequired(), authService.GetProfile)

    // USERS
    admin := api.Group("/users", middleware.AuthRequired(), middleware.PermissionCheck("Admin"))
    admin.Post("/", adminService.CreateUser)
    admin.Get("/", adminService.ListUsers)
    admin.Delete("/:id", adminService.DeleteUser)
    admin.Get("/:id", adminService.GetUserDetail) 
    admin.Put("/:id", adminService.UpdateUser) 
    admin.Put("/:id/role", adminService.UpdateRole)

    // ACHIEVEMENTS
    ach := api.Group("/achievements", middleware.AuthRequired())
    ach.Get("/", achieveService.GetMyAchievements)
    ach.Get("/:id", achieveService.GetAchievementDetail)
    ach.Post("/:id/attachments", achieveService.UploadAttachment)
    ach.Get("/:id/history", achieveService.GetHistory)
    ach.Post("/", achieveService.CreateAchievement)		
    ach.Put("/:id", middleware.PermissionCheck("Mahasiswa"), achieveService.UpdateAchievement) 
    ach.Post("/:id/submit", middleware.PermissionCheck("Mahasiswa"), achieveService.SubmitForVerification)
    ach.Delete("/:id", middleware.PermissionCheck("Mahasiswa"), achieveService.DeleteAchievement)
    ach.Get("/advisees", middleware.PermissionCheck("Dosen Wali"), achieveService.GetAdviseeAchievements)
    ach.Post("/:id/verify", middleware.PermissionCheck("Dosen Wali"), achieveService.VerifyAchievement)
    ach.Post("/:id/reject", middleware.PermissionCheck("Dosen Wali"), achieveService.RejectAchievement)

    // STUDENTS & LECTURERS
    std := api.Group("/students", middleware.AuthRequired())
    std.Get("/", academicService.GetAllStudents)
    std.Get("/:id", academicService.GetStudentByID)
    std.Get("/:id/achievements", achieveService.GetStudentAchievements)
    std.Put("/:id/advisor", middleware.PermissionCheck("Admin"), academicService.UpdateStudentAdvisor)

    lec := api.Group("/lecturers", middleware.AuthRequired())
    lec.Get("/", academicService.GetAllLecturers)
    lec.Get("/:id/advisees", academicService.GetLecturerAdvisees)

    // REPORTS
    reports := api.Group("/reports", middleware.AuthRequired())
    reports.Get("/statistics", reportService.GetStatistics)
    reports.Get("/student/:id", reportService.GetStudentReport)
}