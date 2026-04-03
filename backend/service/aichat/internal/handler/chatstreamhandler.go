package handler

import (
	"net/http"

	"github.com/ErizJ/JMall/backend/service/aichat/internal/logic"
	"github.com/ErizJ/JMall/backend/service/aichat/internal/svc"
	"github.com/ErizJ/JMall/backend/service/aichat/internal/types"
	"github.com/zeromicro/go-zero/rest/httpx"
)

func ChatStreamHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.ChatReq
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		l := logic.NewChatStreamLogic(r.Context(), svcCtx)
		l.ChatStream(&req, w, r)
	}
}
