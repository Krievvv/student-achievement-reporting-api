package mongodb

import (
	"be_uas/app/model/mongodb"
	"context"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type IAchievementRepoMongo interface {
	InsertAchievement(ctx context.Context, data mongodb.Achievement) (string, error)
	SoftDeleteAchievement(ctx context.Context, hexID string) error
	FindAchievementByID(ctx context.Context, hexID string) (*mongodb.Achievement, error)
	UpdateAchievement(ctx context.Context, hexID string, data mongodb.Achievement) error
	AddAttachment(ctx context.Context, hexID string, attachment mongodb.Attachment) error
}

type AchievementRepoMongo struct {
	Collection *mongo.Collection
}

func NewAchievementRepoMongo(db *mongo.Database) IAchievementRepoMongo {
	return &AchievementRepoMongo{Collection: db.Collection("achievements")}
}

func (r *AchievementRepoMongo) InsertAchievement(ctx context.Context, data mongodb.Achievement) (string, error) {
	res, err := r.Collection.InsertOne(ctx, data)
	if err != nil {
		return "", err
	}
	return res.InsertedID.(primitive.ObjectID).Hex(), nil
}

func (r *AchievementRepoMongo) SoftDeleteAchievement(ctx context.Context, hexID string) error {
	objID, _ := primitive.ObjectIDFromHex(hexID)
	update := bson.M{"$set": bson.M{"deletedAt": time.Now()}}
	_, err := r.Collection.UpdateOne(ctx, bson.M{"_id": objID}, update)
	return err
}

func (r *AchievementRepoMongo) FindAchievementByID(ctx context.Context, hexID string) (*mongodb.Achievement, error) {
	objID, _ := primitive.ObjectIDFromHex(hexID)
	var achievement mongodb.Achievement
	
	err := r.Collection.FindOne(ctx, bson.M{"_id": objID}).Decode(&achievement)
	if err != nil {
		return nil, err
	}
	return &achievement, nil
}

func (r *AchievementRepoMongo) UpdateAchievement(ctx context.Context, hexID string, data mongodb.Achievement) error {
    objID, _ := primitive.ObjectIDFromHex(hexID)
    
    // Kita update field-field content saja
    update := bson.M{"$set": bson.M{
        "achievementType": data.AchievementType,
        "title":           data.Title,
        "description":     data.Description,
        "details":         data.Details,
        "tags":            data.Tags,
        "points":          data.Points,
        "updatedAt":       data.UpdatedAt,
    }}
    
    _, err := r.Collection.UpdateOne(ctx, bson.M{"_id": objID}, update)
    return err
}

func (r *AchievementRepoMongo) AddAttachment(ctx context.Context, hexID string, attachment mongodb.Attachment) error {
	objID, _ := primitive.ObjectIDFromHex(hexID)

	// Gunakan operator $push untuk menambah item ke array 'attachments'
	update := bson.M{
		"$push": bson.M{
			"attachments": attachment,
		},
	}

	_, err := r.Collection.UpdateOne(ctx, bson.M{"_id": objID}, update)
	return err
}