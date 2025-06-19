package database

import (
	"context"
	"errors"

	"github.com/redis/go-redis/v9"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var (
	DB          *gorm.DB
	RedisClient *redis.Client
)

func InitDatabase(user, password, dbname, host, port, sslMode string) error {
	if DB != nil {
		return nil
	}
	// https://github.com/go-gorm/postgres
	db, err := gorm.Open(postgres.New(postgres.Config{
		DSN:                  "host=" + host + " user=" + user + " password=" + password + " dbname=" + dbname + " port=" + port + " sslmode=" + sslMode,
		PreferSimpleProtocol: true, // disables implicit prepared statement usage
	}), &gorm.Config{})
	if err != nil {
		return errors.New("failed to connect to database, " + err.Error())
	}
	// Create uuid extension if not exists
	db.Exec("CREATE EXTENSION IF NOT EXISTS \"uuid-ossp\"")
	DB = db
	return nil
}

func InitRedis(url string) {
	opts, err := redis.ParseURL(url)
	if err != nil {
		panic(err)
	}
	opts.ReadTimeout = 5 * 60 * 1000000000 // 5 minutes

	RedisClient = redis.NewClient(opts)
}

func GetRedisContext() context.Context {
	return context.Background()
}
