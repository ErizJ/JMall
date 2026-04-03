package svc

import (
	"github.com/ErizJ/JMall/backend/cache"
	"github.com/ErizJ/JMall/backend/model"
	"github.com/ErizJ/JMall/backend/service/payment/internal/config"
	"github.com/ErizJ/JMall/backend/service/payment/internal/middleware"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
	"github.com/zeromicro/go-zero/rest"
)

type ServiceContext struct {
	Config             config.Config
	Cache              *cache.Client
	AuthMiddleware     rest.Middleware
	PaymentOrderModel  model.PaymentOrderModel
	PaymentRefundModel model.PaymentRefundModel
	OrdersModel        model.OrdersModel
	ProductModel       model.ProductModel
}

func NewServiceContext(c config.Config) *ServiceContext {
	conn := sqlx.NewMysql(c.DB.DataSource)
	return &ServiceContext{
		Config:             c,
		Cache:              cache.NewClient(c.Cache.Addr, c.Cache.Password, c.Cache.DB),
		AuthMiddleware:     middleware.NewAuthMiddleware(c.Auth.Secret).Handle,
		PaymentOrderModel:  model.NewPaymentOrderModel(conn),
		PaymentRefundModel: model.NewPaymentRefundModel(conn),
		OrdersModel:        model.NewOrdersModel(conn),
		ProductModel:       model.NewProductModel(conn),
	}
}
