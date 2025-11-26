package postgres

import (
	"time"
)

type AchievementReference struct {
	ID                 string    `json:"id"`
	StudentID          string    `json:"student_id"`
	MongoAchievementID string    `json:"mongo_achievement_id"`
	Status             string    `json:"status"` // draft, submitted, verified, rejected, deleted
	RejectionNote      *string   `json:"rejection_note"`
	VerifiedBy         *string   `json:"verified_by"`
	CreatedAt          time.Time `json:"created_at"`
	UpdatedAt          time.Time `json:"updated_at"`
}