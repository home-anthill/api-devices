package api

import (
	"api-devices/api/device"
	"api-devices/db"
	"api-devices/models"
	mqttclient "api-devices/mqttclient"
	"context"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const devicesTimeout = 5 * time.Second
const commandNonceBytes = 16

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

func randomHex(byteLen int) (string, error) {
	bytes := make([]byte, byteLen)
	if _, err := rand.Read(bytes); err != nil {
		return "", fmt.Errorf("generate nonce: %w", err)
	}
	return hex.EncodeToString(bytes), nil
}

func hmacSha256Hex(key, message string) string {
	mac := hmac.New(sha256.New, []byte(key))
	mac.Write([]byte(message))
	return hex.EncodeToString(mac.Sum(nil))
}

func signCommand(controller models.Controller, value float32, timestamp int64, nonce string) (models.MqttFeatureValue, error) {
	payload := models.Payload{Value: value}
	payloadJSON, err := json.Marshal(payload)
	if err != nil {
		return models.MqttFeatureValue{}, fmt.Errorf("marshal command payload: %w", err)
	}
	signedPayload := fmt.Sprintf("%s\n%s\n%s\n%s\n%s\n%d\n%s\n%s",
		controller.DeviceUUID,
		controller.Mac,
		controller.Model,
		controller.FeatureUUID,
		controller.FeatureName,
		timestamp,
		nonce,
		string(payloadJSON),
	)

	return models.MqttFeatureValue{
		DeviceUUID:  controller.DeviceUUID,
		Mac:         controller.Mac,
		Model:       controller.Model,
		FeatureUUID: controller.FeatureUUID,
		FeatureName: controller.FeatureName,
		Timestamp:   timestamp,
		Nonce:       nonce,
		Signature:   hmacSha256Hex(controller.APIToken, signedPayload),
		Payload:     payload,
	}, nil
}

// GetValue retrieves the current value of a device feature from the database.
func (d *DevicesGrpc) GetValue(ctx context.Context, in *device.GetValueRequest) (*device.GetValueResponse, error) {
	d.logger.Infof("gRPC - GetValue - Called for deviceUuid: %s, mac: %s", in.DeviceUuid, in.Mac)

	if _, err := uuid.Parse(in.DeviceUuid); err != nil {
		d.logger.Errorf("gRPC - GetValue - invalid deviceUuid: %v", err)
		return nil, status.Errorf(codes.InvalidArgument, "deviceUuid is not a valid UUID")
	}

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
	if _, err := uuid.Parse(in.DeviceUuid); err != nil {
		d.logger.Errorf("gRPC - SetValue - invalid deviceUuid: %v", err)
		return nil, status.Errorf(codes.InvalidArgument, "deviceUuid is not a valid UUID")
	}

	results := make([]models.MqttFeatureValue, len(in.FeatureValues))
	for i, value := range in.FeatureValues {
		var controller models.Controller
		err := d.controllersCollection.FindOne(ctx, bson.M{
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
			d.logger.Errorf("gRPC - SetValue - Cannot find device: %v", err)
			return nil, status.Errorf(codes.NotFound, "cannot find controller: %v", err)
		}

		now := time.Now()
		updatedStatus := models.Status{
			Value:      value.Value,
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
			"featureUuid": value.FeatureUuid,
			"featureName": value.FeatureName,
		}, bson.M{
			"$set": bson.M{
				"status":     updatedStatus,
				"modifiedAt": now,
			},
		})

		if err != nil {
			d.logger.Errorf("gRPC - SetValue - Cannot update db with the registered device: %v", err)
			return nil, status.Errorf(codes.Internal, "cannot update controller: %v", err)
		}

		if updateResult.MatchedCount != 1 {
			d.logger.Error("gRPC - SetValue - Cannot find a unique controller")
			return nil, status.Errorf(codes.NotFound, "cannot find a unique controller")
		}

		nonce, err := randomHex(commandNonceBytes)
		if err != nil {
			d.logger.Errorf("gRPC - SetValue - Cannot generate command nonce: %v", err)
			return nil, status.Errorf(codes.Internal, "cannot generate command nonce: %v", err)
		}
		command, err := signCommand(controller, value.Value, now.Unix(), nonce)
		if err != nil {
			d.logger.Errorf("gRPC - SetValue - Cannot sign mqtt payload: %v", err)
			return nil, status.Errorf(codes.Internal, "cannot sign mqtt payload: %v", err)
		}
		results[i] = command
	}

	messageJSON, err := json.Marshal(results)
	if err != nil {
		d.logger.Errorf("gRPC - SetValue - Cannot create mqtt payload: %v", err)
		return nil, status.Errorf(codes.Internal, "cannot create mqtt payload: %v", err)
	}
	t, err := mqttclient.SendValues(in.DeviceUuid, messageJSON)
	if err != nil {
		d.logger.Errorf("gRPC - SetValue - invalid MQTT publish topic: %v", err)
		return nil, status.Errorf(codes.InvalidArgument, "deviceUuid is not a valid UUID")
	}
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
