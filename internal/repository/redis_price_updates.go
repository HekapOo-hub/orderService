package repository

import (
	"context"
	"fmt"
	"github.com/HekapOo-hub/orderService/internal/config"
	"github.com/HekapOo-hub/orderService/internal/model"
	"github.com/go-redis/redis"
	log "github.com/sirupsen/logrus"
	"sync"
)

type PriceUpdates interface {
	Get() (map[string]model.GeneratedPrice, error)
}

type RedisPriceUpdates struct {
	flag   chan bool
	client *redis.Client
	cache  map[string]model.GeneratedPrice
	mu     sync.RWMutex
}

func NewRedisPriceUpdates(ctx context.Context) (*RedisPriceUpdates, error) {
	redisCfg, err := config.NewRedisConfig()
	if err != nil {
		return nil, fmt.Errorf("new order service: %w", err)
	}

	redisClient := redis.NewClient(&redis.Options{
		Addr:     redisCfg.Addr,
		Password: redisCfg.Password,
		DB:       redisCfg.DB,
	})
	priceUpdates := &RedisPriceUpdates{client: redisClient,
		cache: make(map[string]model.GeneratedPrice), mu: sync.RWMutex{}, flag: make(chan bool)}
	go priceUpdates.listen(ctx)
	return priceUpdates, nil
}

func (r *RedisPriceUpdates) listen(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		default:
			res, err := r.client.XRead(&redis.XReadArgs{
				Block:   0,
				Count:   1,
				Streams: []string{config.RedisStream, "$"},
			}).Result()
			if err != nil {
				log.Warnf("redis price updates read stream error:%v", err)
				continue
			}
			if res[0].Messages == nil {
				log.Warn("message is empty")
				continue
			}
			pricesMap := res[0].Messages[0].Values
			for key, val := range pricesMap {
				price, err := model.DecodePrice([]byte(val.(string)))
				if err != nil {
					log.Warnf("redis price updates listen: %v", err)
					break
				}
				r.mu.Lock()
				r.cache[key] = price
				r.mu.Unlock()
			}
			r.flag <- true
		}
	}
}

func (r *RedisPriceUpdates) Get() (map[string]model.GeneratedPrice, error) {
	<-r.flag
	r.mu.RLock()
	defer r.mu.RUnlock()
	if len(r.cache) == 0 {
		return nil, fmt.Errorf("get redis price updates: map is empty")
	}
	prices := make(map[string]model.GeneratedPrice)
	for key, val := range r.cache {
		prices[key] = val
	}
	return prices, nil
}
