package config

import (
	"os"
	"strconv"
	"time"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

type Config struct {
	ServerPort        string
	RunWorkerCount    int
	SubmitWorkerCount int
	MySQLConfig       MySQLConfig
	RedisConfig       RedisConfig
	JWTSecret         string
	AccessTokenTTL    time.Duration
	RefreshTokenTTL   time.Duration
	CodeRunnerConfig  CodeRunnerConfig
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
	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		jwtSecret = "change-me-in-production"
	}

	accessTokenTTL := 15 * time.Minute
	if v := os.Getenv("ACCESS_TOKEN_TTL_SECONDS"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			accessTokenTTL = time.Duration(n) * time.Second
		}
	}

	refreshTokenTTL := 7 * 24 * time.Hour
	if v := os.Getenv("REFRESH_TOKEN_TTL_SECONDS"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			refreshTokenTTL = time.Duration(n) * time.Second
		}
	}

	return Config{
		ServerPort:        port,
		RunWorkerCount:    runWorkerCountInt,
		SubmitWorkerCount: submitWorkerCountInt,
		MySQLConfig:       LoadMySQLConfig(),
		RedisConfig:       LoadRedisConfig(),
		JWTSecret:         jwtSecret,
		AccessTokenTTL:    accessTokenTTL,
		RefreshTokenTTL:   refreshTokenTTL,
		CodeRunnerConfig:  LoadCodeRunnerConfig(),
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
