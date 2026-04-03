package handler

import (
	"net/http"

	"github.com/ErizJ/JMall/backend/service/seckill/internal/logic"
	"github.com/ErizJ/JMall/backend/service/seckill/internal/svc"
	"github.com/ErizJ/JMall/backend/service/seckill/internal/types"
	"github.com/zeromicro/go-zero/rest/httpx"
)

func GetSeckillActivityHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.GetSeckillActivityReq
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		l := logic.NewGetSeckillActivityLogic(r.Context(), svcCtx)
		resp, err := l.GetSeckillActivity(&req)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}
