package main

import (
	"log"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/NemCaBong/executify/internal/adapter/repository"
	"github.com/NemCaBong/executify/internal/adapter/worker"
	"github.com/NemCaBong/executify/internal/config"
	"github.com/NemCaBong/executify/internal/core/service"
)

func init() {
	viper.SetDefault("WORKER_COUNT", 1)
	viper.SetDefault("WORKER_TYPE", 1) // 1 for submit worker, 2 for run worker
	viper.AutomaticEnv()               // read from environment variables
}

func main() {
	cmd := &cobra.Command{
		Use: "worker",
	}

	cmd.AddCommand(&cobra.Command{
		Use: "run",
		Run: HandleRunSubmissions,
	})

	cmd.AddCommand(&cobra.Command{
		Use: "submit",
		Run: HandleSubmitSubmissions,
	})

	if err := cmd.Execute(); err != nil {
		log.Fatal(err)
		os.Exit(1)
	}
}

func HandleRunSubmissions(_ *cobra.Command, _ []string) {
	cfg := config.Load()
	db := config.NewMySQLConnection(cfg)
	cache := config.NewRedisClient(cfg)
	submissionRepo := repository.NewSubmissionRepository(db)
	submissionSvc := service.NewSubmissionService(submissionRepo)
	runWorker := worker.NewRunWorker(cache, submissionSvc)

	if err := runWorker.Execute(); err != nil {
		log.Fatal(err)
	}
}

func HandleSubmitSubmissions(_ *cobra.Command, _ []string) {
	cfg := config.Load()
	db := config.NewMySQLConnection(cfg)
	cache := config.NewRedisClient(cfg)
	submissionRepo := repository.NewSubmissionRepository(db)
	submissionSvc := service.NewSubmissionService(submissionRepo)
	submitWorker := worker.NewSubmitWorker(cache, submissionSvc)

	if err := submitWorker.Execute(); err != nil {
		log.Fatal(err)
	}
}
