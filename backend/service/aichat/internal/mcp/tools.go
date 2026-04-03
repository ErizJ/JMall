// Package mcp implements MCP-style tool definitions that the AI model can invoke
// to query product, category, and discount information from the database.
package mcp

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/ErizJ/JMall/backend/service/aichat/internal/svc"
)

// ToolDef describes a tool the LLM can call.
type ToolDef struct {
	Type     string      `json:"type"`
	Function FunctionDef `json:"function"`
}

type FunctionDef struct {
	Name        string          `json:"name"`
	Description string          `json:"description"`
	Parameters  json.RawMessage `json:"parameters"`
}

// ToolCall represents a parsed tool invocation from the model.
type ToolCall struct {
	ID       string `json:"id"`
	Type     string `json:"type"`
	Function struct {
		Name      string `json:"name"`
		Arguments string `json:"arguments"`
	} `json:"function"`
}

// GetToolDefinitions returns the MCP tool schemas for the LLM.
func GetToolDefinitions() []ToolDef {
	return []ToolDef{
		{
			Type: "function",
			Function: FunctionDef{
				Name:        "search_products",
				Description: "根据关键词搜索商品，返回商品名称、价格、库存等信息",
				Parameters:  json.RawMessage(`{"type":"object","properties":{"keyword":{"type":"string","description":"搜索关键词，如手机、电视等"}},"required":["keyword"]}`),
			},
		},
		{
			Type: "function",
			Function: FunctionDef{
				Name:        "get_categories",
				Description: "获取所有商品分类列表",
				Parameters:  json.RawMessage(`{"type":"object","properties":{}}`),
			},
		},
		{
			Type: "function",
			Function: FunctionDef{
				Name:        "get_product_detail",
				Description: "根据商品ID获取商品详细信息，包括价格、库存、销量等",
				Parameters:  json.RawMessage(`{"type":"object","properties":{"product_id":{"type":"integer","description":"商品ID"}},"required":["product_id"]}`),
			},
		},
		{
			Type: "function",
			Function: FunctionDef{
				Name:        "get_products_by_category",
				Description: "根据分类ID获取该分类下的所有商品",
				Parameters:  json.RawMessage(`{"type":"object","properties":{"category_id":{"type":"integer","description":"分类ID"}},"required":["category_id"]}`),
			},
		},
		{
			Type: "function",
			Function: FunctionDef{
				Name:        "get_hot_products",
				Description: "获取热门商品排行，返回最受欢迎的商品列表",
				Parameters:  json.RawMessage(`{"type":"object","properties":{"limit":{"type":"integer","description":"返回数量，默认10"}}}`),
			},
		},
		{
			Type: "function",
			Function: FunctionDef{
				Name:        "get_promotion_products",
				Description: "获取当前促销/打折商品列表",
				Parameters:  json.RawMessage(`{"type":"object","properties":{"limit":{"type":"integer","description":"返回数量，默认10"}}}`),
			},
		},
		{
			Type: "function",
			Function: FunctionDef{
				Name:        "get_combination_discounts",
				Description: "获取组合优惠/满减活动信息，包括主商品、副商品、满减门槛和减免金额",
				Parameters:  json.RawMessage(`{"type":"object","properties":{}}`),
			},
		},
	}
}

// ExecuteTool runs the named tool and returns a JSON string result.
func ExecuteTool(ctx context.Context, svcCtx *svc.ServiceContext, name string, argsJSON string) (string, error) {
	switch name {
	case "search_products":
		return execSearchProducts(ctx, svcCtx, argsJSON)
	case "get_categories":
		return execGetCategories(ctx, svcCtx)
	case "get_product_detail":
		return execGetProductDetail(ctx, svcCtx, argsJSON)
	case "get_products_by_category":
		return execGetProductsByCategory(ctx, svcCtx, argsJSON)
	case "get_hot_products":
		return execGetHotProducts(ctx, svcCtx, argsJSON)
	case "get_promotion_products":
		return execGetPromotionProducts(ctx, svcCtx, argsJSON)
	case "get_combination_discounts":
		return execGetCombinationDiscounts(ctx, svcCtx)
	default:
		return "", fmt.Errorf("unknown tool: %s", name)
	}
}

