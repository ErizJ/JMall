package main

import (
	"context"
	"flag"
	"fmt"

	"github.com/ErizJ/JMall/backend/kafka"
	"github.com/ErizJ/JMall/backend/model"
	"github.com/ErizJ/JMall/backend/service/seckill/internal/config"
	"github.com/ErizJ/JMall/backend/service/seckill/internal/consumer"
	"github.com/ErizJ/JMall/backend/service/seckill/internal/handler"
	"github.com/ErizJ/JMall/backend/service/seckill/internal/logic"
	"github.com/ErizJ/JMall/backend/service/seckill/internal/svc"

	"github.com/zeromicro/go-zero/core/conf"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
	"github.com/zeromicro/go-zero/rest"
)

var configFile = flag.String("f", "etc/seckill-api.yaml", "the config file")

func main() {
	flag.Parse()

	var c config.Config
	conf.MustLoad(*configFile, &c)

	server := rest.MustNewServer(c.RestConf)
	defer server.Stop()

	svcCtx := svc.NewServiceContext(c)
	handler.RegisterHandlers(server, svcCtx)

	// 预热：将活动数据加载到 Redis
	logic.WarmUp(context.Background(), svcCtx)

	// 启动 Kafka Consumer（后台 goroutine，随 server 生命周期）
	cancelConsumer := startConsumer(c, svcCtx)
	defer cancelConsumer()

	// 关闭 Kafka Producer
	defer func() {
		if err := svcCtx.KafkaProducer.Close(); err != nil {
			logx.Errorf("kafka producer close error: %v", err)
		}
	}()

	fmt.Printf("Starting seckill server at %s:%d...\n", c.Host, c.Port)
	server.Start()
}

// startConsumer 启动 Kafka 消费者，返回 cancel 函数用于优雅关闭
func startConsumer(c config.Config, svcCtx *svc.ServiceContext) func() {
	conn := sqlx.NewMysql(c.DB.DataSource)

	orderConsumer := consumer.NewSeckillOrderConsumer(
		svcCtx.Cache,
		model.NewOrdersModel(conn),
		model.NewProductModel(conn),
		model.NewSeckillOrderModel(conn),
		model.NewSeckillActivityModel(conn),
	)

	kafkaConsumer := kafka.NewConsumer(
		c.Kafka.Brokers,
		c.Kafka.SeckillOrderTopic,
		c.Kafka.ConsumerGroup,
	)

	ctx, cancel := context.WithCancel(context.Background())

	go func() {
		defer func() {
			if err := kafkaConsumer.Close(); err != nil {
				logx.Errorf("kafka consumer close error: %v", err)
			}
		}()
		logx.Info("seckill kafka consumer starting...")
		kafkaConsumer.Start(ctx, orderConsumer.Consume, orderConsumer.OnExhausted)
	}()

	return cancel
}
