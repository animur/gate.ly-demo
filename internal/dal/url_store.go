package dal

import (
	"context"
	"log"

	"go.mongodb.org/mongo-driver/mongo"
)

type UrlMappingEntry struct {
	ShortUrl  string `bson:"short_url" json:"short_url"`
	LongUrl   string `bson:"long_url" json:"long_url"`
	Hits      int64  `bson:"hits" json:"hits"`
	CreatedTs int64  `bson:"created_ts" json:"created_ts"`
}

type UrlStore interface {
	AddUrlEntry(ctx context.Context, entry *UrlMappingEntry) error
	GetMappedUrl(ctx context.Context, shortUrl string) (string, error)
	DeleteUrlEntry(ctx context.Context, shortUrl string) error
	CheckIfUrlExists(ctx context.Context, longUrl string) bool
}

type MongoUrlStore struct {
	UrlStore
	c          *mongo.Client
	name       string
	collection string
}
type UrlStoreOption func(store *MongoUrlStore)

func WithMongoDB(c *mongo.Client) UrlStoreOption {
	return func(store *MongoUrlStore) {
		store.c = c
	}
}

func WithDatabase(name string) UrlStoreOption {
	return func(store *MongoUrlStore) {
		store.name = name
	}
}

func WithTable(cname string) UrlStoreOption {
	return func(store *MongoUrlStore) {
		store.collection = cname
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

	gatelyDB := ms.c.Database("gately_store")
	urlTbl := gatelyDB.Collection("url_mappings")

	insertResult, err := urlTbl.InsertOne(ctx, *entry)
	if err != nil {
		panic(err)
	}
	log.Printf("Successfully inserted insertId %s", insertResult)
	return nil
}

func (ms *MongoUrlStore) GetMappedUrl(ctx context.Context, shortUrl string) (string, error) {

	return "", nil
}

func (ms *MongoUrlStore) CheckIfUrlExists(ctx context.Context, longUrl string) bool {

	return false
}

func (ms *MongoUrlStore) DeleteUrlEntry(ctx context.Context, shortUrl string) error {

	return nil
}
