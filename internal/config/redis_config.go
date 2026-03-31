package config

import (
	"os"
	"strconv"
)

type RedisConfig struct {
	Address         string
	Password        string
	DB              int
	SubmitQueueName string
	RunQueueName    string
}

func LoadRedisConfig() RedisConfig {
	dbStr := os.Getenv("REDIS_DB")
	db, err := strconv.Atoi(dbStr)
	if err != nil {
		db = 0 // default value
	}

	address := os.Getenv("REDIS_ADDRESS")
	if address == "" {
		address = "localhost:6379"
	}

	submitQueue := os.Getenv("REDIS_SUBMIT_QUEUE")
	if submitQueue == "" {
		// Default queue name for submitting tasks
		submitQueue = "executify:queue:submit"
	}

	runQueue := os.Getenv("REDIS_RUN_QUEUE")
	if runQueue == "" {
		// Default queue name for running tasks
		runQueue = "executify:queue:run"
	}

	return RedisConfig{
		Address:         address,
		Password:        os.Getenv("REDIS_PASSWORD"),
		DB:              db,
		SubmitQueueName: submitQueue,
		RunQueueName:    runQueue,
	}
}
