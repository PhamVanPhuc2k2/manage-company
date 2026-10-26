package router

import (
	"context"
	"net/http"
	"time"

	"github.com/PhamVanPhuc2k2/manage-company/internal/delivery/http/handler"
	"github.com/PhamVanPhuc2k2/manage-company/pkg/config"
	"github.com/PhamVanPhuc2k2/manage-company/pkg/httpx"
	"github.com/PhamVanPhuc2k2/manage-company/pkg/metrics"
	"github.com/PhamVanPhuc2k2/manage-company/pkg/postgres"
	"github.com/PhamVanPhuc2k2/manage-company/pkg/rabbitmq"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/cors"
	"github.com/rs/zerolog"

	appmw "github.com/PhamVanPhuc2k2/manage-company/internal/delivery/http/middleware"
	chimw "github.com/go-chi/chi/v5/middleware"
	goredis "github.com/redis/go-redis/v9"
)

type Deps struct {
	Config  *config.Config
	Logger  zerolog.Logger
	Version string

	Commit   string
	Postgres *postgres.DB
	Redis    *goredis.Client
	Rabbitmq *rabbitmq.Client

	PingUc handler.PingUsecase
}

func New(d Deps) http.Handler {
	r := chi.NewRouter()

	r.Use(chimw.RequestID)
	r.Use(appmw.RealIP)
	r.Use(appmw.RequestLogger(d.Logger))
	r.Use(chimw.Recoverer)
	r.Use(chimw.Timeout(30 * time.Second))
	r.Use(chimw.Compress(5))
	r.Use(appmw.Metrics)
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   d.Config.CORSAllowedOrigins,
		AllowedMethods:   []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "x-Request-Id"},
		ExposedHeaders:   []string{"X-Request-Id"},
		AllowCredentials: true,
		MaxAge:           300,
	}))

	systemHandler := handler.NewSystemHandler(d.PingUc, d.Version, d.Commit)
	r.Method(http.MethodGet, "/metrics", metrics.Handler())

	r.Get("/health", func(w http.ResponseWriter, _ *http.Request) {
		httpx.OK(w, map[string]string{"status": "ok", "version": d.Version, "commit": d.Commit})
	})

	r.Get("/ready", func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), 3*time.Second)
		defer cancel()

		checks := map[string]string{}
		ready := true

		if err := d.Postgres.HealthCheck(ctx); err != nil {
			checks["postgres"] = err.Error()
			ready = false
		} else {
			checks["postgres"] = "ok"
		}
		if err := d.Redis.Ping(ctx).Err(); err != nil {
			checks["redis"] = err.Error()
			ready = false
		} else {
			checks["redis"] = "ok"
		}
		if err := d.Rabbitmq.HealthCheck(ctx); err != nil {
			checks["rabbitmq"] = err.Error()
			ready = false
		} else {
			checks["rabbitmq"] = "ok"
		}

		status := http.StatusOK
		if !ready {
			status = http.StatusServiceUnavailable
		}
		httpx.JSON(w, status, httpx.Envelope{Data: checks})
	})

	r.Route("/api/v1", func(r chi.Router) {
		r.Get("/ping", systemHandler.Ping)
	})

	return r
}
