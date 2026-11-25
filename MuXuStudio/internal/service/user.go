package service

import (
	"context"

	emptypb "google.golang.org/protobuf/types/known/emptypb"
	v1 "jieyuc.cn/muxu/studio/api/user/v1"
	userbiz "jieyuc.cn/muxu/studio/internal/biz/user"
)

type UserService struct {
	usecase *userbiz.Usecase
}

func NewUserService(u *userbiz.Usecase) *UserService {
	return &UserService{u}
}

var _ v1.UserServiceHTTPServer = (*UserService)(nil)

func (u UserService) Register(ctx context.Context, request *v1.RegisterRequest) (*emptypb.Empty, error) {
	err := u.usecase.Register(ctx, request.Phone, request.Password)
	return &emptypb.Empty{}, err
}
