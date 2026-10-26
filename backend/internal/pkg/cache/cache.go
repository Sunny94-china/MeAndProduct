package cache

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/redis/go-redis/v9"
)

// nullPlaceholder 用于防止缓存穿透
const nullPlaceholder = "__NULL__"

var ErrNotFound = errors.New("cache: not found")

// GetOrSet 读穿透缓存。loader 返回 nil,nil 时存 null 占位并返回 nil。
func GetOrSet[T any](
	ctx context.Context,
	rd *redis.Client,
	key string,
	ttl time.Duration,
	loader func() (*T, error),
) (*T, error) {
	raw, err := rd.Get(ctx, key).Result()
	if err == nil {
		if raw == nullPlaceholder {
			return nil, nil
		}
		var v T
		if e := json.Unmarshal([]byte(raw), &v); e == nil {
			return &v, nil
		}
		// 反序列化失败继续 loader 回源
	}
	v, err := loader()
	if err != nil {
		return nil, err
	}
	if v == nil {
		// 存空占位，防穿透
		rd.Set(ctx, key, nullPlaceholder, 30*time.Second)
		return nil, nil
	}
	b, e := json.Marshal(v)
	if e == nil {
		rd.Set(ctx, key, b, ttl)
	}
	return v, nil
}

// Invalidate 删除若干缓存键
func Invalidate(ctx context.Context, rd *redis.Client, keys ...string) error {
	if len(keys) == 0 {
		return nil
	}
	n, err := rd.Del(ctx, keys...).Result()
	_ = n
	return err
}
