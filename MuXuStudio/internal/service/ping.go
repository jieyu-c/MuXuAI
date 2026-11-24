package service

import (
	"context"

	"google.golang.org/protobuf/types/known/emptypb"
	v1 "jieyuc.cn/muxu/studio/api/ping/v1"
)

type PingService struct {
}

func NewPingService() PingService {
	return PingService{}
}

var _ v1.PingHTTPServer = (*PingService)(nil)

func (p PingService) Ping(ctx context.Context, empty *emptypb.Empty) (*v1.PingResponse, error) {
	return &v1.PingResponse{
		Ping: "pong",
	}, nil
}
