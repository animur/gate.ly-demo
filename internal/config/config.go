package config

type AppConfig struct {
	Port                string `mapstructure:"port"`
	RedisHost           string `mapstructure:"redis-host"`
	RedisPass           string `mapstructure:"redis-pass"`
	MongoHost           string `mapstructure:"mongo-host"`
	MongoUser           string `mapstructure:"mongo-user"`
	MongoPass           string `mapstructure:"mongo-pass"`
	MongoDbName         string `mapstructure:"mongo-db-name"`
	MongoCollectionName string `mapstructure:"mongo-collection-name"`
}

func (cfg AppConfig) Check() bool {

	return true
}
