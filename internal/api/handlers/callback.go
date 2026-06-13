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
			failure := callbackFailure{Message: "Server misconfigured, please contact staff."}
			return c.Status(fiber.StatusInternalServerError).Type("html").SendString(failureHTML(failure, "", ""))
		}
	}
	return NewCallbackHandler(cfg, service, logger)
}

func NewCallbackHandler(
	cfg *config.Config,
	service *application.CallbackService,
	logger zerolog.Logger,
) fiber.Handler {
	repoName, repoURL := callbackRepo(cfg)
	return func(c *fiber.Ctx) error {
		code := strings.TrimSpace(c.Query("code"))
		state := strings.TrimSpace(c.Query("state"))
		if code == "" || state == "" {
			failure := callbackFailure{
				Message:  "This authorization link is incomplete.",
				NextStep: "Return to Discord and run /enter-giveaway again.",
			}
			return c.Status(fiber.StatusBadRequest).Type("html").SendString(failureHTML(failure, repoName, repoURL))
		}
		user, err := service.Handle(c.Context(), code, state)
		if err != nil {
			logger.Warn().Err(err).Str("state_hash", hashForLog(state)).Msg("callback failed")
			return c.Status(fiber.StatusBadRequest).Type("html").SendString(failureHTML(lookupCallbackFailure(err, repoName), repoName, repoURL))
		}
		login := strings.TrimSpace(user.Login)
		if login == "" {
			login = "participant"
		}
		return c.Status(fiber.StatusOK).Type("html").SendString(successHTML(html.EscapeString(login)))
	}
}

type callbackFailure struct {
	Message      string
	NextStep     string
	ShowRepoLink bool
}

func lookupCallbackFailure(err error, repoName string) callbackFailure {
	if err == nil {
		return callbackFailure{Message: "Could not complete OAuth flow."}
	}
	switch {
	case strings.Contains(strings.ToLower(err.Error()), "expired"):
		return callbackFailure{
			Message:  "Authorization link expired.",
			NextStep: "Return to Discord and run /enter-giveaway again.",
		}
	case strings.Contains(strings.ToLower(err.Error()), "state"):
		return callbackFailure{
			Message:  "Authorization link is invalid or has expired.",
			NextStep: "Return to Discord and run /enter-giveaway again.",
		}
	case strings.Contains(strings.ToLower(err.Error()), "does not star"):
		if repoName != "" {
			return callbackFailure{
				Message:      "Your GitHub account must star " + repoName + " before you can enter.",
				NextStep:     "After starring the repository, return to Discord and run /enter-giveaway again.",
				ShowRepoLink: true,
			}
		}
		return callbackFailure{
			Message:  "Your GitHub account must star the configured repository before you can enter.",
			NextStep: "After starring the repository, return to Discord and run /enter-giveaway again.",
		}
	case errors.Is(err, domain.ErrGitHubAlreadyLinked):
		return callbackFailure{
			Message:  "This GitHub account is already linked to another Discord account.",
			NextStep: "Contact staff if you think this is wrong.",
		}
	case strings.Contains(strings.ToLower(err.Error()), "missing permissions"):
		return callbackFailure{
			Message:  "Your GitHub account was verified, but Discord would not let the bot assign the giveaway role.",
			NextStep: "Please contact staff.",
		}
	case strings.Contains(strings.ToLower(err.Error()), "missing access"):
		return callbackFailure{
			Message:  "Your GitHub account was verified, but the bot cannot access this Discord server.",
			NextStep: "Please contact staff.",
		}
	case strings.Contains(strings.ToLower(err.Error()), "exchange"):
		return callbackFailure{
			Message:  "Unable to exchange authorization token with GitHub right now.",
			NextStep: "Return to Discord and run /enter-giveaway again later.",
		}
	default:
		return callbackFailure{
			Message:  "Unable to complete authorization.",
			NextStep: "Please try again later.",
		}
	}
}

// TODO: move this to a template file and make pretty
func successHTML(login string) string {
	return pageHTML("Success", fmt.Sprintf("Thanks %s. Giveaway access granted.", login), "", "", "", false)
}

// TODO: move this to a template file and make pretty
func failureHTML(failure callbackFailure, repoName string, repoURL string) string {
	return pageHTML("Callback failed", failure.Message, failure.NextStep, repoName, repoURL, failure.ShowRepoLink)
}

// TODO: move this to a template file and make pretty
func pageHTML(title string, message string, nextStep string, repoName string, repoURL string, showRepoLink bool) string {
	escapedTitle := html.EscapeString(title)
	escapedMessage := html.EscapeString(message)
	body := "<p>" + escapedMessage + "</p>"
	if showRepoLink && repoName != "" && repoURL != "" {
		body += `<p><a class="button" href="` + html.EscapeString(repoURL) + `">Open ` + html.EscapeString(repoName) + `</a></p>`
	}
	if nextStep != "" {
		body += "<p>" + html.EscapeString(nextStep) + "</p>"
	}
	if title == "Success" {
		body += "<p>You can return to Discord.</p>"
	}
	return `<!doctype html><html lang="en"><head><meta charset="utf-8"><meta name="viewport" content="width=device-width,initial-scale=1"><title>` + escapedTitle + `</title><style>body{margin:0;font-family:system-ui,-apple-system,BlinkMacSystemFont,"Segoe UI",sans-serif;background:#111217;color:#f3f4f8;display:grid;min-height:100vh;place-items:center}.panel{width:min(560px,calc(100vw - 32px));padding:32px;border:1px solid #30323d;border-radius:12px;background:#191a22}h1{margin:0 0 12px;font-size:28px;line-height:1.2}p{margin:12px 0;color:#c9cad3;line-height:1.5}.button{display:inline-block;margin-top:8px;padding:10px 14px;border-radius:8px;background:#f3f4f8;color:#111217;text-decoration:none;font-weight:700}code{padding:2px 5px;border-radius:4px;background:#282a35;color:#fff}</style></head><body><main class="panel"><h1>` + escapedTitle + `</h1>` + body + `</main></body></html>`
}

func callbackRepo(cfg *config.Config) (string, string) {
	owner, repo, err := cfg.GitHubRepoParts()
	if err != nil {
		return "", ""
	}
	repoName := owner + "/" + repo
	return repoName, "https://github.com/" + repoName
}

func hashForLog(value string) string {
	sum := sha256.Sum256([]byte(value))
	return hex.EncodeToString(sum[:8])
}
