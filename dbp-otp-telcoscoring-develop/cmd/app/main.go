package main

import (
	"context"
	"fmt"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/jackc/tern/migrate"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/rs/zerolog/log"
	"gitlab.apps.bcc.kz/digital-banking-platform/microservices/technical/dbp-otp-telcoscoring/configs"
	"gitlab.apps.bcc.kz/digital-banking-platform/microservices/technical/dbp-otp-telcoscoring/internal/log_bcc"
	handler2 "gitlab.apps.bcc.kz/digital-banking-platform/microservices/technical/dbp-otp-telcoscoring/internal/otp/handler"
	"gitlab.apps.bcc.kz/digital-banking-platform/microservices/technical/dbp-otp-telcoscoring/internal/otp/repository"
	"gitlab.apps.bcc.kz/digital-banking-platform/microservices/technical/dbp-otp-telcoscoring/internal/otp/services"
	"net/http"
	"time"
)

var (
	httpPort = ":8080"
)

func main() {

	cfg := configs.InitEnvDBPostgre()
	log_bcc.Setup(cfg)
	pool, err := PostgresConnection(cfg)
	if err != nil {
		log.Error().Err(err)
		return
	}

	repo := repository.NewRepository(pool, cfg)
	serv := services.NewService(repo)
	hand := handler2.NewHandler(serv)

	ch := InitHandler(hand)
	startHttp(ch)
}

func startHttp(r *chi.Mux) {
	srv := &http.Server{
		Addr:    httpPort,
		Handler: r,
	}
	err := srv.ListenAndServe()
	if err != nil {
		log.Fatal().Err(err)
	}
}

func InitHandler(hand *handler2.Handler) *chi.Mux {
	router := chi.NewRouter()
	router.Use(middleware.AllowContentType("application/json"))
	router.Use(middleware.CleanPath)
	router.Use(cors.Handler(cors.Options{
		AllowedMethods:   []string{"POST", "GET", "PUT", "DELETE"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: false,
		MaxAge:           300,
	}))
	router.Get("/healthcheck", GetHealthCheck)
	router.Handle("/metrics", promhttp.Handler())
	otpRouter := chi.NewRouter()
	router.Mount("/v1/otp", otpRouter)
	otpRouter.Use(handler2.Protect)
	otpRouter.Use(middleware.Logger)
	otpRouter.Use(middleware.Recoverer)
	otpRouter.Post("/sendotp", hand.SentOTP)
	otpRouter.Post("/verifyotp", hand.Verify)
	return router
}

func GetHealthCheck(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(200)
	w.Write([]byte("Healthcheck works-V7"))
}

func PostgresConnection(env configs.Configs) (pool *pgxpool.Pool, err error) {
	postgresURL := fmt.Sprintf("postgresql://%s:%s@%s:%s/%s", env.UsernameDB, env.PasswordDB, env.HostDB, env.PortDB, env.DB)
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	pool, err = pgxpool.Connect(ctx, postgresURL)
	if err != nil {
		log.Error().Err(err)
		return
	}
	conn, err := pool.Acquire(context.Background())
	if err != nil {
		log.Error().Err(err).Msg("Unable to acquire connection")
		return nil, err
	}
	migrateDatabase(conn.Conn(), env.SchemaVersionTable, env.Migrations)
	conn.Release()
	return
}

func migrateDatabase(conn *pgx.Conn, schemeTable string, migrationPath string) {
	migrator, err := migrate.NewMigrator(context.Background(), conn, schemeTable)
	if err != nil {
		log.Error().Err(err).Msg("Unable to create migrator")
		return
	}
	err = migrator.LoadMigrations("./migrations")
	if err != nil {
		log.Error().Err(err).Msg("Unable to load migrations")
		return
	}

	err = migrator.Migrate(context.Background())
	if err != nil {
		log.Error().Err(err).Msg("Unable to migrate")
	}
	ver, err := migrator.GetCurrentVersion(context.Background())
	if err != nil {
		log.Error().Err(err).Msg("Unable to get current version")
		return
	}
	log.Info().Msgf("Current version: %d", ver)
}
