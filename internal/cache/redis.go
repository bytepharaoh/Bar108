package cache

import (
	"bar108/config"
	"context"
	"fmt"
	"log"
	"time"

	"github.com/redis/go-redis/v9"
)

type Client struct {
	rdb *redis.Client
}

func New(cfg config.RedisConfig) *Client {
	rdb := redis.NewClient(&redis.Options{
		Addr:        fmt.Sprintf("%s:%s", cfg.Host, cfg.Port),
		PoolSize:    10,
		DialTimeout: 5 * time.Second,
		ReadTimeout: 3 * time.Second,
	})
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := rdb.Ping(ctx).Err(); err != nil {
		log.Fatalf("cache: failed to connect to Redis: %v", err)
	}
	log.Println("cache: connected to Redis successfully")
	return &Client{
		rdb: rdb,
	}

}
func (c *Client) Close() error {
	return c.rdb.Close()
}

// Token Blacklisting
// BlacklistToken stores a JWT ID in Redis so it can never
// be used again, even if it hasn't expired yet.
func (c *Client) BlacklistToken(ctx context.Context, jti string, ttl time.Duration) error {
	// SET key value EX ttl
	// Key format: "blacklist:<jti>"
	// Value: "1" (we only care if the key exists, not its value)
	return c.rdb.Set(ctx, blacklistKey(jti), "1", ttl).Err()
}
func (c *Client) IsTokenBlacklisted(ctx context.Context, jti string) (bool, error) {
	// EXISTS returns 1 if key exists, 0 if not

	result, err := c.rdb.Exists(ctx, blacklistKey(jti)).Result()
	if err != nil {
		return false, fmt.Errorf("cache: IsTokenBlacklisted: %w", err)
	}
	return result > 0, nil
}
func blacklistKey(jti string) string {
	return fmt.Sprintf("blacklist:%s", jti)
}

// Menu Caching

func (c *Client) GetMenu(ctx context.Context) (string, error) {

	val, err := c.rdb.Get(ctx, "menu:all").Result()
	if err == redis.Nil {
		// Cache miss — this is normal, not an error
		return "", nil
	}
	if err != nil {
		return "", fmt.Errorf("cache: GetMenu: %w", err)
	}

	return val, nil

}

// SetMenu stores the menu JSON in Redis for ttl duration.
func (c *Client) SetMenu(ctx context.Context, data string, ttl time.Duration) error {
	return c.rdb.Set(ctx, "menu:all", data, ttl).Err()
}

// InvalidateMenu removes the cached menu.
func (c *Client) InvalidateMenu(ctx context.Context) error {
	return c.rdb.Del(ctx, "menu:all").Err()
}

// Rate Limiting

func (c *Client) IncrementRateLimit(ctx context.Context, key string, window time.Duration) (int64, error) {

	pipe := c.rdb.Pipeline()
	// Increment the counter atomically
	incr := pipe.Incr(ctx, rateLimitKey(key))
	// Set expiry on first request — EXPIRE is ignored if key already has a TTL
	pipe.Expire(ctx, rateLimitKey(key), window)
	if _, err := pipe.Exec(ctx); err != nil {
		return 0, fmt.Errorf("cache: IncrementRateLimit: %w", err)
	}
	return incr.Val(), nil

}
func rateLimitKey(key string) string {
	return fmt.Sprintf("ratelimit:%s", key)
}
