package svc

import (
	"github.com/ErizJ/JMall/backend/cache"
	"github.com/ErizJ/JMall/backend/model"
	"github.com/ErizJ/JMall/backend/service/recommendation/internal/config"
	"github.com/ErizJ/JMall/backend/service/recommendation/internal/middleware"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
	"github.com/zeromicro/go-zero/rest"
)

type ServiceContext struct {
	Config                  config.Config
	Cache                   *cache.Client
	AuthMiddleware          rest.Middleware
	ProductModel            model.ProductModel
	ShoppingcartModel       model.ShoppingcartModel
	OrdersModel             model.OrdersModel
	CollectModel            model.CollectModel
	CombinationProductModel model.CombinationProductModel
	CategoryModel           model.CategoryModel
	UserBehaviorModel       model.UserBehaviorModel
	ProductSimilarityModel  model.ProductSimilarityModel
}

func NewServiceContext(c config.Config) *ServiceContext {
	conn := sqlx.NewMysql(c.DB.DataSource)
	return &ServiceContext{
		Config:                  c,
		Cache:                   cache.NewClient(c.Cache.Addr, c.Cache.Password, c.Cache.DB),
		AuthMiddleware:          middleware.NewAuthMiddleware(c.Auth.Secret).Handle,
		ProductModel:            model.NewProductModel(conn),
		ShoppingcartModel:       model.NewShoppingcartModel(conn),
		OrdersModel:             model.NewOrdersModel(conn),
		CollectModel:            model.NewCollectModel(conn),
		CombinationProductModel: model.NewCombinationProductModel(conn),
		CategoryModel:           model.NewCategoryModel(conn),
		UserBehaviorModel:       model.NewUserBehaviorModel(conn),
		ProductSimilarityModel:  model.NewProductSimilarityModel(conn),
	}
}
