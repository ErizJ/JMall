package config

import "github.com/zeromicro/go-zero/rest"

type Config struct {
	rest.RestConf
	DB      DBConfig
	Auth    AuthConfig
	Cache   CacheConfig
	Payment PaymentConfig
}

type DBConfig struct {
	DataSource string
}

type AuthConfig struct {
	Secret        string
	ExpireSeconds int64
}

type CacheConfig struct {
	Addr     string
	Password string
	DB       int
}

type PaymentConfig struct {
	ExpireSeconds int64  // 支付单过期时间
	NotifyUrl     string // 回调通知地址
}
