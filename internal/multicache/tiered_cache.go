package multicache

import (
	"time"

	"gately/internal/app"
	"github.com/dgraph-io/ristretto"
	"github.com/eko/gocache/v3/cache"
	"github.com/eko/gocache/v3/store"
	"github.com/go-redis/redis/v8"
)

func New(cfg app.Config) *cache.ChainCache[string] {
	// Ristretto is our in-memory Layer-1 LRU multicache
	// The least frequently accessed sites will be evicted first
	lruCache, err := ristretto.NewCache(
		&ristretto.Config{
			NumCounters: 10000,
			MaxCost:     1000,
			BufferItems: 128},
	)

	if err != nil {
		// Ok to panic as we are still in application bootstrap
		panic(err)
	}

	redisClient := redis.NewClient(&redis.Options{
		Addr:     cfg.RedisHost, // host:port of the redis server
		Password: cfg.RedisPass, // no password set for demo purposes
		DB:       0,             // use default DB
	})

	// Initialize stores
	ristrettoStore := store.NewRistretto(lruCache)
	redisStore := store.NewRedis(redisClient, store.WithExpiration(5*time.Second))

	// Initialize our tiered multicache
	cacheManager := cache.NewChain[string](
		cache.New[string](ristrettoStore),
		cache.New[string](redisStore),
	)

	return cacheManager
}
