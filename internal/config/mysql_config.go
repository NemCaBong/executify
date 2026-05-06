package config

import (
	"fmt"
	"os"
)

type MySQLConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	Database string
	Options  string
}

func LoadMySQLConfig() MySQLConfig {
	return MySQLConfig{
		Host:     os.Getenv("MYSQL_HOST"),
		Port:     os.Getenv("MYSQL_PORT"),
		User:     os.Getenv("MYSQL_USER"),
		Password: os.Getenv("MYSQL_PASSWORD"),
		Database: os.Getenv("MYSQL_DATABASE"),
		Options:  os.Getenv("MYSQL_OPTIONS"),
	}
}

func (c *MySQLConfig) GetDSN() string {
	opts := c.Options
	if opts == "" {
		opts = "?parseTime=true"
	}
	return fmt.Sprintf("%s:%s@tcp(%s:%s)/%s%s", c.User, c.Password, c.Host, c.Port, c.Database, opts)
}
