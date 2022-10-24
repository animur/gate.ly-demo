package dal

import "go.mongodb.org/mongo-driver/mongo"

func WithMongoClient(c *mongo.Client) UrlStoreOption {
	return func(store *MongoUrlStore) {
		store.c = c
	}
}

func WithDatabase(db string) UrlStoreOption {
	return func(store *MongoUrlStore) {
		store.name = db
	}
}

func WithTable(tbl string) UrlStoreOption {
	return func(store *MongoUrlStore) {
		store.name = tbl
	}
}
