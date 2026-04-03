package handler

import (
	"net/http"

	"github.com/ErizJ/JMall/backend/service/recommendation/internal/svc"
	"github.com/zeromicro/go-zero/rest"
)

func RegisterHandlers(server *rest.Server, serverCtx *svc.ServiceContext) {
	server.AddRoutes(
		rest.WithMiddlewares(
			[]rest.Middleware{serverCtx.AuthMiddleware},
			[]rest.Route{
				{
					Method:  http.MethodPost,
					Path:    "/recommend/fillup",
					Handler: FillUpHandler(serverCtx),
				},
			}...,
		),
	)
}
