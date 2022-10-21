package service

import (
	"github.com/eko/gocache/v3/cache"
	"go.mongodb.org/mongo-driver/mongo"
)

type Option func(service *UrlShorteningService)

func WithMultiCache(cache *cache.ChainCache[string]) Option {
	return func(service *UrlShorteningService) {
		service.cache = cache
	}
}

func WithMongoDB(mongo *mongo.Client) Option {
	return func(service *UrlShorteningService) {
		service.mongo = mongo
	}
}
