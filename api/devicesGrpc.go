package api

import (
	"api-devices/api/device"
	"api-devices/db"
	"api-devices/models"
	mqttclient "api-devices/mqttclient"
	"api-devices/utils"
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

type statusUpdate struct {
	controllerID bson.ObjectID
	status       models.Status
}

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
	apiToken, err := controllerAPIToken(controller)
	if err != nil {
		return models.MqttFeatureValue{}, err
	}
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
		Signature:   hmacSha256Hex(apiToken, signedPayload),
		Payload:     payload,
	}, nil
}

func controllerAPIToken(controller models.Controller) (string, error) {
	if controller.APITokenEncrypted != "" {
		return utils.DecryptAPIToken(controller.APITokenEncrypted)
	}
	return "", fmt.Errorf("controller has no usable api token")
}

// GetValue retrieves the current value of a device feature from the database.
func (d *DevicesGrpc) GetValue(ctx context.Context, in *device.GetValueRequest) (*device.GetValueResponse, error) {
	d.logger.Infof("gRPC - GetValue - Called for deviceUuid: %s, mac: %s", in.DeviceUuid, in.Mac)

	if _, err := uuid.Parse(in.DeviceUuid); err != nil {
		d.logger.Errorf("gRPC - GetValue - invalid deviceUuid: %v", err)
		return nil, status.Errorf(codes.InvalidArgument, "deviceUuid is not a valid UUID")
	}
	apiTokenHash, err := utils.HashAPIToken(in.ApiToken)
	if err != nil {
		d.logger.Errorf("gRPC - GetValue - Cannot hash apiToken: %v", err)
		return nil, status.Errorf(codes.Internal, "cannot get controller")
	}

	var controller models.Controller
	err = d.controllersCollection.FindOne(ctx, bson.M{
		// profile info
		"apiTokenHash": apiTokenHash,
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

// SetValues publishes device feature values via MQTT and records them after publish success.
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
	apiTokenHash, err := utils.HashAPIToken(in.ApiToken)
	if err != nil {
		d.logger.Errorf("gRPC - SetValue - Cannot hash apiToken: %v", err)
		return nil, status.Errorf(codes.Internal, "cannot set controller value")
	}

	results := make([]models.MqttFeatureValue, len(in.FeatureValues))
	updates := make([]statusUpdate, len(in.FeatureValues))
	for i, value := range in.FeatureValues {
		var controller models.Controller
		err = d.controllersCollection.FindOne(ctx, bson.M{
			// profile info
			"apiTokenHash": apiTokenHash,
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

		var nonce string
		nonce, err = randomHex(commandNonceBytes)
		if err != nil {
			d.logger.Errorf("gRPC - SetValue - Cannot generate command nonce: %v", err)
			return nil, status.Errorf(codes.Internal, "cannot generate command nonce: %v", err)
		}
		var command models.MqttFeatureValue
		command, err = signCommand(controller, value.Value, now.Unix(), nonce)
		if err != nil {
			d.logger.Errorf("gRPC - SetValue - Cannot sign mqtt payload: %v", err)
			return nil, status.Errorf(codes.Internal, "cannot sign mqtt payload: %v", err)
		}
		results[i] = command
		updates[i] = statusUpdate{
			controllerID: controller.ID,
			status:       updatedStatus,
		}
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
	for _, update := range updates {
		updateResult, err := d.controllersCollection.UpdateOne(ctx, bson.M{
			"_id": update.controllerID,
		}, bson.M{
			"$set": bson.M{
				"status":     update.status,
				"modifiedAt": update.status.ModifiedAt,
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
	}
	d.logger.Debug("gRPC - SetValue - Sending response")
	return &device.SetValueResponse{Status: "200", Message: "Updated"}, nil
}
