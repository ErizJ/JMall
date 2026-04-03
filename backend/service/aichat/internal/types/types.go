package types

type ChatReq struct {
	Message string `json:"message"`
}

type ChatResp struct {
	Code  string `json:"code"`
	Reply string `json:"reply"`
}
