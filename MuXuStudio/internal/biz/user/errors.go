package user

import (
	"github.com/go-kratos/kratos/v3/errors"
	v1 "jieyuc.cn/muxu/studio/api/user/v1"
)

// ErrUserAlreadyExists identifies duplicate registration internally. Its public
// reason and message deliberately do not disclose whether a phone is registered.

var (
	ErrUserAlreadyExists = errors.BadRequest(
		v1.ErrorReason_REGISTER_FAILED.String(),
		"无法完成注册，请检查填写信息；如已有账号，请登录或找回密码。",
	)

	ErrUserNotExists = errors.Unauthorized(
		v1.ErrorReason_LOGIN_FAILED.String(),
		"登录失败，请确认用户名密码是否正确")
	ErrPasswordWrong = errors.Unauthorized(
		v1.ErrorReason_LOGIN_FAILED.String(), "登录失败，请确认用户名密码是否正确")
)
