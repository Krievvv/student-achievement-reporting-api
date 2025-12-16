package service

import (
	"context"
	repoMongo "be_uas/app/repository/mongodb"
	repoPG "be_uas/app/repository/postgres"
	
	"github.com/gofiber/fiber/v2"
)

type ReportService struct {
	ReportPG    repoPG.IReportRepoPG
	ReportMongo repoMongo.IReportRepoMongo
}

func NewReportService(pg repoPG.IReportRepoPG, mongo repoMongo.IReportRepoMongo) *ReportService {
	return &ReportService{
		ReportPG:    pg,
		ReportMongo: mongo,
	}
}

// Get Global Statistics

// GetStatistics godoc
// @Summary      Get Global Statistics
// @Description  Mendapatkan data statistik untuk Dashboard: total prestasi per tipe (untuk Pie Chart) dan Top 5 Mahasiswa (Leaderboard).
// @Tags         Reports
// @Security     BearerAuth
// @Produce      json
// @Success      200  {object} map[string]interface{} "Format: {message: string, data: {achievements_by_type: Array, top_students: Array}}"
// @Failure      500  {object} map[string]interface{}
// @Router       /reports/statistics [get]
func (s *ReportService) GetStatistics(c *fiber.Ctx) error {
	ctx := context.Background()

	// 1. Ambil Statistik Tipe dari MongoDB
	typeStats, err := s.ReportMongo.GetStatsByType(ctx)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to fetch type stats"})
	}

	// 2. Ambil Top 5 Mahasiswa dari PostgreSQL
	topStudents, err := s.ReportPG.GetTopStudents(5)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to fetch top students"})
	}

	// 3. Gabungkan Response
	return c.JSON(fiber.Map{
		"message": "Statistics fetched successfully",
		"data": fiber.Map{
			"achievements_by_type": typeStats,
			"top_students":         topStudents,
		},
	})
}

// GetStudentReport godoc
// @Summary      Get Student Statistics
// @Description  Melihat statistik prestasi spesifik untuk satu mahasiswa (Total Verified, Submitted, Rejected)
// @Tags         Reports
// @Security     BearerAuth
// @Param        id   path      string  true  "Student UUID"
// @Produce      json
// @Success      200  {object} map[string]interface{}
// @Failure      500  {object} map[string]interface{}
// @Router       /reports/student/{id} [get]
func (s *ReportService) GetStudentReport(c *fiber.Ctx) error {
    studentID := c.Params("id") // UUID dari tabel Students (bukan Users)

    stats, err := s.ReportPG.GetStudentStats(studentID)
    if err != nil {
        return c.Status(500).JSON(fiber.Map{"error": "Failed to get report"})
    }

    return c.JSON(fiber.Map{
        "student_id": studentID,
        "statistics": stats,
    })
}