package driver

import (
	"context"
	"log/slog"

	"github.com/dgraph-io/ristretto"
	"github.com/eko/gocache/lib/v4/cache"
	"github.com/eko/gocache/lib/v4/store"
	store_redis "github.com/eko/gocache/store/redis/v4"
	store_memory "github.com/eko/gocache/store/ristretto/v4"
	"github.com/gowok/gowok/config"
	"github.com/gowok/gowok/must"
	"github.com/gowok/gowok/some"
	"github.com/redis/go-redis/v9"
)

type Cache map[string]store.StoreInterface

func NewCache(config map[string]config.Cache) (Cache, error) {
	caches := Cache{}

	for name, dbC := range config {
		if !dbC.Enabled {
			continue
		}

		var client store.StoreInterface
		if dbC.Driver == "memory" {
			clientOpt := must.Must(ristretto.NewCache(&ristretto.Config{}))
			client = store_memory.NewRistretto(clientOpt)
		} else if dbC.Driver == "redis" {
			clientOpt := must.Must(redis.ParseURL(dbC.DSN))
			client = store_redis.NewRedis(redis.NewClient(clientOpt))
		}
		caches[name] = client
	}

	return caches, nil
}

func (d Cache) Get(name ...string) some.Some[*cache.Cache[any]] {
	n := ""
	if len(name) > 0 {
		n = name[0]
		if db, ok := d[n]; ok {
			c := cache.New[any](db)
			return some.Of(&c)
		}
	}

	if n != "" {
		slog.Info("using default connection", "not_found", n)
	}

	if db, ok := d["default"]; ok {
		c := cache.New[any](db)
		return some.Of(&c)
	}

	return some.Empty[*cache.Cache[any]]()
}

type KVClient interface {
	PingContext(ctx context.Context) error
	Close() error
}

type KVReader interface {
	GetContext(ctx context.Context, key string) (any, error)
}

type KVWriter interface {
	SetContext(ctx context.Context, key string, value any) error
}
