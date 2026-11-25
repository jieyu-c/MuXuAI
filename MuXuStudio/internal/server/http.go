package server

import (
	"encoding/json"
	http2 "net/http"
	"time"

	"github.com/go-kratos/kratos/v3/errors"
	"github.com/go-kratos/kratos/v3/middleware/validate"
	"github.com/go-kratos/kratos/v3/transport/http"
	pingv1 "jieyuc.cn/muxu/studio/api/ping/v1"
	userv1 "jieyuc.cn/muxu/studio/api/user/v1"
	"jieyuc.cn/muxu/studio/internal/conf"
	"jieyuc.cn/muxu/studio/internal/middlewire"
	"jieyuc.cn/muxu/studio/internal/service"
	"jieyuc.cn/muxu/studio/internal/token"

	"go.einride.tech/aip/fieldbehavior"
	"google.golang.org/protobuf/proto"
)

type HttpResponse struct {
	Code       int       `json:"code"`
	Msg        string    `json:"msg"`
	Data       any       `json:"data"`
	ServerTime time.Time `json:"server_time"`
}

// NewHTTPServer new an HTTP server.
func NewHTTPServer(c *conf.Server, ping service.PingService, user *service.UserService, tokens *token.Manager) *http.Server {
	var opts = []http.ServerOption{
		http.Middleware(
			middlewire.Recovery(),
			middlewire.Authorization(tokens, publicRoutes...),
			validate.Validator(func(req any) error {
				if msg, ok := req.(proto.Message); ok {
					if err := fieldbehavior.ValidateRequiredFields(msg); err != nil {
						return err
					}
				}
				return nil
			}),
		),
	}
	if c.Http.Network != "" {
		opts = append(opts, http.Network(c.Http.Network))
	}
	if c.Http.Addr != "" {
		opts = append(opts, http.Address(c.Http.Addr))
	}
	if c.Http.Timeout != nil {
		opts = append(opts, http.Timeout(c.Http.Timeout.AsDuration()))
	}
	opts = append(opts, http.ResponseEncoder(encodeResponseFunc))
	opts = append(opts, http.ErrorEncoder(encodeErrorFunc))

	srv := http.NewServer(opts...)
	pingv1.RegisterPingHTTPServer(srv, ping)
	userv1.RegisterUserServiceHTTPServer(srv, user)

	return srv
}

func encodeResponseFunc(writer http2.ResponseWriter, request *http2.Request, a any) error {
	return writeResponse(writer, a)
}

func encodeErrorFunc(writer http2.ResponseWriter, request *http2.Request, err error) {
	_ = writeResponse(writer, err)
}

func writeResponse(writer http2.ResponseWriter, a any) error {
	status := http2.StatusOK
	response := HttpResponse{
		Code:       0,
		Msg:        "ok",
		Data:       a,
		ServerTime: time.Now().UTC(),
	}
	if err, ok := a.(error); ok {
		se := errors.FromError(err)
		status = int(se.Code)
		response.Code = status
		response.Msg = se.Message
		response.Data = nil
	}
	responseJson, err := json.Marshal(response)
	if err != nil {
		return err
	}
	writer.Header().Set("Content-Type", "application/json")
	writer.WriteHeader(status)
	_, err = writer.Write(responseJson)
	return err
}
