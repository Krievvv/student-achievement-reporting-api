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

// FR-011: Get Global Statistics
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