// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package handler

import (
	"net/http"

	"github.com/ErizJ/JMall/backend/service/product/internal/logic"
	"github.com/ErizJ/JMall/backend/service/product/internal/svc"
	"github.com/zeromicro/go-zero/rest/httpx"
)

func GetAllProductHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		l := logic.NewGetAllProductLogic(r.Context(), svcCtx)
		resp, err := l.GetAllProduct()
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}
