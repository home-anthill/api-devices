package testutils

import (
	"context"

	"github.com/onsi/gomega"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

func DropAllCollections(ctx context.Context, collectionDevices *mongo.Collection) {
	err := collectionDevices.Drop(ctx)
	gomega.Expect(err).ShouldNot(gomega.HaveOccurred())
}

func FindOneById[T interface{}](ctx context.Context, collection *mongo.Collection, id bson.ObjectID) (T, error) {
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
