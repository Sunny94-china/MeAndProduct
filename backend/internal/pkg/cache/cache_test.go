package cache

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
)

func newMini(t *testing.T) *redis.Client {
	t.Helper()
	mr := miniredis.RunT(t)
	return redis.NewClient(&redis.Options{Addr: mr.Addr()})
}

type Dummy struct {
	Name string `json:"name"`
	Age  int    `json:"age"`
}

func TestGetOrSet_HitLoader(t *testing.T) {
	rd := newMini(t)
	ctx := context.Background()
	var called int
	v, err := GetOrSet[Dummy](ctx, rd, "k1", time.Minute, func() (*Dummy, error) {
		called++
		return &Dummy{Name: "me", Age: 30}, nil
	})
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if v == nil || v.Name != "me" {
		t.Fatalf("wrong value: %+v", v)
	}
	// 第二次命中缓存，loader 不应再调用
	v2, err := GetOrSet[Dummy](ctx, rd, "k1", time.Minute, func() (*Dummy, error) {
		called++
		return &Dummy{Name: "other"}, nil
	})
	if err != nil {
		t.Fatal(err)
	}
	if called != 1 {
		t.Errorf("loader called %d times, expected 1", called)
	}
	if v2.Name != "me" {
		t.Errorf("expected cache hit, got name=%s", v2.Name)
	}
}

func TestGetOrSet_NilLoader(t *testing.T) {
	rd := newMini(t)
	ctx := context.Background()
	v, err := GetOrSet[Dummy](ctx, rd, "k_null", time.Minute, func() (*Dummy, error) {
		return nil, nil
	})
	if err != nil {
		t.Fatal(err)
	}
	if v != nil {
		t.Errorf("expected nil, got %+v", v)
	}
	// 第二次应命中空占位，防止穿透
	var called int
	_, _ = GetOrSet[Dummy](ctx, rd, "k_null", time.Minute, func() (*Dummy, error) {
		called++
		return nil, nil
	})
	if called != 0 {
		t.Errorf("null placeholder not respected: called=%d", called)
	}
}

func TestInvalidate(t *testing.T) {
	rd := newMini(t)
	ctx := context.Background()
	_, _ = GetOrSet[Dummy](ctx, rd, "x1", time.Minute, func() (*Dummy, error) {
		return &Dummy{Name: "x1"}, nil
	})
	_, _ = GetOrSet[Dummy](ctx, rd, "x2", time.Minute, func() (*Dummy, error) {
		return &Dummy{Name: "x2"}, nil
	})
	if err := Invalidate(ctx, rd, "x1", "x2"); err != nil {
		t.Fatal(err)
	}
	var called int
	_, _ = GetOrSet[Dummy](ctx, rd, "x1", time.Minute, func() (*Dummy, error) {
		called++
		return &Dummy{Name: "new"}, nil
	})
	fmt.Println("called=", called)
	if called != 1 {
		t.Errorf("x1 should be invalidated, called=%d", called)
	}
}
