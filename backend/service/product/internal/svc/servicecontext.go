// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package svc

import (
	"github.com/ErizJ/JMall/backend/cache"
	"github.com/ErizJ/JMall/backend/model"
	"github.com/ErizJ/JMall/backend/service/product/internal/config"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

type ServiceContext struct {
	Config              config.Config
	Cache               *cache.Client
	ProductModel        model.ProductModel
	CategoryModel       model.CategoryModel
	ProductPictureModel model.ProductPictureModel
	CarouselModel       model.CarouselModel
}

func NewServiceContext(c config.Config) *ServiceContext {
	conn := sqlx.NewMysql(c.DB.DataSource)
	return &ServiceContext{
		Config:              c,
		Cache:               cache.NewClient(c.Cache.Addr, c.Cache.Password, c.Cache.DB),
		ProductModel:        model.NewProductModel(conn),
		CategoryModel:       model.NewCategoryModel(conn),
		ProductPictureModel: model.NewProductPictureModel(conn),
		CarouselModel:       model.NewCarouselModel(conn),
	}
}
