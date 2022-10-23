package service

import (
	"context"
	"fmt"
	"log"
	"net/url"
	"strings"
	"time"

	"gately/internal/dal"
	"github.com/eko/gocache/v3/cache"
	"github.com/google/uuid"
)

const (
	appPrefix = "gate.ly"
)

type UrlShortener interface {
	CreateUrlMapping(ctx context.Context, url string) (string, error)
	DeleteUrlMapping(ctx context.Context, url string) error
	RedirectUrl(ctx context.Context, shortUrl string) (string, error)
	CheckAndSanitizeUrl(longUrl string) (string, bool)
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

func (uss *UrlShorteningService) CheckAndSanitizeUrl(longUrl string) (string, bool) {

	_, err := url.Parse(longUrl)
	if err != nil {
		log.Printf("CheckUrl returning false. Err=%w", err)
		return "", false
	}

	// Remove "www" to maintain uniformity
	// This will avoid duplicate entries for www.apple.com and apple.com
	longUrl = strings.Replace(longUrl, "www.", "", 1)

	// Remove the trailing slash. Again, to eliminate duplicates
	longUrl = strings.TrimSuffix(longUrl, "/")
	if strings.HasPrefix(longUrl, "https") {
		return longUrl, true
	}
	if strings.HasPrefix(longUrl, "http") {
		return longUrl, true
	}

	// default to https if no protocol is specified in the incoming URL

	return "https://" + longUrl, true
}

func (uss *UrlShorteningService) CreateUrlMapping(ctx context.Context, longUrl string) (string, error) {

	shortUrl := uuid.New()

	err := uss.store.AddUrlEntry(ctx, &dal.UrlMappingEntry{
		LongUrl:   longUrl,
		ShortUrl:  shortUrl.String(),
		Hits:      1,
		CreatedTs: time.Now().Unix(),
	})

	if err != nil {
		switch err {
		case dal.ErrUrlEntryAlreadyExists:
			log.Printf("A URL already exists for %s", longUrl)
			return "", fmt.Errorf("A URL already exists. Err=%w", err)
		default:
			log.Printf("Unable to add URL mapping into the UrlStore")
			return "", err
		}
	}

	// Return the newly created short url
	return fmt.Sprintf("%s/%s", appPrefix, shortUrl.String()), nil
}

func (uss *UrlShorteningService) DeleteUrlMapping(ctx context.Context, longUrl string) error {

	return nil
}

func (uss *UrlShorteningService) RedirectUrl(ctx context.Context, shortUrl string) (string, error) {

	defer func(ctx context.Context, c dal.UrlStore, shortUrl string) {
		err := c.UpdateUrlHitCount(ctx, shortUrl)

		if err != nil {
			log.Printf("Unable to update hit count for %s", shortUrl)
		}
	}(ctx, uss.store, shortUrl)
	cached, err := uss.cache.Get(ctx, shortUrl)

	if err == nil {
		log.Printf("Cached URL entry found for %s. Cached=%s", shortUrl, cached)

		return cached, nil
	}

	mapped, err := uss.store.GetMappedUrl(ctx, shortUrl)
	log.Printf("Short URL %s --> Long URL %s", shortUrl, mapped)

	_ = uss.cache.Set(ctx, shortUrl, mapped)

	if err != nil || mapped == "" {
		log.Printf("Unable to get Long URL %v", err)
		return "", err
	}
	return mapped, nil
}
