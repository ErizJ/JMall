// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package svc

import (
	"github.com/ErizJ/JMall/backend/cache"
	"github.com/ErizJ/JMall/backend/model"
	"github.com/ErizJ/JMall/backend/service/management/internal/config"
	"github.com/ErizJ/JMall/backend/service/management/internal/middleware"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
	"github.com/zeromicro/go-zero/rest"
)

type ServiceContext struct {
	Config                  config.Config
	Cache                   *cache.Client
	AuthMiddleware          rest.Middleware
	ProductModel            model.ProductModel
	CategoryModel           model.CategoryModel
	CombinationProductModel model.CombinationProductModel
	OrdersModel             model.OrdersModel
	UsersModel              model.UsersModel
	CarouselModel           model.CarouselModel
}

func NewServiceContext(c config.Config) *ServiceContext {
	conn := sqlx.NewMysql(c.DB.DataSource)
	return &ServiceContext{
		Config:                  c,
		Cache:                   cache.NewClient(c.Cache.Addr, c.Cache.Password, c.Cache.DB),
		AuthMiddleware:          middleware.NewAuthMiddleware(c.Auth.Secret).Handle,
		ProductModel:            model.NewProductModel(conn),
		CategoryModel:           model.NewCategoryModel(conn),
		CombinationProductModel: model.NewCombinationProductModel(conn),
		OrdersModel:             model.NewOrdersModel(conn),
		UsersModel:              model.NewUsersModel(conn),
		CarouselModel:           model.NewCarouselModel(conn),
	}
}
