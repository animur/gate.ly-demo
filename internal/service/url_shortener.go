package service

import (
	"context"

	"gately/internal/dal"
	"github.com/eko/gocache/v3/cache"
)

type UrlShortener interface {
	CreateUrlMapping(ctx context.Context, url string) (string, error)
	DeleteUrlMapping(ctx context.Context, url string) error
	RedirectUrl(ctx context.Context, shortUrl string) (string, error)
}

type UrlShorteningService struct {
	UrlShortener
	cache *cache.ChainCache[string]
	store dal.UrlStore
}

func New(opts ...Option) *UrlShorteningService {

	service := &UrlShorteningService{}
	for _, opt := range opts {
		opt(service)
	}
	return service
}

func (uss *UrlShorteningService) CreateUrlMapping(ctx context.Context, longUrl string) (string, error) {

	return "", nil
}

func (uss *UrlShorteningService) DeleteUrlMapping(ctx context.Context, longUrl string) error {

	return nil
}

func (uss *UrlShorteningService) RedirectUrl(ctx context.Context, shortUrl string) (string, error) {

	return "", nil
}
