// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package handler

import (
	"net/http"

	"github.com/ErizJ/JMall/backend/service/cart/internal/logic"
	"github.com/ErizJ/JMall/backend/service/cart/internal/svc"
	"github.com/ErizJ/JMall/backend/service/cart/internal/types"
	"github.com/zeromicro/go-zero/rest/httpx"
)

func DeleteCartHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.DeleteCartReq
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		l := logic.NewDeleteCartLogic(r.Context(), svcCtx)
		resp, err := l.DeleteCart(&req)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}
