package api

import (
	"api-devices/api/keepalive"
	"context"
	"go.uber.org/zap"
)

// KeepAliveGrpc struct
type KeepAliveGrpc struct {
	keepalive.UnimplementedKeepAliveServer
	ctx    context.Context
	logger *zap.SugaredLogger
}

// NewKeepAliveGrpc function
func NewKeepAliveGrpc(ctx context.Context, logger *zap.SugaredLogger) *KeepAliveGrpc {
	return &KeepAliveGrpc{
		ctx:    ctx,
		logger: logger,
	}
}

// GetKeepAlive function
func (handler *KeepAliveGrpc) GetKeepAlive(ctx context.Context, in *keepalive.StatusRequest) (*keepalive.StatusResponse, error) {
	return &keepalive.StatusResponse{}, nil
}
