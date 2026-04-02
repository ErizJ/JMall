package config

import "github.com/zeromicro/go-zero/rest"

// ServiceConfig holds the common configuration for all services.
type ServiceConfig struct {
	rest.RestConf
	DB DBConfig
	Auth AuthConfig
}

type DBConfig struct {
	DataSource string
}

type AuthConfig struct {
	// Secret is used for JWT signing.
	Secret string
	// ExpireSeconds is the JWT expiry duration.
	ExpireSeconds int64
}
