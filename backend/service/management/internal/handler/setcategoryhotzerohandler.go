package handler

import (
	"net/http"

	"github.com/ErizJ/JMall/backend/service/management/internal/logic"
	"github.com/ErizJ/JMall/backend/service/management/internal/svc"
	"github.com/zeromicro/go-zero/rest/httpx"
)

func SetCategoryHotZeroHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		l := logic.NewSetCategoryHotZeroLogic(r.Context(), svcCtx)
		resp, err := l.SetCategoryHotZero()
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}