func execSearchProducts(ctx context.Context, svcCtx *svc.ServiceContext, argsJSON string) (string, error) {
	var args struct {
		Keyword string `json:"keyword"`
	}
	if err := json.Unmarshal([]byte(argsJSON), &args); err != nil {
		return "", err
	}
	products, err := svcCtx.ProductModel.FindBySearch(ctx, args.Keyword)
	if err != nil {
		return "", err
	}
	type item struct {
		ID           int64   `json:"id"`
		Name         string  `json:"name"`
		Price        float64 `json:"price"`
		SellingPrice float64 `json:"selling_price"`
		Stock        int64   `json:"stock"`
		Sales        int64   `json:"sales"`
		IsPromotion  int64   `json:"is_promotion"`
	}
	items := make([]item, 0, len(products))
	for _, p := range products {
		items = append(items, item{
			ID: p.ProductId, Name: p.ProductName,
			Price: p.ProductPrice, SellingPrice: p.ProductSellingPrice,
			Stock: p.ProductNum, Sales: nullInt64Val(p.ProductSales),
			IsPromotion: p.ProductIsPromotion,
		})
	}
	b, _ := json.Marshal(items)
	return string(b), nil
}

func execGetCategories(ctx context.Context, svcCtx *svc.ServiceContext) (string, error) {
	cats, err := svcCtx.CategoryModel.FindAll(ctx)
	if err != nil {
		return "", err
	}
	type item struct {
		ID   int64  `json:"id"`
		Name string `json:"name"`
	}
	items := make([]item, 0, len(cats))
	for _, c := range cats {
		items = append(items, item{ID: c.CategoryId, Name: c.CategoryName})
	}
	b, _ := json.Marshal(items)
	return string(b), nil
}

func execGetProductDetail(ctx context.Context, svcCtx *svc.ServiceContext, argsJSON string) (string, error) {
	var args struct {
		ProductID int64 `json:"product_id"`
	}
	if err := json.Unmarshal([]byte(argsJSON), &args); err != nil {
		return "", err
	}
	// support string or number
	if args.ProductID == 0 {
		var raw map[string]interface{}
		json.Unmarshal([]byte(argsJSON), &raw)
		if v, ok := raw["product_id"]; ok {
			switch t := v.(type) {
			case string:
				args.ProductID, _ = strconv.ParseInt(t, 10, 64)
			case float64:
				args.ProductID = int64(t)
			}
		}
	}
	p, err := svcCtx.ProductModel.FindOne(ctx, args.ProductID)
	if err != nil {
		return "", err
	}
	detail := map[string]interface{}{
		"id": p.ProductId, "name": p.ProductName,
		"title": p.ProductTitle, "intro": p.ProductIntro,
		"price": p.ProductPrice, "selling_price": p.ProductSellingPrice,
		"stock": p.ProductNum, "sales": nullInt64Val(p.ProductSales),
		"is_promotion": p.ProductIsPromotion,
	}
	b, _ := json.Marshal(detail)
	return string(b), nil
}

func execGetProductsByCategory(ctx context.Context, svcCtx *svc.ServiceContext, argsJSON string) (string, error) {
	var args struct {
		CategoryID int64 `json:"category_id"`
	}
	if err := json.Unmarshal([]byte(argsJSON), &args); err != nil {
		return "", err
	}
	products, err := svcCtx.ProductModel.FindByCategory(ctx, args.CategoryID)
	if err != nil {
		return "", err
	}
	type item struct {
		ID           int64   `json:"id"`
		Name         string  `json:"name"`
		Price        float64 `json:"price"`
		SellingPrice float64 `json:"selling_price"`
		Stock        int64   `json:"stock"`
	}
	items := make([]item, 0, len(products))
	for _, p := range products {
		items = append(items, item{
			ID: p.ProductId, Name: p.ProductName,
			Price: p.ProductPrice, SellingPrice: p.ProductSellingPrice,
			Stock: p.ProductNum,
		})
	}
	b, _ := json.Marshal(items)
	return string(b), nil
}

