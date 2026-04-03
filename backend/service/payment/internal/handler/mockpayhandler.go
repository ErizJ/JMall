package handler

import (
	"net/http"

	"github.com/ErizJ/JMall/backend/service/payment/internal/logic"
	"github.com/ErizJ/JMall/backend/service/payment/internal/svc"
	"github.com/ErizJ/JMall/backend/service/payment/internal/types"
	"github.com/zeromicro/go-zero/rest/httpx"
)

func MockPayHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.MockPayReq
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		l := logic.NewMockPayLogic(r.Context(), svcCtx)
		resp, err := l.MockPay(&req)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}
