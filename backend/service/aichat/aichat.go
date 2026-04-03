package main

import (
	"flag"
	"fmt"

	"github.com/ErizJ/JMall/backend/service/aichat/internal/config"
	"github.com/ErizJ/JMall/backend/service/aichat/internal/handler"
	"github.com/ErizJ/JMall/backend/service/aichat/internal/svc"

	"github.com/zeromicro/go-zero/core/conf"
	"github.com/zeromicro/go-zero/rest"
)

var configFile = flag.String("f", "etc/aichat-api.yaml", "the config file")

func main() {
	flag.Parse()

	var c config.Config
	conf.MustLoad(*configFile, &c)

	server := rest.MustNewServer(c.RestConf, rest.WithCors())
	defer server.Stop()

	ctx := svc.NewServiceContext(c)
	handler.RegisterHandlers(server, ctx)

	fmt.Printf("Starting aichat server at %s:%d...\n", c.Host, c.Port)
	server.Start()
}