func execGetHotProducts(ctx context.Context, svcCtx *svc.ServiceContext, argsJSON string) (string, error) {
	var args struct {
		Limit int `json:"limit"`
	}
	json.Unmarshal([]byte(argsJSON), &args)
	if args.Limit <= 0 {
		args.Limit = 10
	}
	products, err := svcCtx.ProductModel.FindTopHot(ctx, args.Limit)
	if err != nil {
		return "", err
	}
	type item struct {
		ID    int64   `json:"id"`
		Name  string  `json:"name"`
		Price float64 `json:"selling_price"`
		Hot   int64   `json:"hot"`
	}
	items := make([]item, 0, len(products))
	for _, p := range products {
		items = append(items, item{
			ID: p.ProductId, Name: p.ProductName,
			Price: p.ProductSellingPrice, Hot: nullInt64Val(p.ProductHot),
		})
	}
	b, _ := json.Marshal(items)
	return string(b), nil
}

func execGetPromotionProducts(ctx context.Context, svcCtx *svc.ServiceContext, argsJSON string) (string, error) {
	var args struct {
		Limit int `json:"limit"`
	}
	json.Unmarshal([]byte(argsJSON), &args)
	if args.Limit <= 0 {
		args.Limit = 10
	}
	products, err := svcCtx.ProductModel.FindByIsPromotion(ctx, args.Limit)
	if err != nil {
		return "", err
	}
	type item struct {
		ID           int64   `json:"id"`
		Name         string  `json:"name"`
		OrigPrice    float64 `json:"original_price"`
		SellingPrice float64 `json:"selling_price"`
	}
	items := make([]item, 0, len(products))
	for _, p := range products {
		items = append(items, item{
			ID: p.ProductId, Name: p.ProductName,
			OrigPrice: p.ProductPrice, SellingPrice: p.ProductSellingPrice,
		})
	}
	b, _ := json.Marshal(items)
	return string(b), nil
}

func execGetCombinationDiscounts(ctx context.Context, svcCtx *svc.ServiceContext) (string, error) {
	combos, err := svcCtx.CombinationProductModel.FindAll(ctx)
	if err != nil {
		return "", err
	}

	// 收集所有需要查询的商品 ID，一次批量查询消除 N+1
	idSet := make(map[int64]bool, len(combos)*2)
	for _, c := range combos {
		idSet[c.MainProductId] = true
		idSet[c.ViceProductId] = true
	}
	allIDs := make([]int64, 0, len(idSet))
	for id := range idSet {
		allIDs = append(allIDs, id)
	}
	productMap := make(map[int64]string, len(allIDs))
	if len(allIDs) > 0 {
		products, fetchErr := svcCtx.ProductModel.FindByIds(ctx, allIDs)
		if fetchErr == nil {
			for _, p := range products {
				productMap[p.ProductId] = p.ProductName
			}
		}
	}

	type comboItem struct {
		MainProductID       int64  `json:"main_product_id"`
		MainProductName     string `json:"main_product_name"`
		ViceProductID       int64  `json:"vice_product_id"`
		ViceProductName     string `json:"vice_product_name"`
		AmountThreshold     int64  `json:"amount_threshold"`
		PriceReductionRange int64  `json:"price_reduction"`
	}
	items := make([]comboItem, 0, len(combos))
	for _, c := range combos {
		items = append(items, comboItem{
			MainProductID:       c.MainProductId,
			MainProductName:     productMap[c.MainProductId],
			ViceProductID:       c.ViceProductId,
			ViceProductName:     productMap[c.ViceProductId],
			AmountThreshold:     nullInt64Val(c.AmountThreshold),
			PriceReductionRange: nullInt64Val(c.PriceReductionRange),
		})
	}
	b, _ := json.Marshal(items)
	return string(b), nil
}

func nullInt64Val(n sql.NullInt64) int64 {
	if n.Valid {
		return n.Int64
	}
	return 0
}
