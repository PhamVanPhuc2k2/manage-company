package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/PhamVanPhuc2k2/manage-company/pkg/config"
	"github.com/PhamVanPhuc2k2/manage-company/pkg/logger"
	"github.com/PhamVanPhuc2k2/manage-company/pkg/metrics"
	"github.com/PhamVanPhuc2k2/manage-company/pkg/postgres"
	"github.com/PhamVanPhuc2k2/manage-company/pkg/rabbitmq"

	"github.com/PhamVanPhuc2k2/manage-company/internal/delivery/http/router"
	repopg "github.com/PhamVanPhuc2k2/manage-company/internal/repository/postgres"
	repomq "github.com/PhamVanPhuc2k2/manage-company/internal/repository/rabbitmq"
	reporedis "github.com/PhamVanPhuc2k2/manage-company/internal/repository/redis"
	goredis "github.com/redis/go-redis/v9"

	ucsystem "github.com/PhamVanPhuc2k2/manage-company/internal/usecase/system"
)

var (
	version   = "dev"
	gitSHA    = "unknown"
	buildTime = "unknown"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "api dừng do lỗi: %v\n", err)
		os.Exit(1)
	}
}

func run() error {
	cfg, err := config.Load("api")
	if err != nil {
		return err
	}

	log := logger.New("api", cfg.LogLeval, cfg.Env)
	metrics.SetBuildInfo("api", version, gitSHA, buildTime)
	log.Info().Str("version", version).Str("git_sha", gitSHA).Str("build_time", buildTime).Str("env", cfg.Env).Msg("đang khởi động api")

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGTERM, syscall.SIGINT)
	defer stop()

	db, err := postgres.New(ctx, cfg.PostgresDSN, cfg.PostgresMaxConn, log)
	if err != nil {
		return nil
	}
	defer db.Close()

	rdb := goredis.NewClient(&goredis.Options{
		Addr:     cfg.RedisAddr,
		Password: cfg.RedisPassword,
		DB:       cfg.RedisDB,
	})
	defer func() { _ = rdb.Close() }()

	mqClient, err := rabbitmq.New(ctx, cfg.RabbitMQURL, log)
	if err != nil {
		return err
	}
	defer func() { _ = mqClient.Close() }()

	// ----- Repository -----
	clockRepo := repopg.NewClockRepository(db)
	cacheRepo := reporedis.NewCacheRepository(rdb)
	publisher := repomq.NewJobPublisher(mqClient)

	// ----- Usecase -----
	pingUc := ucsystem.NewPingUsecase(clockRepo, cacheRepo, publisher)

	// ----- Handler -----
	srv := &http.Server{
		Addr: fmt.Sprintf(":%d", cfg.HTTPPort),
		Handler: router.New(router.Deps{
			Config:   cfg,
			Logger:   log,
			Version:  version,
			Commit:   gitSHA,
			Postgres: db,
			Redis:    rdb,
			Rabbitmq: mqClient,
			PingUc:   pingUc,
		}),
		ReadHeaderTimeout: 10 * time.Second,
		ReadTimeout:       30 * time.Second,
		WriteTimeout:      60 * time.Second,
		IdleTimeout:       120 * time.Second,
	}

	errCh := make(chan error, 1)
	go func() {
		log.Info().Int("port", cfg.HTTPPort).Msg("HTTP server đang chạy")
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			errCh <- err
		}
	}()

	select {
	case err := <-errCh:
		return err
	case <-ctx.Done():
		log.Info().Msg("nhận tín hiệu dừng, đang tắt êm")
		shutdownCtx, cancel := context.WithTimeout(context.Background(), cfg.ShutdownTimeout)
		defer cancel()

		if err := srv.Shutdown(shutdownCtx); err != nil {
			return fmt.Errorf("tắt HTTP server: %w", err)
		}
		log.Info().Msg("api đã dừng")
		return nil
	}
}
