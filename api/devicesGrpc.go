package api

import (
	"api-devices/api/device"
	"api-devices/db"
	"api-devices/models"
	mqtt_client "api-devices/mqttclient"
	"context"
	"encoding/json"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.uber.org/zap"
)

const devicesTimeout = 5 * time.Second

// DevicesGrpc struct
type DevicesGrpc struct {
	device.UnimplementedDeviceServer
	client                *mongo.Client
	controllersCollection *mongo.Collection
	contextRef            context.Context
	logger                *zap.SugaredLogger
}

// NewDevicesGrpc function
func NewDevicesGrpc(ctx context.Context, logger *zap.SugaredLogger, client *mongo.Client) *DevicesGrpc {
	return &DevicesGrpc{
		client:                client,
		controllersCollection: db.GetCollections(client).Controllers,
		contextRef:            ctx,
		logger:                logger,
	}
}

// GetValue function
func (handler *DevicesGrpc) GetValue(ctx context.Context, in *device.GetValueRequest) (*device.GetValueResponse, error) {
	handler.logger.Infof("gRPC - GetValue - Called with in: %#v", in)

	var controller models.Controller
	err := handler.controllersCollection.FindOne(handler.contextRef, bson.M{
		// profile info
		"apiToken": in.ApiToken,
		// device info
		"deviceUuid": in.DeviceUuid,
		"mac":        in.Mac,
		// feature info
		"featureUuid": in.FeatureUuid,
		"featureName": in.FeatureName,
	}).Decode(&controller)
	if err != nil {
		handler.logger.Error("gRPC - GetValue - Cannot get device with specified mac " + in.Mac)
		return nil, fmt.Errorf("cannot find controller with mac %s", in.Mac)
	}

	statusResponse := device.GetValueResponse{
		FeatureUuid: controller.FeatureUUID,
		FeatureName: controller.FeatureName,
		Value:       controller.Status.Value,
		CreatedAt:   controller.Status.CreatedAt.UnixMilli(),
		ModifiedAt:  controller.Status.ModifiedAt.UnixMilli(),
	}
	return &statusResponse, err
}

// SetValues function
func (handler *DevicesGrpc) SetValues(ctx context.Context, in *device.SetValuesRequest) (*device.SetValueResponse, error) {
	handler.logger.Infof("gRPC - SetValue - Called with in: %#v", in)

	var mqttValues []models.MqttFeatureValue

	for _, value := range in.FeatureValues {
		var controller models.Controller
		err := handler.controllersCollection.FindOne(handler.contextRef, bson.M{
			// profile info
			"apiToken": in.ApiToken,
			// device info
			"deviceUuid": in.DeviceUuid,
			"mac":        in.Mac,
			// feature info
			"featureUuid": value.FeatureUuid,
			"featureName": value.FeatureName,
		}).Decode(&controller)
		if err != nil {
			handler.logger.Error("gRPC - SetValue - Cannot find device with specified mac " + in.Mac)
			return nil, fmt.Errorf("cannot find controller with mac %s", in.Mac)
		}

		updatedStatue := models.Status{
			Value:      value.Value,
			CreatedAt:  time.Now(),
			ModifiedAt: time.Now(),
		}

		handler.logger.Debugf("gRPC - SetValue - updatedStatue %#v ", updatedStatue)

		updateResult, err := handler.controllersCollection.UpdateOne(handler.contextRef, bson.M{
			// profile info
			"apiToken": in.ApiToken,
			// device info
			"deviceUuid": in.DeviceUuid,
			"mac":        in.Mac,
			// feature info
			"featureUuid": value.FeatureUuid,
			"featureName": value.FeatureName,
		}, bson.M{
			"$set": bson.M{
				"status":     updatedStatue,
				"modifiedAt": time.Now(),
			},
		})

		if err != nil {
			handler.logger.Error("gRPC - SetValue - Cannot update db with the registered device with mac " + in.Mac)
			return nil, err
		}

		if updateResult.MatchedCount != 1 {
			handler.logger.Error("gRPC - SetValue - Cannot find a unique controller with mac " + in.Mac)
			return nil, fmt.Errorf("cannot find a unique controller with mac %s", in.Mac)
		}

		mqttFeatureValue := models.MqttFeatureValue{
			// profile info
			APIToken: controller.APIToken,
			// device info
			DeviceUUID: controller.DeviceUUID,
			Mac:        controller.Mac,
			Model:      controller.Model,
			// feature info
			FeatureUUID: controller.FeatureUUID,
			FeatureName: controller.FeatureName,
			Value:       value.Value,
		}
		mqttValues = append(mqttValues, mqttFeatureValue)
	}

	messageJSON, err := json.Marshal(mqttValues)
	if err != nil {
		handler.logger.Errorf("gRPC - SetValue - Cannot create mqtt payload %v\n", err)
		return nil, err
	}
	t := mqtt_client.SendValues(in.DeviceUuid, messageJSON)
	timeoutResult := t.WaitTimeout(devicesTimeout)
	if t.Error() != nil || !timeoutResult {
		handler.logger.Errorf("gRPC - SetValue - Cannot send data via mqtt %v\n", t.Error())
		return nil, t.Error()
	}
	handler.logger.Debug("gRPC - SetValue - Sending response")
	return &device.SetValueResponse{Status: "200", Message: "Updated"}, err
}
