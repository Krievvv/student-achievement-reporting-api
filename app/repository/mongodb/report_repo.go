package mongodb

import (
	"context"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

type IReportRepoMongo interface {
	GetStatsByType(ctx context.Context) ([]bson.M, error)
}

type ReportRepoMongo struct {
	Collection *mongo.Collection
}

func NewReportRepoMongo(db *mongo.Database) IReportRepoMongo {
	return &ReportRepoMongo{Collection: db.Collection("achievements")}
}

// Menghitung jumlah prestasi per Tipe (contoh: Kompetisi: 5, Organisasi: 2)
func (r *ReportRepoMongo) GetStatsByType(ctx context.Context) ([]bson.M, error) {
	pipeline := mongo.Pipeline{
		{{Key: "$group", Value: bson.D{
			{Key: "_id", Value: "$achievementType"},
			{Key: "count", Value: bson.D{{Key: "$sum", Value: 1}}},
		}}},
	}

	cursor, err := r.Collection.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, err
	}

	var results []bson.M
	if err = cursor.All(ctx, &results); err != nil {
		return nil, err
	}
	return results, nil
}