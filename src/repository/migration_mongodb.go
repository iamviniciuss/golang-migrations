package repository

import (
	"context"

	entity "github.com/iamviniciuss/golang-migrations/src/dto"
	"github.com/pkg/errors"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

const defaultMigrationsCollection = "migrations"

const AllAvailable = -1

type MigrationRepositoryMongo struct {
	db *mongo.Database
}

func NewMigrationRepositoryMongo(db *mongo.Database) *MigrationRepositoryMongo {
	return &MigrationRepositoryMongo{db}
}

func (erm *MigrationRepositoryMongo) Insert(rec *entity.VersionRecord) (*entity.VersionRecord, error) {
	_, err := erm.db.Collection(defaultMigrationsCollection).InsertOne(context.TODO(), rec)
	if err != nil {
		return rec, err
	}

	return rec, nil
}

func (erm *MigrationRepositoryMongo) FindOne(reccord *entity.VersionRecord) (*entity.VersionRecord, error) {

	filter := bson.M{"version": reccord.Version, "description": reccord.Description, "type": reccord.Type}
	sort := bson.D{bson.E{Key: "_id", Value: -1}}
	options := options.FindOne().SetSort(sort)

	result := erm.db.Collection(defaultMigrationsCollection).FindOne(context.TODO(), filter, options)
	err := result.Err()
	switch {
	case err == mongo.ErrNoDocuments:
		return nil, err
	case err != nil:
		return nil, err
	}

	var output entity.VersionRecord
	if err := result.Decode(&output); err != nil {
		return nil, err
	}

	return &output, nil
}

func (erm *MigrationRepositoryMongo) isCollectionExist(name string) (isExist bool, err error) {
	collections, err := erm.getCollections()
	if err != nil {
		return false, err
	}

	for _, c := range collections {
		if name == c.Name {
			return true, nil
		}
	}
	return false, nil
}

func (erm *MigrationRepositoryMongo) CreateCollectionIfNotExists(name string) error {
	exist, err := erm.isCollectionExist(name)
	if err != nil {
		return err
	}
	if exist {
		return nil
	}

	command := bson.D{bson.E{Key: "create", Value: name}}
	err = erm.db.RunCommand(context.TODO(), command).Err()
	if err != nil {
		return err
	}

	return nil
}

func (erm *MigrationRepositoryMongo) getCollections() (collections []entity.CollectionSpecification, err error) {
	filter := bson.D{bson.E{Key: "type", Value: "collection"}}
	options := options.ListCollections().SetNameOnly(true)

	cursor, err := erm.db.ListCollections(context.Background(), filter, options)
	if err != nil {
		return nil, err
	}

	if cursor != nil {
		defer func(cursor *mongo.Cursor) {
			curErr := cursor.Close(context.TODO())
			if curErr != nil {
				if err != nil {
					err = errors.Wrapf(curErr, "migrate: get collection failed: %s", err.Error())
				} else {
					err = curErr
				}
			}
		}(cursor)
	}

	for cursor.Next(context.TODO()) {
		var collection entity.CollectionSpecification

		err := cursor.Decode(&collection)
		if err != nil {
			return nil, err
		}

		collections = append(collections, collection)
	}

	if err := cursor.Err(); err != nil {
		return nil, err
	}

	return
}
