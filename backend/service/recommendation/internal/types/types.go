package types

// PromotionRule 满减规则
type PromotionRule struct {
	Threshold float64 `json:"threshold"`
	Reduction float64 `json:"reduction"`
}

// RecommendItem 推荐商品项
type RecommendItem struct {
	ProductID           int64   `json:"product_id"`
	ProductName         string  `json:"product_name"`
	CategoryID          int64   `json:"category_id"`
	ProductTitle        string  `json:"product_title"`
	ProductPicture      string  `json:"product_picture"`
	ProductPrice        float64 `json:"product_price"`
	ProductSellingPrice float64 `json:"product_selling_price"`
	ProductSales        int64   `json:"product_sales"`
	ProductHot          int64   `json:"product_hot"`
	RecommendReason     string  `json:"recommend_reason"`
	Score               float64 `json:"score"`
}

// FillUpReq 凑单推荐请求
type FillUpReq struct {
	UserID int64 `json:"user_id"`
}

// FillUpResp 凑单推荐响应
type FillUpResp struct {
	Code            string          `json:"code"`
	CartTotal       float64         `json:"cart_total"`
	NearestRule     PromotionRule   `json:"nearest_rule"`
	Gap             float64         `json:"gap"`
	Recommendations []RecommendItem `json:"recommendations,omitempty"`
}

// GuessYouLikeReq 猜你喜欢请求
type GuessYouLikeReq struct {
	Page     int `json:"page,optional"`
	PageSize int `json:"page_size,optional"`
}

// GuessYouLikeResp 猜你喜欢响应
type GuessYouLikeResp struct {
	Code            string          `json:"code"`
	Recommendations []RecommendItem `json:"recommendations"`
	HasMore         bool            `json:"has_more"`
}

// ReportBehaviorReq 上报用户行为请求
type ReportBehaviorReq struct {
	ProductID    int64 `json:"product_id"`
	CategoryID   int64 `json:"category_id"`
	BehaviorType int64 `json:"behavior_type"`
}

// ReportBehaviorResp 上报用户行为响应
type ReportBehaviorResp struct {
	Code string `json:"code"`
}
