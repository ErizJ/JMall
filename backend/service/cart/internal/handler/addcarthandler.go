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

func AddCartHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.AddCartReq
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		l := logic.NewAddCartLogic(r.Context(), svcCtx)
		resp, err := l.AddCart(&req)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}
