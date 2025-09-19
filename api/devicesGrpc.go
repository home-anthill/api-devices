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

  "go.mongodb.org/mongo-driver/bson"
  "go.mongodb.org/mongo-driver/mongo"
  "go.uber.org/zap"
)

const devicesTimeout = 5 * time.Second

// DevicesGrpc struct
type DevicesGrpc struct {
  device.UnimplementedDeviceServer
  client                   *mongo.Client
  airConditionerCollection *mongo.Collection
  setpointCollection       *mongo.Collection
  toleranceCollection      *mongo.Collection
  contextRef               context.Context
  logger                   *zap.SugaredLogger
}

// NewDevicesGrpc function
func NewDevicesGrpc(ctx context.Context, logger *zap.SugaredLogger, client *mongo.Client) *DevicesGrpc {
  return &DevicesGrpc{
    client:                   client,
    airConditionerCollection: db.GetCollections(client).AirConditioners,
    setpointCollection:       db.GetCollections(client).Setpoints,
    toleranceCollection:      db.GetCollections(client).Tolerances,
    contextRef:               ctx,
    logger:                   logger,
  }
}

// GetValue function
func (handler *DevicesGrpc) GetValue(ctx context.Context, in *device.GetValueRequest) (*device.GetValueResponse, error) {
  handler.logger.Infof("gRPC - GetValue - Called with in: %#v", in)

  var collection *mongo.Collection
  switch in.FeatureName {
  case "ac-lg", "ac-beko":
    collection = handler.airConditionerCollection
  case "setpoint":
    collection = handler.setpointCollection
  case "tolerance":
    collection = handler.toleranceCollection
  }
  if collection == nil {
    handler.logger.Error("gRPC - GetValue - Unknown controller feature name = '" + in.FeatureName + "', mac =" + in.Mac)
    return nil, fmt.Errorf("unknown controller feature name")
  }

  var controller models.Device
  err := collection.FindOne(handler.contextRef, bson.M{
    "mac": in.Mac,
  }).Decode(&controller)
  if err != nil {
    handler.logger.Error("gRPC - GetValue - Cannot get device with specified mac " + in.Mac)
    return nil, fmt.Errorf("cannot find controller with mac %s", in.Mac)
  }

  statusResponse := device.GetValueResponse{
    FeatureUuid: controller.UUID,
    FeatureName: controller.Name,
    Value:       controller.Status.Value,
    CreatedAt:   controller.Status.CreatedAt.UnixMilli(),
    ModifiedAt:  controller.Status.ModifiedAt.UnixMilli(),
  }
  return &statusResponse, err
}

// SetValue function
func (handler *DevicesGrpc) SetValue(ctx context.Context, in *device.SetValueRequest) (*device.SetValueResponse, error) {
  handler.logger.Infof("gRPC - SetValue - Called with in: %#v", in)

  var collection *mongo.Collection
  switch in.FeatureName {
  case "ac-lg", "ac-beko":
    collection = handler.airConditionerCollection
  case "setpoint":
    collection = handler.setpointCollection
  case "tolerance":
    collection = handler.toleranceCollection
  }
  if collection == nil {
    handler.logger.Error("gRPC - GetValue - Unknown controller feature name = '" + in.FeatureName + "', mac =" + in.Mac)
    return nil, fmt.Errorf("unknown controller feature name")
  }

  var controller models.Device
  err := collection.FindOne(handler.contextRef, bson.M{
    "mac": in.Mac,
  }).Decode(&controller)
  if err != nil {
    handler.logger.Error("gRPC - SetValue - Cannot find device with specified mac " + in.Mac)
    return nil, fmt.Errorf("cannot find controller with mac %s", in.Mac)
  }

  updatedStatues := models.Status{
    Value:      in.Value,
    CreatedAt:  time.Now(),
    ModifiedAt: time.Now(),
  }

  handler.logger.Debugf("gRPC - SetValue - updatedStatues %#v ", updatedStatues)

  updateResult, err := collection.UpdateOne(handler.contextRef, bson.M{
    "mac": in.Mac,
  }, bson.M{
    "$set": bson.M{
      "status":     updatedStatues,
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

  values := models.Values{
    UUID:     in.FeatureUuid,
    APIToken: in.ApiToken,
    Value:    float64(in.Value),
  }
  messageJSON, err := json.Marshal(values)
  if err != nil {
    handler.logger.Errorf("gRPC - SetValue - Cannot create mqtt payload %v\n", err)
    return nil, err
  }
  t := mqtt_client.SendValues(values.UUID, messageJSON)
  timeoutResult := t.WaitTimeout(devicesTimeout)
  if t.Error() != nil || !timeoutResult {
    handler.logger.Errorf("gRPC - SetValue - Cannot send data via mqtt %v\n", t.Error())
    return nil, t.Error()
  }
  handler.logger.Debug("gRPC - SetValue - Sending response")
  return &device.SetValueResponse{Status: "200", Message: "Updated"}, err
}
