package config

import (
	"context"
	"os"
	"time"

	"github.com/redis/go-redis/v9"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

type Config struct {
	ServerPort  string
	MySQLConfig MySQLConfig
	RedisConfig RedisConfig
}

func LoadConfig() Config {
	port := os.Getenv("SERVER_PORT")
	if port == "" {
		port = "8080"
	}
	mysqlConf := LoadMySQLConfig()
	redisConf := LoadRedisConfig()

	return Config{
		ServerPort:  port,
		MySQLConfig: mysqlConf,
		RedisConfig: redisConf,
	}
}

func NewMySQLConnection(config Config) *gorm.DB {
	sqlDB, err := gorm.Open(mysql.Open(config.MySQLConfig.GetDSN()), &gorm.Config{})
	if err != nil {
		panic("failed to connect database")
	}
	db, err := sqlDB.DB()
	if err != nil {
		panic("failed to get database")
	}
	db.SetMaxIdleConns(10)
	db.SetMaxOpenConns(20)
	db.SetConnMaxIdleTime(8 * time.Minute)

	return sqlDB
}

func NewRedisClient(config Config) *redis.Client {
	rdb := redis.NewClient(&redis.Options{
		Addr:     config.RedisConfig.Address,
		Password: config.RedisConfig.Password,
		DB:       config.RedisConfig.DB,
	})

	// Check if the connection is alive
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := rdb.Ping(ctx).Err(); err != nil {
		panic("failed to connect to redis: " + err.Error())
	}

	return rdb
}
