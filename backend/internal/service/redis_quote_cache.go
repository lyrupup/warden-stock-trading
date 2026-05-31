package service

import (
	"context"
	"encoding/json"
	"time"

	"github.com/redis/go-redis/v9"

	"warden/internal/model"
)

// redisQuoteCache 是基于 Redis 的行情缓存实现（key: warden:market:quote:{code}）。
// TTL 可随交易时段调整（见 BACKEND.md §5 M1 数据流）。
type redisQuoteCache struct {
	cli *redis.Client
	ttl time.Duration
}

// NewRedisQuoteCache 构造 Redis 行情缓存。
func NewRedisQuoteCache(cli *redis.Client, ttl time.Duration) QuoteCache {
	return &redisQuoteCache{cli: cli, ttl: ttl}
}

func quoteKey(code string) string { return "warden:market:quote:" + code }

func (c *redisQuoteCache) GetQuotes(ctx context.Context, codes []string) ([]model.StockQuote, []string, error) {
	if len(codes) == 0 {
		return nil, nil, nil
	}
	keys := make([]string, len(codes))
	for i, code := range codes {
		keys[i] = quoteKey(code)
	}
	vals, err := c.cli.MGet(ctx, keys...).Result()
	if err != nil {
		return nil, codes, err
	}
	hits := make([]model.StockQuote, 0, len(codes))
	missing := make([]string, 0)
	for i, v := range vals {
		s, ok := v.(string)
		if !ok || s == "" {
			missing = append(missing, codes[i])
			continue
		}
		var q model.StockQuote
		if err := json.Unmarshal([]byte(s), &q); err != nil {
			missing = append(missing, codes[i])
			continue
		}
		hits = append(hits, q)
	}
	return hits, missing, nil
}

func (c *redisQuoteCache) SetQuotes(ctx context.Context, quotes []model.StockQuote) error {
	if len(quotes) == 0 {
		return nil
	}
	pipe := c.cli.Pipeline()
	for _, q := range quotes {
		b, err := json.Marshal(q)
		if err != nil {
			continue
		}
		pipe.Set(ctx, quoteKey(q.StockCode), b, c.ttl)
	}
	_, err := pipe.Exec(ctx)
	return err
}
