package postgres

import (
	"database/sql"
)

type IAcademicRepoPG interface {
	GetAllStudents() ([]map[string]interface{}, error)
	GetStudentByID(id string) (map[string]interface{}, error)
	UpdateStudentAdvisor(studentID, advisorID string) error
	
	GetAllLecturers() ([]map[string]interface{}, error)
	GetLecturerAdvisees(lecturerID string) ([]map[string]interface{}, error)
}

type AcademicRepoPG struct {
	DB *sql.DB
}

func NewAcademicRepoPG(db *sql.DB) IAcademicRepoPG {
	return &AcademicRepoPG{DB: db}
}

func (r *AcademicRepoPG) GetAllStudents() ([]map[string]interface{}, error) {
	// Query JOIN users untuk ambil nama lengkap
	query := `SELECT s.id, s.student_id, u.full_name, s.program_study 
			  FROM students s 
			  JOIN users u ON s.user_id = u.id`
	
	rows, err := r.DB.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	
	var results []map[string]interface{}
	for rows.Next() {
		var id, nim, name, prodi string
		if err := rows.Scan(&id, &nim, &name, &prodi); err != nil {
			return nil, err
		}
		results = append(results, map[string]interface{}{
			"id": id, "nim": nim, "name": name, "prodi": prodi,
		})
	}
	return results, nil
}

func (r *AcademicRepoPG) GetStudentByID(id string) (map[string]interface{}, error) {
	query := `SELECT s.id, s.student_id, u.full_name, s.program_study, s.advisor_id 
			  FROM students s 
			  JOIN users u ON s.user_id = u.id 
			  WHERE s.id = $1`
	
	var sid, nim, name, prodi string
	var advisorID sql.NullString // Handle null advisor

	err := r.DB.QueryRow(query, id).Scan(&sid, &nim, &name, &prodi, &advisorID)
	if err != nil {
		return nil, err
	}
	
	return map[string]interface{}{
		"id": sid, "nim": nim, "name": name, "prodi": prodi, "advisor_id": advisorID.String,
	}, nil
}

func (r *AcademicRepoPG) UpdateStudentAdvisor(studentID, advisorID string) error {
	_, err := r.DB.Exec("UPDATE students SET advisor_id = $1 WHERE id = $2", advisorID, studentID)
	return err
}

func (r *AcademicRepoPG) GetAllLecturers() ([]map[string]interface{}, error) {
	query := `SELECT l.id, l.lecturer_id, u.full_name, l.department 
			  FROM lecturers l 
			  JOIN users u ON l.user_id = u.id`
	
	rows, err := r.DB.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	
	var results []map[string]interface{}
	for rows.Next() {
		var id, nip, name, dept string
		if err := rows.Scan(&id, &nip, &name, &dept); err != nil {
			return nil, err
		}
		results = append(results, map[string]interface{}{
			"id": id, "nip": nip, "name": name, "department": dept,
		})
	}
	return results, nil
}

func (r *AcademicRepoPG) GetLecturerAdvisees(lecturerID string) ([]map[string]interface{}, error) {
	query := `SELECT s.id, s.student_id, u.full_name 
			  FROM students s 
			  JOIN users u ON s.user_id = u.id 
			  WHERE s.advisor_id = $1`
	
	rows, err := r.DB.Query(query, lecturerID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	
	var results []map[string]interface{}
	for rows.Next() {
		var id, nim, name string
		if err := rows.Scan(&id, &nim, &name); err != nil {
			return nil, err
		}
		results = append(results, map[string]interface{}{
			"id": id, "nim": nim, "name": name,
		})
	}
	return results, nil
}