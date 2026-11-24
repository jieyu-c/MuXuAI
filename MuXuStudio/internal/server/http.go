package server

import (
	"encoding/json"
	"log"
	http2 "net/http"
	"time"

	"github.com/go-kratos/kratos/v3/middleware/recovery"
	"github.com/go-kratos/kratos/v3/middleware/validate"
	"github.com/go-kratos/kratos/v3/transport/http"
	pingv1 "jieyuc.cn/muxu/studio/api/ping/v1"
	"jieyuc.cn/muxu/studio/internal/conf"
	"jieyuc.cn/muxu/studio/internal/service"

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
func NewHTTPServer(c *conf.Server, ping service.PingService) *http.Server {
	var opts = []http.ServerOption{
		http.Middleware(
			recovery.Recovery(),
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

	srv := http.NewServer(opts...)
	pingv1.RegisterPingHTTPServer(srv, ping)

	return srv
}

func encodeResponseFunc(writer http2.ResponseWriter, request *http2.Request, a any) error {
	writer.Header().Add("Content-Type", "application/json")
	log.Println(a)
	response := HttpResponse{
		Code:       0,
		Msg:        "ok",
		Data:       a,
		ServerTime: time.Now().UTC(),
	}
	responseJson, err := json.Marshal(response)
	if err != nil {
		return err
	}
	_, err = writer.Write(responseJson)
	return err
}
