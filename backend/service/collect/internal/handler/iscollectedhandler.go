package handler

import (
	"net/http"

	"github.com/ErizJ/JMall/backend/service/collect/internal/logic"
	"github.com/ErizJ/JMall/backend/service/collect/internal/svc"
	"github.com/ErizJ/JMall/backend/service/collect/internal/types"
	"github.com/zeromicro/go-zero/rest/httpx"
)

func IsCollectedHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.IsCollectedReq
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		l := logic.NewIsCollectedLogic(r.Context(), svcCtx)
		resp, err := l.IsCollected(&req)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}
