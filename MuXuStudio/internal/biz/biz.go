package biz

import (
	"github.com/google/wire"
	"jieyuc.cn/muxu/studio/internal/biz/user"
)

// ProviderSet is biz providers.
var ProviderSet = wire.NewSet(user.NewUserUsecase)
