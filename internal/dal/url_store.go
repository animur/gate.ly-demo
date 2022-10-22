package dal

import (
	"context"

	"go.mongodb.org/mongo-driver/mongo"
)

type UrlMappingEntry struct {
	LongUrl    string
	ShortUrl   string
	Hits       int64
	CreatedTs  int64
	ModifiedTs int64
}

type UrlStore interface {
	AddUrlEntry(ctx context.Context, entry *UrlMappingEntry) error
	GetMappedUrl(ctx context.Context, shortUrl string) error
	DeleteUrlEntry(ctx context.Context, ShortUrl string) error
}

type MongoUrlStore struct {
	UrlStore
	c *mongo.Client
}
type UrlStoreOption func(store *MongoUrlStore)

func WithMongoDB(c *mongo.Client) UrlStoreOption {
	return func(store *MongoUrlStore) {
		store.c = c
	}
}

func New(opts ...UrlStoreOption) UrlStore {
	store := &MongoUrlStore{}
	for _, opt := range opts {
		opt(store)
	}
	return store
}

func (ms *MongoUrlStore) AddUrlEntry(ctx context.Context, entry *UrlMappingEntry) error {

	return nil
}

func (ms *MongoUrlStore) GetMappedUrl(ctx context.Context, shortUrl string) error {

	return nil
}

func (ms *MongoUrlStore) DeleteUrlEntry(ctx context.Context, shortUrl string) error {

	return nil
}
