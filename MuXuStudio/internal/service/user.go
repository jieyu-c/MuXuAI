package service

import (
	"context"

	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/timestamppb"
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

func (u *UserService) Login(ctx context.Context, request *v1.LoginRequest) (*v1.LoginResponse, error) {
	result, err := u.usecase.Login(ctx, request.Phone, request.Password)
	if err != nil {
		return nil, err
	}
	return &v1.LoginResponse{
		Token:       result.Token,
		ExpiredTime: timestamppb.New(result.ExpireAt),
	}, nil
}

func (u UserService) Me(ctx context.Context, empty *emptypb.Empty) (*v1.MeResponse, error) {
	userInfo, err := u.usecase.Me(ctx)
	if err != nil {
		return nil, err
	}
	return &v1.MeResponse{
		Phone: userInfo.Phone,
	}, nil
}
