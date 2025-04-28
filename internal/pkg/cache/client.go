package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/bradfitz/gomemcache/memcache"
	"gitlab.ozon.dev/dimabelunin7/homework/hw-4/internal/config"
	"gitlab.ozon.dev/dimabelunin7/homework/hw-4/internal/domain"
	"gitlab.ozon.dev/dimabelunin7/homework/hw-4/internal/pkg/repository"
	"log"
	"time"
)

type Client struct {
	mc  *memcache.Client
	ttl int32
}

func NewCacheClient(cfg *config.Config, ttl int32) *Client {
	return &Client{
		mc:  memcache.New(cfg.MemCacheHost),
		ttl: ttl,
	}
}

func (c *Client) StartPeriodicUpdate(ctx context.Context, interval time.Duration, repo interface {
	FindAll(ctx context.Context, filter repository.Filter, lastId *int64, limit *int) ([]domain.Order, error)
}) {
	ticker := time.NewTicker(time.Second * interval)
	go func() {
		for {
			select {
			case <-ctx.Done():
				ticker.Stop()
				return
			case <-ticker.C:
				orders, err := repo.FindAll(ctx, repository.Filter{}, nil, nil)
				if err != nil {
					log.Printf("periodic cache update failed: %v", err)
					continue
				}

				filter := repository.Filter{}
				filterString := filter.GetFilterStringView()
				base := fmt.Sprintf("findAll:%v", filterString)
				if err := c.SetOrdersToCache(base, orders); err != nil {
					log.Printf("failed to update cache: %v", err)
				} else {
					log.Printf("history orders cache updated :)")
				}
			}
		}
	}()
}

func (c *Client) InvalidateOrderCache(cacheKey string) {
	_ = c.mc.Delete(cacheKey)
}

func (c *Client) GetOrdersFromCache(key string) ([]domain.Order, error) {
	data, err := c.get(key)
	if err != nil {
		return nil, err
	}

	var orders []domain.Order
	if err := json.Unmarshal(data, &orders); err != nil {
		return nil, err
	}

	return orders, nil
}

func (c *Client) SetOrdersToCache(key string, orders []domain.Order) error {
	data, err := json.Marshal(orders)
	if err != nil {
		return err
	}

	return c.set(key, data)
}

func (c *Client) get(key string) ([]byte, error) {
	item, err := c.mc.Get(key)
	if err != nil {
		return nil, err
	}
	return item.Value, nil
}

func (c *Client) set(key string, value []byte) error {
	return c.mc.Set(&memcache.Item{
		Key:        key,
		Value:      value,
		Expiration: c.ttl,
	})
}
