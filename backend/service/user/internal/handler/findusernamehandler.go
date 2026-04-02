// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package handler

import (
	"net/http"

	"github.com/ErizJ/JMall/backend/service/user/internal/logic"
	"github.com/ErizJ/JMall/backend/service/user/internal/svc"
	"github.com/ErizJ/JMall/backend/service/user/internal/types"
	"github.com/zeromicro/go-zero/rest/httpx"
)

func FindUserNameHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.CheckUserNameReq
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		l := logic.NewFindUserNameLogic(r.Context(), svcCtx)
		resp, err := l.FindUserName(&req)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}
