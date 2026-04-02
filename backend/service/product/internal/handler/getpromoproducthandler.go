// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package handler

import (
	"net/http"

	"github.com/ErizJ/JMall/backend/service/product/internal/logic"
	"github.com/ErizJ/JMall/backend/service/product/internal/svc"
	"github.com/ErizJ/JMall/backend/service/product/internal/types"
	"github.com/zeromicro/go-zero/rest/httpx"
)

func GetPromoProductHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.GetPromoProductReq
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		l := logic.NewGetPromoProductLogic(r.Context(), svcCtx)
		resp, err := l.GetPromoProduct(&req)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}
