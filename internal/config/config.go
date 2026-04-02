package config

import (
	"context"
	"os"
	"strconv"
	"time"

	"github.com/redis/go-redis/v9"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

type Config struct {
	ServerPort        string
	RunWorkerCount    int
	SubmitWorkerCount int
	MySQLConfig       MySQLConfig
	RedisConfig       RedisConfig
}

func Load() Config {
	port := os.Getenv("SERVER_PORT")
	if port == "" {
		port = "8080"
	}
	runWorkerCount := os.Getenv("RUN_WORKER_COUNT")
	runWorkerCountInt, err := strconv.Atoi(runWorkerCount)
	if err != nil || runWorkerCountInt <= 0 {
		runWorkerCountInt = 1
	}
	submitWorkerCount := os.Getenv("SUBMIT_WORKER_COUNT")
	submitWorkerCountInt, err := strconv.Atoi(submitWorkerCount)
	if err != nil || submitWorkerCountInt <= 0 {
		submitWorkerCountInt = 1
	}

	mysqlConf := LoadMySQLConfig()
	redisConf := LoadRedisConfig()

	return Config{
		ServerPort:        port,
		RunWorkerCount:    runWorkerCountInt,
		SubmitWorkerCount: submitWorkerCountInt,
		MySQLConfig:       mysqlConf,
		RedisConfig:       redisConf,
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
