package test_utils

import (
	"context"
	. "github.com/onsi/gomega"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

func DropAllCollections(ctx context.Context, collectionACs *mongo.Collection) {
	var err error
	err = collectionACs.Drop(ctx)
	Expect(err).ShouldNot(HaveOccurred())
}

func FindOneById[T interface{}](ctx context.Context, collection *mongo.Collection, id primitive.ObjectID) (T, error) {
	var model T
	err := collection.FindOne(ctx, bson.M{
		"_id": id,
	}).Decode(&model)
	return model, err
}

func FindOneByKeyValue[T interface{}](ctx context.Context, collection *mongo.Collection, key, value string) (T, error) {
	var model T
	err := collection.FindOne(ctx, bson.M{
		key: value,
	}).Decode(&model)
	return model, err
}

func InsertOne(ctx context.Context, collection *mongo.Collection, obj interface{}) error {
	_, err := collection.InsertOne(ctx, obj)
	return err
}
