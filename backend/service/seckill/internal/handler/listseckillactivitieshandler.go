package handler

import (
	"net/http"

	"github.com/ErizJ/JMall/backend/service/seckill/internal/logic"
	"github.com/ErizJ/JMall/backend/service/seckill/internal/svc"
	"github.com/zeromicro/go-zero/rest/httpx"
)

func ListSeckillActivitiesHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		l := logic.NewListSeckillActivitiesLogic(r.Context(), svcCtx)
		resp, err := l.ListSeckillActivities()
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}
