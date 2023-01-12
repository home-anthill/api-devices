package test_utils

import (
	"context"
	. "github.com/onsi/gomega"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

func DropAllCollections(ctx context.Context, collProfiles, collHomes, collDevices *mongo.Collection) {
	var err error
	err = collProfiles.Drop(ctx)
	Expect(err).ShouldNot(HaveOccurred())
	err = collHomes.Drop(ctx)
	Expect(err).ShouldNot(HaveOccurred())
	err = collDevices.Drop(ctx)
	Expect(err).ShouldNot(HaveOccurred())
}

func FindAll[T interface{}](ctx context.Context, collection *mongo.Collection) ([]T, error) {
	cur, err := collection.Find(ctx, bson.M{})
	if err != nil {
		return []T{}, err
	}
	defer cur.Close(ctx)
	result := make([]T, 0)
	for cur.Next(ctx) {
		var res T
		cur.Decode(&res)
		result = append(result, res)
	}
	return result, nil
}

func FindOneById[T interface{}](ctx context.Context, collection *mongo.Collection, id primitive.ObjectID) (T, error) {
	var model T
	err := collection.FindOne(ctx, bson.M{
		"_id": id,
	}).Decode(&model)
	return model, err
}

func InsertOne(ctx context.Context, collection *mongo.Collection, obj interface{}) error {
	_, err := collection.InsertOne(ctx, obj)
	return err
}
