package config

import "github.com/zeromicro/go-zero/rest"

type Config struct {
	rest.RestConf
	DB    DBConfig
	Auth  AuthConfig
	Cache CacheConfig
	Kafka KafkaConfig
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

type KafkaConfig struct {
	Brokers           []string
	SeckillOrderTopic string
	ConsumerGroup     string
}
