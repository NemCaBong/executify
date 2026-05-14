package main

import (
	"context"
	"errors"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/spf13/cobra"
	"go.uber.org/zap"

	api_http "github.com/NemCaBong/executify/internal/adapter/http"
	http_handler "github.com/NemCaBong/executify/internal/adapter/http/handler"
	"github.com/NemCaBong/executify/internal/adapter/http/middleware"
	"github.com/NemCaBong/executify/internal/adapter/queue/redis"
	"github.com/NemCaBong/executify/internal/adapter/repository"
	"github.com/NemCaBong/executify/internal/application/problem"
	"github.com/NemCaBong/executify/internal/application/submission"
	"github.com/NemCaBong/executify/internal/application/user"
	"github.com/NemCaBong/executify/internal/config"
	"github.com/NemCaBong/executify/internal/logger"
)

func main() {
	rootCmd := &cobra.Command{
		Use:   "executify",
		Short: "Executify is a powerful backend app for online judge system",
		Long:  `A robust Go service to serve API requests for a scalable online judge system.`,
		Run: func(cmd *cobra.Command, args []string) {
			log := logger.Init()
			defer log.Sync() //nolint:errcheck

			// Create context that listens for the interrupt signal from the OS.
			ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
			defer stop()

			cfg := config.Load()
			db := config.NewMySQLConnection(cfg)
			redisClient := config.NewRedisClient(cfg)

			submissionRepo := repository.NewSubmissionRepository(db)
			problemRepo := repository.NewProblemRepository(db)
			userRepo := repository.NewUserRepository(db)
			refreshTokenRepo := repository.NewRefreshTokenRepository(db)
			redisProducer := redis.NewRedisProducer(redisClient)

			submissionUC := submission.NewUsecase(submissionRepo, problemRepo)
			problemUC := problem.NewUsecase(problemRepo)
			userUC := user.NewUsecase(
				userRepo,
				refreshTokenRepo,
				[]byte(cfg.JWTSecret),
				cfg.AccessTokenTTL,
				cfg.RefreshTokenTTL,
			)
			submissionHandler := http_handler.NewSubmissionHandler(&cfg, submissionUC, redisProducer)
			problemHandler := http_handler.NewProblemHandler(problemUC)
			authHandler := http_handler.NewAuthHandler(userUC)
			app := api_http.NewApp(submissionHandler, problemHandler, authHandler)

			r := setupRouter(app, []byte(cfg.JWTSecret))

			srv := &http.Server{
				Addr:    ":" + cfg.ServerPort,
				Handler: r,
			}

			go func() {
				log.Info("server starting", zap.String("addr", srv.Addr))
				if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
					log.Error("server listen error", zap.Error(err))
				}
			}()

			<-ctx.Done()

			stop()
			log.Info("shutting down gracefully")

			shutdownCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			defer cancel()
			if err := srv.Shutdown(shutdownCtx); err != nil {
				log.Error("server forced to shutdown", zap.Error(err))
			}

			log.Info("server exited")
		},
	}

	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func setupRouter(app *api_http.App, jwtSecret []byte) *gin.Engine {
	r := gin.New()
	r.Use(gin.Recovery(), middleware.RequestLogger())

	v1 := r.Group("/api/v1")
	{
		auth := v1.Group("/auth")
		{
			auth.POST("/register", app.AuthHandler.Register)
			auth.POST("/login", app.AuthHandler.Login)
			auth.POST("/refresh", app.AuthHandler.Refresh)
			auth.POST("/logout", app.AuthHandler.Logout)
		}

		protected := v1.Group("", middleware.Auth(jwtSecret))
		{
			protected.POST("/submissions", app.SubmissionHandler.Submit)
			protected.POST("/submissions/run", app.SubmissionHandler.Run)
			protected.GET("/submissions/:id", app.SubmissionHandler.GetStatus)
			protected.PUT("/problems", app.ProblemHandler.Upsert)
		}
	}

	return r
}
