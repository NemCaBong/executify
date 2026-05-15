package config

import (
	"os"
	"strconv"
)

type CodeRunnerConfig struct {
	FileSecret string
	BoxModulus int
}

func LoadCodeRunnerConfig() CodeRunnerConfig {
	boxModulus, err := strconv.Atoi(os.Getenv("BOX_MODULUS"))
	if err != nil || boxModulus <= 0 {
		boxModulus = 65535
	}

	return CodeRunnerConfig{
		FileSecret: os.Getenv("FILE_SECRET"),
		BoxModulus: boxModulus,
	}
}
