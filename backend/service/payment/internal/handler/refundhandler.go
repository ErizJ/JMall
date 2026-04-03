package handler

import (
	"net/http"

	"github.com/ErizJ/JMall/backend/service/payment/internal/logic"
	"github.com/ErizJ/JMall/backend/service/payment/internal/svc"
	"github.com/ErizJ/JMall/backend/service/payment/internal/types"
	"github.com/zeromicro/go-zero/rest/httpx"
)

func RefundHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.RefundReq
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		l := logic.NewRefundLogic(r.Context(), svcCtx)
		resp, err := l.Refund(&req)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}
