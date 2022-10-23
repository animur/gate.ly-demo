package dal

import (
	"context"
	"errors"
	"fmt"
	"log"

	"go.mongodb.org/mongo-driver/bson"
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
	CheckIfUrlExists(ctx context.Context, url string, isLong bool) bool
	UpdateUrlHitCount(ctx context.Context, shortUrl string) error
}

type MongoUrlStore struct {
	UrlStore
	c          *mongo.Client
	name       string
	collection string
}
type UrlStoreOption func(store *MongoUrlStore)

var (
	ErrUrlEntryAlreadyExists = errors.New("A URL entry already exists")
	ErrUrlEntryNotFound      = errors.New("URL does not exist")
)

func New(opts ...UrlStoreOption) UrlStore {
	store := &MongoUrlStore{}
	for _, opt := range opts {
		opt(store)
	}
	return store
}

func (ms *MongoUrlStore) AddUrlEntry(ctx context.Context, entry *UrlMappingEntry) error {

	if ms.CheckIfUrlExists(ctx, entry.LongUrl, true) {
		log.Printf("A short URL already exists for %s", entry.LongUrl)
		return fmt.Errorf("A short URL already exists for %s.", entry.LongUrl)
	}
	gatelyDB := ms.c.Database(ms.name)
	urlTbl := gatelyDB.Collection(ms.collection)
	insertResult, err := urlTbl.InsertOne(ctx, *entry)
	if err != nil {
		log.Printf("Unable to add new URL entry. Err = %v", err)
		return err
	}
	log.Printf("Successfully inserted insertId %s", insertResult)
	return nil
}

func (ms *MongoUrlStore) GetMappedUrl(ctx context.Context, shortUrl string) (string, error) {

	if !ms.CheckIfUrlExists(ctx, shortUrl, false) {
		log.Printf("No short URL exists for %s", shortUrl)
		return "", ErrUrlEntryNotFound
	}
	gatelyDB := ms.c.Database(ms.name)
	urlTbl := gatelyDB.Collection(ms.collection)
	var result UrlMappingEntry
	err := urlTbl.FindOne(ctx, bson.D{{"short_url", shortUrl}}).Decode(&result)

	if err != nil {
		return "", fmt.Errorf("Unable to fetch original url for %s", shortUrl)
	}

	log.Printf("Found an existing entry for %s. Entry=%+v", shortUrl, result)

	return result.LongUrl, nil
}

func (ms *MongoUrlStore) CheckIfUrlExists(ctx context.Context, url string, isLong bool) bool {
	gatelyDB := ms.c.Database(ms.name)
	urlTbl := gatelyDB.Collection(ms.collection)
	var result bson.M

	if isLong {
		err := urlTbl.FindOne(ctx, bson.D{{"long_url", url}}).Decode(&result)

		if err == mongo.ErrNoDocuments {
			return false
		}
		log.Printf("Found an URL mapping entry for Long Url %s. Entry=%+v", url, result)
	} else {
		err := urlTbl.FindOne(ctx, bson.D{{"short_url", url}}).Decode(&result)

		if err == mongo.ErrNoDocuments {
			return false
		}
		log.Printf("Found an URL mapping entry for Short Url %s. Entry=%+v", url, result)
	}

	return true
}

func (ms *MongoUrlStore) DeleteUrlEntry(ctx context.Context, shortUrl string) error {

	return nil
}

func (ms *MongoUrlStore) UpdateUrlHitCount(ctx context.Context, shortUrl string) error {
	if !ms.CheckIfUrlExists(ctx, shortUrl, false) {
		log.Printf("No short URL exists for %s", shortUrl)
		return ErrUrlEntryNotFound
	}

	gatelyDB := ms.c.Database(ms.name)
	urlTbl := gatelyDB.Collection(ms.collection)
	var result UrlMappingEntry
	_ = urlTbl.FindOne(ctx, bson.D{{"short_url", shortUrl}}).Decode(&result)

	replacement := result
	replacement.Hits += 1
	filter := bson.M{"short_url": result.ShortUrl}
	updResult, err := urlTbl.ReplaceOne(ctx, filter, &replacement)
	if err != nil {
		log.Printf("Unable to add new URL entry. Err = %v", err)
		return err
	}
	log.Printf("Successfully updated document %+v", updResult)

	return nil
}
