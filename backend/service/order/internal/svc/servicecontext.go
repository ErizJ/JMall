// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package svc

import (
	"github.com/ErizJ/JMall/backend/cache"
	"github.com/ErizJ/JMall/backend/model"
	"github.com/ErizJ/JMall/backend/service/order/internal/config"
	"github.com/ErizJ/JMall/backend/service/order/internal/middleware"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
	"github.com/zeromicro/go-zero/rest"
)

type ServiceContext struct {
	Config            config.Config
	Cache             *cache.Client
	AuthMiddleware    rest.Middleware
	OrdersModel       model.OrdersModel
	ProductModel      model.ProductModel
	UsersModel        model.UsersModel
	ShoppingcartModel model.ShoppingcartModel
}

func NewServiceContext(c config.Config) *ServiceContext {
	conn := sqlx.NewMysql(c.DB.DataSource)
	return &ServiceContext{
		Config:            c,
		Cache:             cache.NewClient(c.Cache.Addr, c.Cache.Password, c.Cache.DB),
		AuthMiddleware:    middleware.NewAuthMiddleware(c.Auth.Secret).Handle,
		OrdersModel:       model.NewOrdersModel(conn),
		ProductModel:      model.NewProductModel(conn),
		UsersModel:        model.NewUsersModel(conn),
		ShoppingcartModel: model.NewShoppingcartModel(conn),
	}
}
