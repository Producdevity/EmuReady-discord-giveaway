package handlers

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"html"
	"strings"

	"github.com/producdevity/emuready-discord-giveaway/internal/application"
	"github.com/producdevity/emuready-discord-giveaway/internal/config"
	"github.com/producdevity/emuready-discord-giveaway/internal/domain"
	"github.com/producdevity/emuready-discord-giveaway/internal/integrations/discord"
	"github.com/producdevity/emuready-discord-giveaway/internal/integrations/github"
	"github.com/producdevity/emuready-discord-giveaway/internal/storage"

	"github.com/gofiber/fiber/v2"
	"github.com/rs/zerolog"
)

func NewCallbackHandlerFromDeps(
	cfg *config.Config,
	stateSvc *application.OAuthStateService,
	store storage.EntrantRepository,
	discordClient *discord.Client,
	githubClient *github.Client,
	logger zerolog.Logger,
) fiber.Handler {
	service, err := application.NewCallbackService(cfg, stateSvc, store, discordClient, githubClient, logger)
	if err != nil {
		logger.Error().Err(err).Msg("callback service bootstrap failed")
		return func(c *fiber.Ctx) error {
			return c.Status(fiber.StatusInternalServerError).Type("html").SendString(failureHTML("Server misconfigured, please contact support."))
		}
	}
	return NewCallbackHandler(cfg, service, logger)
}

func NewCallbackHandler(
	cfg *config.Config,
	service *application.CallbackService,
	logger zerolog.Logger,
) fiber.Handler {
	_ = cfg
	return func(c *fiber.Ctx) error {
		code := strings.TrimSpace(c.Query("code"))
		state := strings.TrimSpace(c.Query("state"))
		if code == "" || state == "" {
			return c.Status(fiber.StatusBadRequest).Type("html").SendString(failureHTML("Missing authorization code or state."))
		}
		user, err := service.Handle(c.Context(), code, state)
		if err != nil {
			logger.Warn().Err(err).Str("state_hash", hashForLog(state)).Msg("callback failed")
			return c.Status(fiber.StatusBadRequest).Type("html").SendString(failureHTML(lookupCallbackFailure(err)))
		}
		login := strings.TrimSpace(user.Login)
		if login == "" {
			login = "participant"
		}
		return c.Status(fiber.StatusOK).Type("html").SendString(successHTML(html.EscapeString(login)))
	}
}

func lookupCallbackFailure(err error) string {
	if err == nil {
		return "Could not complete OAuth flow."
	}
	var userMessage string
	switch {
	case strings.Contains(strings.ToLower(err.Error()), "expired"):
		userMessage = "Authorization link expired. Please try /enter again."
	case strings.Contains(strings.ToLower(err.Error()), "state"):
		userMessage = "Authorization link is invalid or has expired. Please try /enter again."
	case strings.Contains(strings.ToLower(err.Error()), "does not star"):
		userMessage = "Your GitHub account must have a star on the configured repository."
	case errors.Is(err, domain.ErrGitHubAlreadyLinked):
		userMessage = "This GitHub account is already linked to another Discord account."
	case strings.Contains(strings.ToLower(err.Error()), "exchange"):
		userMessage = "Unable to exchange authorization token with GitHub right now."
	default:
		userMessage = "Unable to complete authorization. Please try again later."
	}
	return userMessage
}

func successHTML(login string) string {
	return fmt.Sprintf("<!doctype html><html><body><h1>Success</h1><p>Thanks %s. Giveaway access granted.</p></body></html>", login)
}

func failureHTML(message string) string {
	return "<!doctype html><html><body><h1>Callback failed</h1><p>" + html.EscapeString(message) + "</p></body></html>"
}

func hashForLog(value string) string {
	sum := sha256.Sum256([]byte(value))
	return hex.EncodeToString(sum[:8])
}
