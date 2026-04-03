package handler

import (
	"net/http"

	"github.com/ErizJ/JMall/backend/service/recommendation/internal/logic"
	"github.com/ErizJ/JMall/backend/service/recommendation/internal/svc"
	"github.com/ErizJ/JMall/backend/service/recommendation/internal/types"
	"github.com/zeromicro/go-zero/rest/httpx"
)

func FillUpHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.FillUpReq
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		l := logic.NewFillUpLogic(r.Context(), svcCtx)
		resp, err := l.FillUp(&req)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}
