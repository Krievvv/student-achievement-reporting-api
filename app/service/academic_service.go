package service

import (
	repoPG "be_uas/app/repository/postgres"
	"github.com/gofiber/fiber/v2"
)

type AcademicService struct {
	Repo repoPG.IAcademicRepoPG
}

func NewAcademicService(repo repoPG.IAcademicRepoPG) *AcademicService {
	return &AcademicService{Repo: repo}
}

func (s *AcademicService) GetAllStudents(c *fiber.Ctx) error {
	data, err := s.Repo.GetAllStudents()
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to fetch students"})
	}
	return c.JSON(fiber.Map{"data": data})
}

func (s *AcademicService) GetStudentByID(c *fiber.Ctx) error {
	data, err := s.Repo.GetStudentByID(c.Params("id"))
	if err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "Student not found"})
	}
	return c.JSON(fiber.Map{"data": data})
}

func (s *AcademicService) UpdateStudentAdvisor(c *fiber.Ctx) error {
	type Req struct {
		AdvisorID string `json:"advisor_id"`
	}
	var r Req
	if err := c.BodyParser(&r); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid input"})
	}
	
	if err := s.Repo.UpdateStudentAdvisor(c.Params("id"), r.AdvisorID); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to update advisor"})
	}
	return c.JSON(fiber.Map{"message": "Advisor updated successfully"})
}

func (s *AcademicService) GetAllLecturers(c *fiber.Ctx) error {
	data, err := s.Repo.GetAllLecturers()
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to fetch lecturers"})
	}
	return c.JSON(fiber.Map{"data": data})
}

func (s *AcademicService) GetLecturerAdvisees(c *fiber.Ctx) error {
	data, err := s.Repo.GetLecturerAdvisees(c.Params("id"))
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to fetch advisees"})
	}
	return c.JSON(fiber.Map{"data": data})
}