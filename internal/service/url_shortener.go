package service

import (
	"context"

	"github.com/eko/gocache/v3/cache"
	"go.mongodb.org/mongo-driver/mongo"
)

type UrlShortener interface {
	CreateUrlMapping(ctx context.Context, url string) (string, error)
	DeleteUrlMapping(ctx context.Context, url string) error
	RedirectUrl(ctx context.Context, shortUrl string) error
}

type UrlShorteningService struct {
	UrlShortener
	cache *cache.ChainCache[string]
	mongo *mongo.Client
}

func New(opts ...Option) *UrlShorteningService {

	service := &UrlShorteningService{}
	for _, opt := range opts {
		opt(service)
	}
	return service
}
