package dal

import (
	"context"
	"errors"
	"fmt"
	"log"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type UrlMappingEntry struct {
	ShortUrl     string `bson:"short_url" json:"short_url"`
	LongUrl      string `bson:"long_url" json:"long_url"`
	Hits         int64  `bson:"hits" json:"hits"`
	CreatedTs    int64  `bson:"created_ts" json:"created_ts"`
	LastAccessed int64  `bson:"last_accessed" json:"last_accessed"`
}

type UrlStore interface {
	AddUrlEntry(ctx context.Context, entry *UrlMappingEntry) error
	GetMappedUrl(ctx context.Context, shortUrl string) (string, error)
	DeleteUrlEntry(ctx context.Context, shortUrl string) error
	CheckIfUrlExists(ctx context.Context, url string, isLong bool) bool
	UpdateUrlHitCount(ctx context.Context, shortUrl string) error
	GetUrlMetrics(ctx context.Context, start, end int64, asc bool) ([]*UrlMappingEntry, error)
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

func (ms *MongoUrlStore) GetUrlMetrics(ctx context.Context, start, end int64, asc bool) ([]*UrlMappingEntry, error) {
	gatelyDB := ms.c.Database(ms.name)
	urlTbl := gatelyDB.Collection(ms.collection)
	// Specify the Sort option to sort the returned documents by hit count in
	// ascending  or descending order.

	sortOrder := -1
	if asc {
		sortOrder = 1
	}
	opts := options.Find().SetSort(bson.D{{"hits", sortOrder}})

	filter := bson.M{
		"last_accessed": bson.M{
			"$gte": start,
			"$lt":  end,
		},
	}
	log.Printf("Trying to get entries between start=%d and end=%d", start, end)
	cursor, err := urlTbl.Find(ctx, filter, opts)

	if err != nil {
		log.Printf("Unable to get metrics for the given dates. Err=%v", err)
		return nil, err
	}

	var results []*UrlMappingEntry
	// Get a list of all returned Urls and print them out in the required order.
	for cursor.Next(ctx) {
		// Create a value into which the single document can be decoded
		elem := &UrlMappingEntry{}
		err := cursor.Decode(elem)
		if err != nil {
			log.Fatal(err)
		}

		results = append(results, elem)

	}

	if err := cursor.Err(); err != nil {
		log.Fatal(err)
	}

	// Close the cursor once finished
	_ = cursor.Close(ctx)
	return results, nil
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
	gatelyDB := ms.c.Database(ms.name)
	urlTbl := gatelyDB.Collection(ms.collection)
	_, err := urlTbl.DeleteOne(ctx, bson.M{"short_url": shortUrl})

	if err != nil {
		return err
	}

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
	replacement.LastAccessed = time.Now().Unix()
	filter := bson.M{"short_url": result.ShortUrl}
	updResult, err := urlTbl.ReplaceOne(ctx, filter, &replacement)
	if err != nil {
		log.Printf("Unable to add new URL entry. Err = %v", err)
		return err
	}
	log.Printf("Successfully updated document %+v", updResult)

	return nil
}
