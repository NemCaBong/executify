package config

import (
	"os"
	"strconv"

	"github.com/hibiken/asynq"
)

type RedisConfig struct {
	Address  string
	Password string
	DB       int
}

func LoadRedisConfig() RedisConfig {
	dbStr := os.Getenv("REDIS_DB")
	db, err := strconv.Atoi(dbStr)
	if err != nil {
		db = 0
	}

	address := os.Getenv("REDIS_ADDRESS")
	if address == "" {
		address = "localhost:6379"
	}

	return RedisConfig{
		Address:  address,
		Password: os.Getenv("REDIS_PASSWORD"),
		DB:       db,
	}
}

func (c RedisConfig) AsynqRedisOpt() asynq.RedisClientOpt {
	return asynq.RedisClientOpt{
		Addr:     c.Address,
		Password: c.Password,
		DB:       c.DB,
	}
}
