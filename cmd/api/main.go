package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/producdevity/emuready-discord-giveaway/internal/api/handlers"
	"github.com/producdevity/emuready-discord-giveaway/internal/api/middleware"
	"github.com/producdevity/emuready-discord-giveaway/internal/application"
	"github.com/producdevity/emuready-discord-giveaway/internal/config"
	"github.com/producdevity/emuready-discord-giveaway/internal/httpclient"
	"github.com/producdevity/emuready-discord-giveaway/internal/integrations/discord"
	"github.com/producdevity/emuready-discord-giveaway/internal/integrations/github"
	"github.com/producdevity/emuready-discord-giveaway/internal/observability"
	"github.com/producdevity/emuready-discord-giveaway/internal/storage"
	"github.com/producdevity/emuready-discord-giveaway/internal/storage/postgres"
	"github.com/producdevity/emuready-discord-giveaway/internal/worker"

	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rs/zerolog"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		panic(err)
	}
	logger, err := observability.New(cfg.LogLevel)
	if err != nil {
		panic(err)
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	httpClient := httpclient.NewHTTPClient(cfg.HTTPTimeout())
	discordClient := discord.NewClient(httpClient, cfg.DiscordToken, logger)
	githubClient := github.NewClient(httpClient, cfg.GithubClientID, cfg.GithubClientSecret, cfg.GithubApiToken)
	retryPolicy := httpclient.DefaultRetryPolicy()
	retryPolicy.MaxAttempts = cfg.HTTPMaxRetries
	discordClient.SetRetryPolicy(retryPolicy)
	githubClient.SetRetryPolicy(retryPolicy)

	db, err := connectDatabase(ctx, cfg, logger)
	if err != nil {
		logger.Fatal().Err(err).Msg("database connect failed")
	}
	defer db.Close()

	if err := storage.RunMigrations(ctx, db, cfg.MigrationsDir); err != nil {
		logger.Fatal().Err(err).Msg("migration failed")
	}

	store := postgres.NewEntrantRepository(db)
	stateSvc := application.NewOAuthStateService(cfg.SigningSecret, cfg.StateTTL())
	enterSvc := application.NewEnterService(cfg, stateSvc)

	if cfg.DiscordApplicationID == "" {
		appID, err := discordClient.GetApplicationID(context.Background())
		if err != nil {
			logger.Fatal().Err(err).Msg("failed to load application id")
		}
		cfg.DiscordApplicationID = appID
	}

	callbackSvc, err := application.NewCallbackService(cfg, stateSvc, store, discordClient, githubClient, logger)
	if err != nil {
		logger.Fatal().Err(err).Msg("callback service failed")
	}
	winnerSvc, err := application.NewWinnerService(cfg, store, discordClient, githubClient, logger)
	if err != nil {
		logger.Fatal().Err(err).Msg("winner service failed")
	}

	winnerQueue := worker.NewWinnerQueue(ctx, cfg.WinnerWorkerCount, func(c context.Context, t worker.WinnerTask) error {
		return winnerSvc.Run(c, t.Interaction, t.Count)
	}, logger)
	defer winnerQueue.Close()

	app := fiber.New(fiber.Config{
		ErrorHandler: func(c *fiber.Ctx, err error) error {
			logger.Error().Err(err).Str("request_id", getRequestIDFromCtx(c)).Msg("request failure")
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "internal server error"})
		},
		BodyLimit: cfg.MaxInteractionBodyBytes + 1024,
	})
	app.Use(middleware.RequestID())
	app.Use(middleware.Recover(logger))

	app.Get("/health", handlers.NewHealthHandler())
	app.Get("/ready", handlers.NewReadyHandler(db))
	app.Get("/privacy", handlers.NewPrivacyHandler())
	app.Get("/privacy-policy", handlers.NewPrivacyHandler())
	app.Get("/terms", handlers.NewTermsHandler())
	app.Get("/terms-of-service", handlers.NewTermsHandler())
	app.Get("/callback", handlers.NewCallbackHandler(cfg, callbackSvc, logger))
	app.Post("/interactions", handlers.NewInteractionHandler(cfg, enterSvc, discordClient, winnerQueue, logger).Handle)

	go func() {
		if err := app.Listen(":" + cfg.Port); err != nil && err.Error() != "closed" {
			logger.Error().Err(err).Msg("server stopped")
		}
	}()

	<-ctx.Done()
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := app.ShutdownWithContext(shutdownCtx); err != nil {
		logger.Error().Err(err).Msg("graceful shutdown failed")
	}
}

func connectDatabase(ctx context.Context, cfg *config.Config, logger zerolog.Logger) (*pgxpool.Pool, error) {
	deadline := time.Now().Add(cfg.DBConnectTimeout())
	var lastErr error
	for {
		db, err := pgxpool.New(ctx, cfg.DatabaseURL)
		if err == nil {
			pingCtx, cancel := context.WithTimeout(ctx, 2*time.Second)
			err = db.Ping(pingCtx)
			cancel()
			if err == nil {
				return db, nil
			}
			db.Close()
		}
		lastErr = err
		if time.Now().After(deadline) {
			return nil, lastErr
		}
		logger.Warn().Err(lastErr).Msg("database not ready; retrying")
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-time.After(time.Second):
		}
	}
}

func getRequestIDFromCtx(c *fiber.Ctx) string {
	id := c.Locals("request_id")
	if raw, ok := id.(string); ok && raw != "" {
		return raw
	}
	return "unknown"
}
