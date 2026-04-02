// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package config

import "github.com/zeromicro/go-zero/rest"

type Config struct {
	rest.RestConf
	DB    DBConfig
	Auth  AuthConfig
	Cache CacheConfig
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
