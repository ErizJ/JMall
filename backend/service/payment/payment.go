package main

import (
	"flag"
	"fmt"

	"github.com/ErizJ/JMall/backend/service/payment/internal/config"
	"github.com/ErizJ/JMall/backend/service/payment/internal/handler"
	"github.com/ErizJ/JMall/backend/service/payment/internal/svc"

	// 注册支付渠道（init 函数自动注册）
	_ "github.com/ErizJ/JMall/backend/service/payment/internal/channel"

	"github.com/zeromicro/go-zero/core/conf"
	"github.com/zeromicro/go-zero/rest"
)

var configFile = flag.String("f", "etc/payment-api.yaml", "the config file")

func main() {
	flag.Parse()

	var c config.Config
	conf.MustLoad(*configFile, &c)

	server := rest.MustNewServer(c.RestConf)
	defer server.Stop()

	ctx := svc.NewServiceContext(c)
	handler.RegisterHandlers(server, ctx)

	fmt.Printf("Starting payment server at %s:%d...\n", c.Host, c.Port)
	server.Start()
}
