package postgres

import (
	"be_uas/app/model/postgres"
	"database/sql"
)

type IAchievementRepoPG interface {
	CreateReference(ref postgres.AchievementReference) error
	GetReferenceByID(id string) (*postgres.AchievementReference, error)
	UpdateStatus(id string, status string) error
	GetStudentIDByUserID(userID string) (string, error)
	GetAchievementsByAdvisorID(userID string) ([]postgres.AchievementReference, error)
	UpdateVerification(id string, status string, verifiedBy string, rejectionNote *string) error
	GetAllAchievements(limit, offset int) ([]postgres.AchievementReference, int, error)
	GetAchievementsByStudentID(studentID string) ([]postgres.AchievementReference, error)
}

type AchievementRepoPG struct {
	DB *sql.DB
}

func NewAchievementRepoPG(db *sql.DB) IAchievementRepoPG {
	return &AchievementRepoPG{DB: db}
}

func (r *AchievementRepoPG) GetStudentIDByUserID(userID string) (string, error) {
	var studentID string
	query := `SELECT id FROM students WHERE user_id = $1`
	err := r.DB.QueryRow(query, userID).Scan(&studentID)
	return studentID, err
}

func (r *AchievementRepoPG) CreateReference(ref postgres.AchievementReference) error {
	query := `INSERT INTO achievement_references (id, student_id, mongo_achievement_id, status, created_at, updated_at) 
              VALUES ($1, $2, $3, $4, $5, $6)`
	_, err := r.DB.Exec(query, ref.ID, ref.StudentID, ref.MongoAchievementID, ref.Status, ref.CreatedAt, ref.UpdatedAt)
	return err
}

func (r *AchievementRepoPG) GetReferenceByID(id string) (*postgres.AchievementReference, error) {
	ref := &postgres.AchievementReference{}
	query := `
		SELECT 
			id, student_id, mongo_achievement_id, status, 
			submitted_at, verified_at, verified_by, rejection_note, 
			created_at, updated_at 
		FROM achievement_references 
		WHERE id = $1
	`
	
	err := r.DB.QueryRow(query, id).Scan(
		&ref.ID, 
		&ref.StudentID, 
		&ref.MongoAchievementID, 
		&ref.Status,
		&ref.SubmittedAt,   
		&ref.VerifiedAt,    
		&ref.VerifiedBy, 
		&ref.RejectionNote,
		&ref.CreatedAt, 
		&ref.UpdatedAt,
	)
	
	if err != nil {
		return nil, err
	}
	return ref, nil
}

func (r *AchievementRepoPG) UpdateStatus(id string, status string) error {
	query := `UPDATE achievement_references SET status = $1, updated_at = NOW() WHERE id = $2`
	_, err := r.DB.Exec(query, status, id)
	return err
}

func (r *AchievementRepoPG) GetAchievementsByAdvisorID(userID string) ([]postgres.AchievementReference, error) {
	// Query: Cari Lecturer ID dari User ID -> Cari Student yang dibimbing -> Cari Prestasi
	query := `
		SELECT ar.id, ar.student_id, ar.mongo_achievement_id, ar.status, ar.created_at, ar.updated_at
		FROM achievement_references ar
		JOIN students s ON ar.student_id = s.id
		JOIN lecturers l ON s.advisor_id = l.id
		WHERE l.user_id = $1 AND ar.status != 'draft' AND ar.status != 'deleted'
		ORDER BY ar.created_at DESC
	`
	rows, err := r.DB.Query(query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var achievements []postgres.AchievementReference
	for rows.Next() {
		var ar postgres.AchievementReference
		if err := rows.Scan(&ar.ID, &ar.StudentID, &ar.MongoAchievementID, &ar.Status, &ar.CreatedAt, &ar.UpdatedAt); err != nil {
			return nil, err
		}
		achievements = append(achievements, ar)
	}
	return achievements, nil
}

// Update status Verify/Reject
func (r *AchievementRepoPG) UpdateVerification(id string, status string, verifiedBy string, rejectionNote *string) error {
	query := `
		UPDATE achievement_references 
		SET status = $1, verified_by = $2, rejection_note = $3, verified_at = NOW(), updated_at = NOW() 
		WHERE id = $4
	`
	_, err := r.DB.Exec(query, status, verifiedBy, rejectionNote, id)
	return err
}

func (r *AchievementRepoPG) GetAllAchievements(limit, offset int) ([]postgres.AchievementReference, int, error) {
	// Ambil Data
	query := `
		SELECT id, student_id, mongo_achievement_id, status, created_at, updated_at
		FROM achievement_references
		ORDER BY created_at DESC
		LIMIT $1 OFFSET $2
	`
	rows, err := r.DB.Query(query, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var achievements []postgres.AchievementReference
	for rows.Next() {
		var ar postgres.AchievementReference
		rows.Scan(&ar.ID, &ar.StudentID, &ar.MongoAchievementID, &ar.Status, &ar.CreatedAt, &ar.UpdatedAt)
		achievements = append(achievements, ar)
	}

	// Hitung Total Data (Untuk metadata pagination)
	var total int
	r.DB.QueryRow("SELECT COUNT(*) FROM achievement_references").Scan(&total)

	return achievements, total, nil
}

func (r *AchievementRepoPG) GetAchievementsByStudentID(studentID string) ([]postgres.AchievementReference, error) {
    query := `
        SELECT id, student_id, mongo_achievement_id, status, created_at, updated_at
        FROM achievement_references
        WHERE student_id = $1
        ORDER BY created_at DESC
    `
    rows, err := r.DB.Query(query, studentID)
    if err != nil {
        return nil, err
    }
    defer rows.Close()

    var achievements []postgres.AchievementReference
    for rows.Next() {
        var ar postgres.AchievementReference
        if err := rows.Scan(&ar.ID, &ar.StudentID, &ar.MongoAchievementID, &ar.Status, &ar.CreatedAt, &ar.UpdatedAt); err != nil {
            return nil, err
        }
        achievements = append(achievements, ar)
    }
    return achievements, nil
}