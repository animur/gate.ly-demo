package service

import (
	"context"
	"fmt"
	"log"
	"time"

	"gately/internal/dal"
	"github.com/eko/gocache/v3/cache"
	"github.com/google/uuid"
)

const (
	appPrefix = "gately.com"
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

	shortUrl := uuid.New()

	isPresent := uss.store.CheckIfUrlExists(ctx, longUrl)

	if isPresent {
		log.Printf("A short URL already exists for %s", longUrl)
		return "", fmt.Errorf("A short URL already exists for %s", longUrl)
	}
	err := uss.store.AddUrlEntry(ctx, &dal.UrlMappingEntry{
		LongUrl:   longUrl,
		ShortUrl:  shortUrl.String(),
		Hits:      1,
		CreatedTs: time.Now().Unix(),
	})

	if err != nil {
		log.Printf("Unable to add URL mapping into the UrlStore")
		return "", err
	}
	// Return the newly created short url
	return fmt.Sprintf("%s/%s", appPrefix, shortUrl.String()), nil
}

func (uss *UrlShorteningService) DeleteUrlMapping(ctx context.Context, longUrl string) error {

	return nil
}

func (uss *UrlShorteningService) RedirectUrl(ctx context.Context, shortUrl string) (string, error) {

	mapped, err := uss.store.GetMappedUrl(ctx, shortUrl)
	if err != nil || mapped == "" {
		log.Printf("Unable to get mapped URL for %s", shortUrl)
		return "", nil
	}
	return mapped, nil
}
