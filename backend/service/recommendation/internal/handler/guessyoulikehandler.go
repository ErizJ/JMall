package handler

import (
	"net/http"

	"github.com/ErizJ/JMall/backend/service/recommendation/internal/logic"
	"github.com/ErizJ/JMall/backend/service/recommendation/internal/svc"
	"github.com/ErizJ/JMall/backend/service/recommendation/internal/types"
	"github.com/zeromicro/go-zero/rest/httpx"
)

func GuessYouLikeHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.GuessYouLikeReq
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		l := logic.NewGuessYouLikeLogic(r.Context(), svcCtx)
		resp, err := l.GuessYouLike(&req)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}
