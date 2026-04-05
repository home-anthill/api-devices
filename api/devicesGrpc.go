package api

import (
	"api-devices/api/device"
	"api-devices/db"
	"api-devices/models"
	mqttclient "api-devices/mqttclient"
	"context"
	"encoding/json"
	"sync"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const devicesTimeout = 5 * time.Second

// DevicesGrpc implements the gRPC Device service.
type DevicesGrpc struct {
	device.UnimplementedDeviceServer
	controllersCollection *mongo.Collection
	logger                *zap.SugaredLogger
}

// NewDevicesGrpc creates a new DevicesGrpc handler.
func NewDevicesGrpc(logger *zap.SugaredLogger, client *mongo.Client) *DevicesGrpc {
	return &DevicesGrpc{
		controllersCollection: db.GetCollections(client).Controllers,
		logger:                logger,
	}
}

// GetValue retrieves the current value of a device feature from the database.
func (d *DevicesGrpc) GetValue(ctx context.Context, in *device.GetValueRequest) (*device.GetValueResponse, error) {
	d.logger.Infof("gRPC - GetValue - Called for deviceUuid: %s, mac: %s", in.DeviceUuid, in.Mac)

	var controller models.Controller
	err := d.controllersCollection.FindOne(ctx, bson.M{
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
		d.logger.Errorf("gRPC - GetValue - Cannot get device: %v", err)
		return nil, status.Errorf(codes.NotFound, "cannot find controller: %v", err)
	}

	statusResponse := device.GetValueResponse{
		FeatureUuid: controller.FeatureUUID,
		FeatureName: controller.FeatureName,
		Value:       controller.Status.Value,
		CreatedAt:   controller.Status.CreatedAt.UnixMilli(),
		ModifiedAt:  controller.Status.ModifiedAt.UnixMilli(),
	}
	return &statusResponse, nil
}

// SetValues updates device feature values in the database and publishes them via MQTT.
func (d *DevicesGrpc) SetValues(ctx context.Context, in *device.SetValuesRequest) (*device.SetValueResponse, error) {
	d.logger.Infof("gRPC - SetValue - Called for deviceUuid: %s, mac: %s, featureValues: %d", in.DeviceUuid, in.Mac, len(in.FeatureValues))

	const maxFeatureValues = 100
	if len(in.FeatureValues) > maxFeatureValues {
		d.logger.Errorf("gRPC - SetValue - too many feature values: %d", len(in.FeatureValues))
		return nil, status.Errorf(codes.InvalidArgument, "too many feature values: got %d, max %d", len(in.FeatureValues), maxFeatureValues)
	}

	results := make([]models.MqttFeatureValue, len(in.FeatureValues))
	var mu sync.Mutex
	var firstErr error

	var wg sync.WaitGroup
	for i, value := range in.FeatureValues {
		wg.Add(1)
		go func(idx int, val *device.SetValueRequest) {
			defer wg.Done()

			var controller models.Controller
			err := d.controllersCollection.FindOne(ctx, bson.M{
				// profile info
				"apiToken": in.ApiToken,
				// device info
				"deviceUuid": in.DeviceUuid,
				"mac":        in.Mac,
				// feature info
				"featureUuid": val.FeatureUuid,
				"featureName": val.FeatureName,
			}).Decode(&controller)
			if err != nil {
				mu.Lock()
				if firstErr == nil {
					d.logger.Errorf("gRPC - SetValue - Cannot find device: %v", err)
					firstErr = status.Errorf(codes.NotFound, "cannot find controller: %v", err)
				}
				mu.Unlock()
				return
			}

			now := time.Now()
			updatedStatus := models.Status{
				Value:      val.Value,
				CreatedAt:  now,
				ModifiedAt: now,
			}

			d.logger.Debugf("gRPC - SetValue - updatedStatus %#v ", updatedStatus)

			updateResult, err := d.controllersCollection.UpdateOne(ctx, bson.M{
				// profile info
				"apiToken": in.ApiToken,
				// device info
				"deviceUuid": in.DeviceUuid,
				"mac":        in.Mac,
				// feature info
				"featureUuid": val.FeatureUuid,
				"featureName": val.FeatureName,
			}, bson.M{
				"$set": bson.M{
					"status":     updatedStatus,
					"modifiedAt": now,
				},
			})

			if err != nil {
				mu.Lock()
				if firstErr == nil {
					d.logger.Errorf("gRPC - SetValue - Cannot update db with the registered device: %v", err)
					firstErr = status.Errorf(codes.Internal, "cannot update controller: %v", err)
				}
				mu.Unlock()
				return
			}

			if updateResult.MatchedCount != 1 {
				mu.Lock()
				if firstErr == nil {
					d.logger.Error("gRPC - SetValue - Cannot find a unique controller")
					firstErr = status.Errorf(codes.NotFound, "cannot find a unique controller")
				}
				mu.Unlock()
				return
			}

			results[idx] = models.MqttFeatureValue{
				// profile info
				APIToken: controller.APIToken,
				// device info
				DeviceUUID: controller.DeviceUUID,
				Mac:        controller.Mac,
				Model:      controller.Model,
				// feature info
				FeatureUUID: controller.FeatureUUID,
				FeatureName: controller.FeatureName,
				Value:       val.Value,
			}
		}(i, value)
	}
	wg.Wait()

	if firstErr != nil {
		return nil, firstErr
	}

	messageJSON, err := json.Marshal(results)
	if err != nil {
		d.logger.Errorf("gRPC - SetValue - Cannot create mqtt payload: %v", err)
		return nil, status.Errorf(codes.Internal, "cannot create mqtt payload: %v", err)
	}
	t := mqttclient.SendValues(in.DeviceUuid, messageJSON)
	if !t.WaitTimeout(devicesTimeout) {
		d.logger.Error("gRPC - SetValue - MQTT publish timed out")
		return nil, status.Errorf(codes.Unavailable, "mqtt publish timed out")
	}
	if t.Error() != nil {
		d.logger.Errorf("gRPC - SetValue - Cannot send data via mqtt: %v", t.Error())
		return nil, status.Errorf(codes.Internal, "cannot send data via mqtt: %v", t.Error())
	}
	d.logger.Debug("gRPC - SetValue - Sending response")
	return &device.SetValueResponse{Status: "200", Message: "Updated"}, nil
}
