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

// GetAllStudents godoc
// @Summary      Get All Students
// @Description  Mendapatkan daftar semua mahasiswa beserta data akademiknya
// @Tags         Academic
// @Security     BearerAuth
// @Produce      json
// @Success      200  {object} map[string]interface{} "Format: { data: [StudentObjects...] }"
// @Failure      500  {object} map[string]interface{}
// @Router       /students [get]
func (s *AcademicService) GetAllStudents(c *fiber.Ctx) error {
	data, err := s.Repo.GetAllStudents()
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to fetch students"})
	}
	return c.JSON(fiber.Map{"data": data})
}

// GetStudentByID godoc
// @Summary      Get Student Detail
// @Description  Mendapatkan data lengkap mahasiswa (termasuk program studi & advisor) berdasarkan ID
// @Tags         Academic
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id   path      string  true  "Student UUID"
// @Success      200  {object}  map[string]interface{}
// @Failure      404  {object}  map[string]interface{}
// @Router       /students/{id} [get]
func (s *AcademicService) GetStudentByID(c *fiber.Ctx) error {
	data, err := s.Repo.GetStudentByID(c.Params("id"))
	if err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "Student not found"})
	}
	return c.JSON(fiber.Map{"data": data})
}

// UpdateStudentAdvisor godoc
// @Summary      Update Student Advisor
// @Description  Admin menugaskan atau mengubah Dosen Wali untuk mahasiswa tertentu
// @Tags         Academic
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        id    path      string             true  "Student UUID"
// @Param        body  body      map[string]string  true  "Request Body: { \"advisor_id\": \"UUID_DOSEN\" }"
// @Success      200   {object}  map[string]interface{}
// @Failure      400   {object}  map[string]interface{}
// @Failure      500   {object}  map[string]interface{}
// @Router       /students/{id}/advisor [put]
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

// GetAllLecturers godoc
// @Summary      Get All Lecturers
// @Description  Mendapatkan daftar semua dosen yang terdaftar dalam sistem
// @Tags         Academic
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Success      200  {object} map[string]interface{} "Format: {data: [Array of Lecturers]}"
// @Failure      500  {object} map[string]interface{}
// @Router       /lecturers [get]
func (s *AcademicService) GetAllLecturers(c *fiber.Ctx) error {
	data, err := s.Repo.GetAllLecturers()
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to fetch lecturers"})
	}
	return c.JSON(fiber.Map{"data": data})
}

// GetLecturerAdvisees godoc
// @Summary      Get Lecturer's Advisees
// @Description  Melihat daftar mahasiswa yang dibimbing (perwalian) oleh dosen tertentu berdasarkan ID Dosen.
// @Tags         Academic
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        id   path      string  true  "Lecturer UUID"
// @Success      200  {object}  map[string]interface{} "Returns object {data: [Student Array]}"
// @Failure      500  {object}  map[string]interface{}
// @Router       /lecturers/{id}/advisees [get]
func (s *AcademicService) GetLecturerAdvisees(c *fiber.Ctx) error {
	data, err := s.Repo.GetLecturerAdvisees(c.Params("id"))
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to fetch advisees"})
	}
	return c.JSON(fiber.Map{"data": data})
}