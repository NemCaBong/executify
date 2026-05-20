package main

import (
	"os"
	"time"

	"github.com/hibiken/asynq"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"go.uber.org/zap"

	"github.com/NemCaBong/executify/internal/adapter/queue"
	"github.com/NemCaBong/executify/internal/adapter/repository"
	"github.com/NemCaBong/executify/internal/adapter/worker"
	"github.com/NemCaBong/executify/internal/application/submission"
	"github.com/NemCaBong/executify/internal/config"
	"github.com/NemCaBong/executify/internal/logger"
)

const workerShutdownTimeout = 30 * time.Second

func init() {
	viper.AutomaticEnv()
}

func main() {
	cmd := &cobra.Command{Use: "worker"}

	cmd.AddCommand(&cobra.Command{
		Use: "run",
		Run: HandleRunSubmissions,
	})

	cmd.AddCommand(&cobra.Command{
		Use: "submit",
		Run: HandleSubmitSubmissions,
	})

	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func HandleRunSubmissions(_ *cobra.Command, _ []string) {
	log := logger.Init()
	defer log.Sync() //nolint:errcheck

	cfg := config.Load()
	db := config.NewMySQLConnection(cfg)
	submissionRepo := repository.NewSubmissionRepository(db)
	problemRepo := repository.NewProblemRepository(db)
	submissionUC := submission.NewUsecase(submissionRepo, problemRepo)

	handler := worker.NewRunHandler(&cfg, submissionUC)
	errorHandler := worker.NewErrorHandler(submissionUC)
	srv := newAsynqServer(cfg, cfg.RunWorkerCount, queue.QueueRun, errorHandler)

	mux := asynq.NewServeMux()
	mux.HandleFunc(queue.TypeSubmissionRun, handler.Handle)

	log.Info("starting run workers",
		zap.Int("concurrency", cfg.RunWorkerCount),
		zap.String("queue", queue.QueueRun),
	)

	if err := srv.Run(mux); err != nil {
		log.Fatal("run worker exited with error", zap.Error(err))
	}
}

func HandleSubmitSubmissions(_ *cobra.Command, _ []string) {
	log := logger.Init()
	defer log.Sync() //nolint:errcheck

	cfg := config.Load()
	db := config.NewMySQLConnection(cfg)
	submissionRepo := repository.NewSubmissionRepository(db)
	problemRepo := repository.NewProblemRepository(db)
	submissionUC := submission.NewUsecase(submissionRepo, problemRepo)

	handler := worker.NewSubmitHandler(&cfg, submissionUC)
	errorHandler := worker.NewErrorHandler(submissionUC)
	srv := newAsynqServer(cfg, cfg.SubmitWorkerCount, queue.QueueSubmit, errorHandler)

	mux := asynq.NewServeMux()
	mux.HandleFunc(queue.TypeSubmissionSubmit, handler.Handle)

	log.Info("starting submit workers",
		zap.Int("concurrency", cfg.SubmitWorkerCount),
		zap.String("queue", queue.QueueSubmit),
	)

	if err := srv.Run(mux); err != nil {
		log.Fatal("submit worker exited with error", zap.Error(err))
	}
}

func newAsynqServer(cfg config.Config, concurrency int, queueName string, errorHandler asynq.ErrorHandler) *asynq.Server {
	return asynq.NewServer(
		cfg.RedisConfig.AsynqRedisOpt(),
		asynq.Config{
			Concurrency:     concurrency,
			Queues:          map[string]int{queueName: 1},
			ShutdownTimeout: workerShutdownTimeout,
			ErrorHandler:    errorHandler,
		},
	)
}
