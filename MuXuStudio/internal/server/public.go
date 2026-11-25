package server

import "jieyuc.cn/muxu/studio/internal/middlewire"

var publicRoutes = []middlewire.Route{
	{
		Method: "POST",
		Path:   "/api/v1/users/register",
	},
	{
		Method: "POST",
		Path:   "/api/v1/users/login",
	},
}
