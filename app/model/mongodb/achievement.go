package mongodb

import (
	"time"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Achievement struct {
	ID              primitive.ObjectID     `bson:"_id,omitempty" json:"id"`
	StudentID       string                 `bson:"studentId" json:"studentId"` // UUID dari Postgres
	AchievementType string                 `bson:"achievementType" json:"achievementType"`
	Title           string                 `bson:"title" json:"title"`
	Description     string                 `bson:"description" json:"description"`
	Details         map[string]interface{} `bson:"details" json:"details"` // Field Dinamis
	Tags            []string               `bson:"tags" json:"tags"`
	CreatedAt       time.Time              `bson:"createdAt" json:"createdAt"`
	UpdatedAt       time.Time              `bson:"updatedAt" json:"updatedAt"`
	DeletedAt       *time.Time             `bson:"deletedAt,omitempty" json:"deletedAt,omitempty"` // Soft Delete
}