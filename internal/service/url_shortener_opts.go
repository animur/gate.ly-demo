package service

import (
	"gately/internal/dal"
	"github.com/eko/gocache/v3/cache"
)

type Option func(service *UrlShorteningService)

func WithMultiCache(cache *cache.ChainCache[string]) Option {
	return func(service *UrlShorteningService) {
		service.cache = cache
	}
}

func WithUrlStore(store dal.UrlStore) Option {
	return func(service *UrlShorteningService) {
		service.store = store
	}
}
