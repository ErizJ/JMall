package cache

import (
	"context"
	"encoding/json"
	"time"

	"github.com/redis/go-redis/v9"
)

type Client struct {
	rdb *redis.Client
}

func NewClient(addr, password string, db int) *Client {
	rdb := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: password,
		DB:       db,
	})
	return &Client{rdb: rdb}
}

func (c *Client) Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
	b, err := json.Marshal(value)
	if err != nil {
		return err
	}
	return c.rdb.Set(ctx, key, b, ttl).Err()
}

func (c *Client) Get(ctx context.Context, key string, dest interface{}) error {
	b, err := c.rdb.Get(ctx, key).Bytes()
	if err != nil {
		return err // redis.Nil means cache miss
	}
	return json.Unmarshal(b, dest)
}

func (c *Client) Del(ctx context.Context, keys ...string) error {
	return c.rdb.Del(ctx, keys...).Err()
}

func (c *Client) IsNil(err error) bool {
	return err == redis.Nil
}

// SetNX 设置 key（仅当 key 不存在时）。
// 如果 key 已存在，返回 error（用于幂等控制 / 分布式锁）。
func (c *Client) SetNX(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
	ok, err := c.rdb.SetNX(ctx, key, value, ttl).Result()
	if err != nil {
		return err
	}
	if !ok {
		return redis.Nil // key 已存在
	}
	return nil
}

// Eval 执行 Lua 脚本（原子操作）。
// 用于需要多步 Redis 操作保证原子性的场景，如库存扣减、分布式锁等。
func (c *Client) Eval(ctx context.Context, script string, keys []string, args ...interface{}) (interface{}, error) {
	return c.rdb.Eval(ctx, script, keys, args...).Result()
}

// Incr 对 key 做自增，返回自增后的值。
func (c *Client) Incr(ctx context.Context, key string) (int64, error) {
	return c.rdb.Incr(ctx, key).Result()
}

// Expire 设置 key 的过期时间。
func (c *Client) Expire(ctx context.Context, key string, ttl time.Duration) error {
	return c.rdb.Expire(ctx, key, ttl).Err()
}
