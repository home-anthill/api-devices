package api

import (
	"api-devices/api/device"
	"api-devices/models"
	mqtt_client "api-devices/mqtt-client"
	"context"
	"encoding/json"
	"fmt"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.uber.org/zap"
	"time"
)

const devicesTimeout = 5 * time.Second

type DevicesGrpc struct {
	device.UnimplementedDeviceServer
	airConditionerCollection *mongo.Collection
	contextRef               context.Context
	logger                   *zap.SugaredLogger
}

func NewDevicesGrpc(ctx context.Context, logger *zap.SugaredLogger, collection *mongo.Collection) *DevicesGrpc {
	return &DevicesGrpc{
		airConditionerCollection: collection,
		contextRef:               ctx,
		logger:                   logger,
	}
}

func (handler *DevicesGrpc) GetStatus(ctx context.Context, in *device.StatusRequest) (*device.StatusResponse, error) {
	handler.logger.Infof("gRPC - GetStatus - Called with in: %#v", in)

	var ac models.AirConditioner
	err := handler.airConditionerCollection.FindOne(handler.contextRef, bson.M{
		"mac": in.Mac,
	}).Decode(&ac)
	if err != nil {
		handler.logger.Error("gRPC - GetStatus -  Cannot get AC with specified mac " + in.Mac)
		return nil, fmt.Errorf("cannot find device with mac %s", in.Mac)
	}

	return &device.StatusResponse{
		On:          ac.Status.On,
		Temperature: int32(ac.Status.Temperature),
		Mode:        int32(ac.Status.Mode),
		FanSpeed:    int32(ac.Status.FanSpeed),
		CreatedAt:   ac.CreatedAt.UnixMilli(),
		ModifiedAt:  ac.ModifiedAt.UnixMilli(),
	}, err
}

func (handler *DevicesGrpc) SetValues(ctx context.Context, in *device.ValuesRequest) (*device.ValuesResponse, error) {
	handler.logger.Infof("gRPC - SetValues - Called with in: %#v", in)

	updateResult, err := handler.airConditionerCollection.UpdateOne(handler.contextRef, bson.M{
		"mac": in.Mac,
	}, bson.M{
		"$set": bson.M{
			"status.on":          in.On,
			"status.temperature": in.Temperature,
			"status.mode":        in.Mode,
			"status.fanSpeed":    in.FanSpeed,
			"modifiedAt":         time.Now(),
		},
	})

	if err != nil {
		handler.logger.Error("gRPC - SetValues -  Cannot update db with the registered AC with mac " + in.Mac)
		return nil, err
	}

	if updateResult.MatchedCount != 1 {
		handler.logger.Error("gRPC - SetValues -  Cannot find a unique device with mac " + in.Mac)
		return nil, fmt.Errorf("cannot find a unique device with mac %s", in.Mac)
	}

	values := models.Values{
		Uuid:        in.Uuid,
		ApiToken:    in.ApiToken,
		On:          in.On,
		Temperature: int(in.Temperature),
		Mode:        int(in.Mode),
		FanSpeed:    int(in.FanSpeed),
	}
	messageJSON, err := json.Marshal(values)
	if err != nil {
		handler.logger.Errorf("gRPC - SetValues - Cannot create mqtt payload %v\n", err)
		return nil, err
	}
	t := mqtt_client.SendValues(values.Uuid, messageJSON)
	timeoutResult := t.WaitTimeout(devicesTimeout)
	if t.Error() != nil || !timeoutResult {
		handler.logger.Errorf("gRPC - SetValues - Cannot send data via mqtt %v\n", t.Error())
		return nil, t.Error()
	} else {
		handler.logger.Debug("gRPC - SetValues - Sending response")
		return &device.ValuesResponse{Status: "200", Message: "Updated"}, err
	}
}
