package api

import (
	devicepb "api-devices/api/device"
	"api-devices/models"
	"api-devices/utils"
	"context"
	"encoding/json"
	"fmt"
	"testing"

	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestRandomHexReturnsRequestedByteLength(t *testing.T) {
	got, err := randomHex(16)
	if err != nil {
		t.Fatalf("random hex: %v", err)
	}
	if len(got) != 32 {
		t.Fatalf("hex length = %d, want 32", len(got))
	}
}

func TestControllerAPITokenRejectsMissingToken(t *testing.T) {
	_, err := controllerAPIToken(models.Controller{})
	if err == nil {
		t.Fatal("expected missing token error")
	}
}

func TestSignCommandBuildsSignedPayload(t *testing.T) {
	t.Setenv("API_TOKEN_ENCRYPTION_KEY", "12345678901234567890123456789012")

	encryptedToken, err := utils.EncryptAPIToken("api-token")
	if err != nil {
		t.Fatalf("encrypt api token: %v", err)
	}
	controller := models.Controller{
		APITokenEncrypted: encryptedToken,
		DeviceUUID:        uuid.NewString(),
		Mac:               "11:22:33:44:55:66",
		Model:             "thermostat",
		FeatureUUID:       uuid.NewString(),
		FeatureName:       "setpoint",
	}

	command, err := signCommand(controller, 21.5, 123, "nonce")
	if err != nil {
		t.Fatalf("sign command: %v", err)
	}

	payloadJSON, err := json.Marshal(models.Payload{Value: 21.5})
	if err != nil {
		t.Fatalf("marshal payload: %v", err)
	}
	signedPayload := fmt.Sprintf("%s\n%s\n%s\n%s\n%s\n%d\n%s\n%s",
		controller.DeviceUUID,
		controller.Mac,
		controller.Model,
		controller.FeatureUUID,
		controller.FeatureName,
		int64(123),
		"nonce",
		string(payloadJSON),
	)
	wantSignature := hmacSha256Hex("api-token", signedPayload)

	if command.Signature != wantSignature {
		t.Fatalf("signature = %q, want %q", command.Signature, wantSignature)
	}
	if command.Payload.Value != 21.5 {
		t.Fatalf("payload value = %v, want 21.5", command.Payload.Value)
	}
}

func TestSignCommandRejectsControllerWithoutToken(t *testing.T) {
	_, err := signCommand(models.Controller{}, 1, 123, "nonce")
	if err == nil {
		t.Fatal("expected missing token error")
	}
}

func TestSignCommandRejectsInvalidEncryptedToken(t *testing.T) {
	t.Setenv("API_TOKEN_ENCRYPTION_KEY", "12345678901234567890123456789012")

	_, err := signCommand(models.Controller{APITokenEncrypted: "%"}, 1, 123, "nonce")
	if err == nil {
		t.Fatal("expected decrypt token error")
	}
}

func TestGetValueRejectsInvalidDeviceUUID(t *testing.T) {
	client := &DevicesGrpc{logger: zap.NewNop().Sugar()}

	response, err := client.GetValue(context.Background(), &devicepb.GetValueRequest{
		DeviceUuid: "bad-device-uuid",
	})

	if response != nil {
		t.Fatal("expected nil response")
	}
	if status.Code(err) != codes.InvalidArgument {
		t.Fatalf("code = %v, want %v", status.Code(err), codes.InvalidArgument)
	}
}

func TestGetValueReturnsInternalWhenHashSecretIsMissing(t *testing.T) {
	t.Setenv("API_TOKEN_HASH_SECRET", "")
	client := &DevicesGrpc{logger: zap.NewNop().Sugar()}

	response, err := client.GetValue(context.Background(), &devicepb.GetValueRequest{
		DeviceUuid: uuid.NewString(),
	})

	if response != nil {
		t.Fatal("expected nil response")
	}
	if status.Code(err) != codes.Internal {
		t.Fatalf("code = %v, want %v", status.Code(err), codes.Internal)
	}
}

func TestSetValuesRejectsInvalidDeviceUUID(t *testing.T) {
	client := &DevicesGrpc{logger: zap.NewNop().Sugar()}

	response, err := client.SetValues(context.Background(), &devicepb.SetValuesRequest{
		DeviceUuid: "bad-device-uuid",
	})

	if response != nil {
		t.Fatal("expected nil response")
	}
	if status.Code(err) != codes.InvalidArgument {
		t.Fatalf("code = %v, want %v", status.Code(err), codes.InvalidArgument)
	}
}

func TestDeleteValueRejectsInvalidDeviceUUID(t *testing.T) {
	client := &DevicesGrpc{logger: zap.NewNop().Sugar()}

	response, err := client.DeleteValue(context.Background(), &devicepb.DeleteValueRequest{
		DeviceUuid: "bad-device-uuid",
	})

	if response != nil {
		t.Fatal("expected nil response")
	}
	if status.Code(err) != codes.InvalidArgument {
		t.Fatalf("code = %v, want %v", status.Code(err), codes.InvalidArgument)
	}
}

func TestDeleteValueRejectsInvalidFeatureUUID(t *testing.T) {
	client := &DevicesGrpc{logger: zap.NewNop().Sugar()}

	response, err := client.DeleteValue(context.Background(), &devicepb.DeleteValueRequest{
		DeviceUuid:  uuid.NewString(),
		FeatureUuid: "bad-feature-uuid",
	})

	if response != nil {
		t.Fatal("expected nil response")
	}
	if status.Code(err) != codes.InvalidArgument {
		t.Fatalf("code = %v, want %v", status.Code(err), codes.InvalidArgument)
	}
}

func TestSetValuesRejectsTooManyFeatureValues(t *testing.T) {
	client := &DevicesGrpc{logger: zap.NewNop().Sugar()}
	features := make([]*devicepb.SetValueRequest, 101)
	for i := range features {
		features[i] = &devicepb.SetValueRequest{FeatureUuid: bson.NewObjectID().Hex()}
	}

	response, err := client.SetValues(context.Background(), &devicepb.SetValuesRequest{
		FeatureValues: features,
	})

	if response != nil {
		t.Fatal("expected nil response")
	}
	if status.Code(err) != codes.InvalidArgument {
		t.Fatalf("code = %v, want %v", status.Code(err), codes.InvalidArgument)
	}
}

func TestSetValuesReturnsInternalWhenHashSecretIsMissing(t *testing.T) {
	t.Setenv("API_TOKEN_HASH_SECRET", "")
	client := &DevicesGrpc{logger: zap.NewNop().Sugar()}

	response, err := client.SetValues(context.Background(), &devicepb.SetValuesRequest{
		DeviceUuid: uuid.NewString(),
	})

	if response != nil {
		t.Fatal("expected nil response")
	}
	if status.Code(err) != codes.Internal {
		t.Fatalf("code = %v, want %v", status.Code(err), codes.Internal)
	}
}
