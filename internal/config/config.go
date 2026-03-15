package config

import (
	"os"
	"time"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

type Config struct {
	ServerPort  string
	MySQLConfig MySQLConfig
}

func LoadConfig() Config {
	port := os.Getenv("SERVER_PORT")
	if port == "" {
		port = "8080"
	}
	mysqlConf := LoadMySQLConfig()

	return Config{
		ServerPort:  port,
		MySQLConfig: mysqlConf,
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
