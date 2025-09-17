package seed

import (
	"context"

	"go.mongodb.org/mongo-driver/v2/mongo"
)

type OnlineReviewRepository interface {
	Insert(rec *OnlineReview) (*OnlineReview, error)
}

type OnlineReview struct {
	Id   string `json:"_id"`
	Name string `json:"name,omitempty"`
}

const defaultOnlineCollection = "online_reviews"

type OnlineRepositoryMongo struct {
	db *mongo.Database
}

func NewOnlineRepositoryMongo(db *mongo.Database) *OnlineRepositoryMongo {
	return &OnlineRepositoryMongo{db}
}

func (erm *OnlineRepositoryMongo) Insert(rec *OnlineReview) (*OnlineReview, error) {
	_, err := erm.db.Collection(defaultOnlineCollection).InsertOne(context.TODO(), rec)
	if err != nil {
		return rec, err
	}

	return rec, nil
}
