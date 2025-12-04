package postgres

import (
	"database/sql"
)

type TopStudent struct {
	StudentName string `json:"student_name"`
	ProgramStudy string `json:"program_study"`
	TotalVerified int    `json:"total_verified"`
}

type IReportRepoPG interface {
	GetTopStudents(limit int) ([]TopStudent, error)
}

type ReportRepoPG struct {
	DB *sql.DB
}

func NewReportRepoPG(db *sql.DB) IReportRepoPG {
	return &ReportRepoPG{DB: db}
}

// Mengambil Top Mahasiswa berdasarkan jumlah prestasi verified
func (r *ReportRepoPG) GetTopStudents(limit int) ([]TopStudent, error) {
	query := `
		SELECT u.full_name, s.program_study, COUNT(ar.id) as total
		FROM achievement_references ar
		JOIN students s ON ar.student_id = s.id
		JOIN users u ON s.user_id = u.id
		WHERE ar.status = 'verified'
		GROUP BY u.full_name, s.program_study
		ORDER BY total DESC
		LIMIT $1
	`
	rows, err := r.DB.Query(query, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []TopStudent
	for rows.Next() {
		var t TopStudent
		if err := rows.Scan(&t.StudentName, &t.ProgramStudy, &t.TotalVerified); err != nil {
			return nil, err
		}
		results = append(results, t)
	}
	return results, nil
}