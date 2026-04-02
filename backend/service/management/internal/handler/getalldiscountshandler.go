// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package handler

import (
	"net/http"

	"github.com/ErizJ/JMall/backend/service/management/internal/logic"
	"github.com/ErizJ/JMall/backend/service/management/internal/svc"
	"github.com/zeromicro/go-zero/rest/httpx"
)

func GetAllDiscountsHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		l := logic.NewGetAllDiscountsLogic(r.Context(), svcCtx)
		resp, err := l.GetAllDiscounts()
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}
