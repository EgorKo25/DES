package cache

import (
	"context"
	"time"
)

func NewCache(ctx context.Context, duration int) *Cache {

	c := &Cache{
		cache:    make(map[string]any),
		duration: time.Duration(duration) * time.Second,
	}

	go c.clearAll(ctx)

	return c
}

type Cache struct {
	cache    map[string]any
	duration time.Duration
}

func (c *Cache) Load(title string, data any) {
	if c.cache != nil {
		c.cache[title] = data
	}
}

func (c *Cache) Search(title string) (any, bool) {
	data, ok := c.cache[title]
	return data, ok
}

func (c *Cache) clearAll(ctx context.Context) {
	ticker := time.NewTicker(c.duration)
	for {
		select {
		case <-ticker.C:
			c.cache = make(map[string]any)
		case <-ctx.Done():
			c.cache = make(map[string]any)
			return
		}
	}
}
