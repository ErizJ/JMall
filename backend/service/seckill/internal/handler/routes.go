package handler

import (
	"net/http"

	"github.com/ErizJ/JMall/backend/service/seckill/internal/svc"
	"github.com/zeromicro/go-zero/rest"
)

func RegisterHandlers(server *rest.Server, serverCtx *svc.ServiceContext) {
	// 需要鉴权的接口
	server.AddRoutes(
		rest.WithMiddlewares(
			[]rest.Middleware{serverCtx.AuthMiddleware},
			[]rest.Route{
				{
					Method:  http.MethodPost,
					Path:    "/seckill/buy",
					Handler: SeckillHandler(serverCtx),
				},
				{
					Method:  http.MethodPost,
					Path:    "/seckill/result",
					Handler: SeckillResultHandler(serverCtx),
				},
			}...,
		),
	)

	// 不需要鉴权的接口
	server.AddRoutes(
		[]rest.Route{
			{
				Method:  http.MethodPost,
				Path:    "/seckill/activity",
				Handler: GetSeckillActivityHandler(serverCtx),
			},
			{
				Method:  http.MethodGet,
				Path:    "/seckill/activities",
				Handler: ListSeckillActivitiesHandler(serverCtx),
			},
		},
	)
}
