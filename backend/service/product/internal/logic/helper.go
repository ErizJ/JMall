package logic

import (
	"github.com/ErizJ/JMall/backend/model"
	"github.com/ErizJ/JMall/backend/service/product/internal/types"
)

func productToItem(p *model.Product) types.ProductItem {
	picture := ""
	if p.ProductPicture.Valid {
		picture = p.ProductPicture.String
	}
	sales := int64(0)
	if p.ProductSales.Valid {
		sales = p.ProductSales.Int64
	}
	hot := int64(0)
	if p.ProductHot.Valid {
		hot = p.ProductHot.Int64
	}
	return types.ProductItem{
		ProductID:           p.ProductId,
		ProductName:         p.ProductName,
		CategoryID:          p.CategoryId,
		ProductTitle:        p.ProductTitle,
		ProductIntro:        p.ProductIntro,
		ProductPicture:      picture,
		ProductPrice:        p.ProductPrice,
		ProductSellingPrice: p.ProductSellingPrice,
		ProductNum:          p.ProductNum,
		ProductSales:        sales,
		ProductIsPromotion:  int(p.ProductIsPromotion),
		ProductHot:          hot,
	}
}

func productsToItems(products []*model.Product) []types.ProductItem {
	items := make([]types.ProductItem, 0, len(products))
	for _, p := range products {
		items = append(items, productToItem(p))
	}
	return items
}
