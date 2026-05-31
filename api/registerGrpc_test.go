package api

import (
	"api-devices/api/register"
	"context"
	"testing"

	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestRegisterRejectsMissingFeature(t *testing.T) {
	client := &RegisterGrpc{logger: zap.NewNop().Sugar()}

	response, err := client.Register(context.Background(), &register.RegisterRequest{})

	if response != nil {
		t.Fatal("expected nil response")
	}
	if status.Code(err) != codes.InvalidArgument {
		t.Fatalf("code = %v, want %v", status.Code(err), codes.InvalidArgument)
	}
}

func TestRegisterRejectsInvalidDeviceUUID(t *testing.T) {
	client := &RegisterGrpc{logger: zap.NewNop().Sugar()}

	response, err := client.Register(context.Background(), &register.RegisterRequest{
		DeviceUuid: "bad-device-uuid",
		Feature:    &register.RegisterFeature{},
	})

	if response != nil {
		t.Fatal("expected nil response")
	}
	if status.Code(err) != codes.InvalidArgument {
		t.Fatalf("code = %v, want %v", status.Code(err), codes.InvalidArgument)
	}
}

func TestRegisterRejectsInvalidProfileOwnerID(t *testing.T) {
	client := &RegisterGrpc{logger: zap.NewNop().Sugar()}

	response, err := client.Register(context.Background(), &register.RegisterRequest{
		DeviceUuid:     uuid.NewString(),
		ProfileOwnerId: "bad-profile-owner-id",
		Feature:        &register.RegisterFeature{},
	})

	if response != nil {
		t.Fatal("expected nil response")
	}
	if status.Code(err) != codes.InvalidArgument {
		t.Fatalf("code = %v, want %v", status.Code(err), codes.InvalidArgument)
	}
}

func TestRegisterReturnsInternalWhenEncryptionKeyIsMissing(t *testing.T) {
	t.Setenv("API_TOKEN_ENCRYPTION_KEY", "")
	client := &RegisterGrpc{logger: zap.NewNop().Sugar()}

	response, err := client.Register(context.Background(), &register.RegisterRequest{
		DeviceUuid:     uuid.NewString(),
		ProfileOwnerId: bson.NewObjectID().Hex(),
		Feature:        &register.RegisterFeature{},
	})

	if response != nil {
		t.Fatal("expected nil response")
	}
	if status.Code(err) != codes.Internal {
		t.Fatalf("code = %v, want %v", status.Code(err), codes.Internal)
	}
}

func TestRegisterReturnsInternalWhenHashSecretIsMissing(t *testing.T) {
	t.Setenv("API_TOKEN_ENCRYPTION_KEY", "12345678901234567890123456789012")
	t.Setenv("API_TOKEN_HASH_SECRET", "")
	client := &RegisterGrpc{logger: zap.NewNop().Sugar()}

	response, err := client.Register(context.Background(), &register.RegisterRequest{
		DeviceUuid:     uuid.NewString(),
		ProfileOwnerId: bson.NewObjectID().Hex(),
		Feature:        &register.RegisterFeature{},
	})

	if response != nil {
		t.Fatal("expected nil response")
	}
	if status.Code(err) != codes.Internal {
		t.Fatalf("code = %v, want %v", status.Code(err), codes.Internal)
	}
}
