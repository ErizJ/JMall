package svc

import (
	"github.com/ErizJ/JMall/backend/cache"
	"github.com/ErizJ/JMall/backend/kafka"
	"github.com/ErizJ/JMall/backend/model"
	"github.com/ErizJ/JMall/backend/service/seckill/internal/config"
	"github.com/ErizJ/JMall/backend/service/seckill/internal/middleware"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
	"github.com/zeromicro/go-zero/rest"
)

type ServiceContext struct {
	Config                config.Config
	Cache                 *cache.Client
	AuthMiddleware        rest.Middleware
	SeckillActivityModel  model.SeckillActivityModel
	SeckillOrderModel     model.SeckillOrderModel
	OrdersModel           model.OrdersModel
	ProductModel          model.ProductModel
	KafkaProducer         *kafka.Producer
}

func NewServiceContext(c config.Config) *ServiceContext {
	conn := sqlx.NewMysql(c.DB.DataSource)
	return &ServiceContext{
		Config:                c,
		Cache:                 cache.NewClient(c.Cache.Addr, c.Cache.Password, c.Cache.DB),
		AuthMiddleware:        middleware.NewAuthMiddleware(c.Auth.Secret).Handle,
		SeckillActivityModel:  model.NewSeckillActivityModel(conn),
		SeckillOrderModel:     model.NewSeckillOrderModel(conn),
		OrdersModel:           model.NewOrdersModel(conn),
		ProductModel:          model.NewProductModel(conn),
		KafkaProducer:         kafka.MustNewProducer(c.Kafka.Brokers, c.Kafka.SeckillOrderTopic),
	}
}
