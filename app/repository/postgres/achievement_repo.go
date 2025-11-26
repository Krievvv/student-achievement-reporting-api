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
	query := `SELECT id, student_id, mongo_achievement_id, status FROM achievement_references WHERE id = $1`
	err := r.DB.QueryRow(query, id).Scan(&ref.ID, &ref.StudentID, &ref.MongoAchievementID, &ref.Status)
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