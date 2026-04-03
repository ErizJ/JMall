package config

import "github.com/zeromicro/go-zero/rest"

type Config struct {
	rest.RestConf
	DB      DBConfig
	Auth    AuthConfig
	Cache   CacheConfig
	Doubao  DoubaoConfig
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

type DoubaoConfig struct {
	ApiKey  string
	Model   string
	BaseUrl string
}
