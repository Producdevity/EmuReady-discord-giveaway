package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"unicode/utf8"

	"github.com/producdevity/emuready-discord-giveaway/internal/application"
	"github.com/producdevity/emuready-discord-giveaway/internal/config"
	"github.com/producdevity/emuready-discord-giveaway/internal/domain"
	"github.com/producdevity/emuready-discord-giveaway/internal/worker"

	"github.com/gofiber/fiber/v2"
	"github.com/rs/zerolog"
)

type interactionDiscordClient interface {
	VerifySignature(timestamp, signature string, body []byte, publicKey string) error
	GetMembersByRole(ctx context.Context, guildID string, roleID string) ([]string, error)
}

type InteractionHandler struct {
	cfg           *config.Config
	enterSvc      *application.EnterService
	discordClient interactionDiscordClient
	winnerQueue   *worker.WinnerQueue
	logger        zerolog.Logger
}

func NewInteractionHandler(
	cfg *config.Config,
	enterSvc *application.EnterService,
	discordClient interactionDiscordClient,
	winnerQueue *worker.WinnerQueue,
	logger zerolog.Logger,
) *InteractionHandler {
	return &InteractionHandler{
		cfg:           cfg,
		enterSvc:      enterSvc,
		discordClient: discordClient,
		winnerQueue:   winnerQueue,
		logger:        logger,
	}
}

func (h *InteractionHandler) Handle(c *fiber.Ctx) error {
	body := c.Body()
	if len(body) > h.cfg.MaxInteractionBodyBytes {
		return c.Status(fiber.StatusRequestEntityTooLarge).JSON(fiber.Map{
			"error": "request body too large",
		})
	}

	sig := c.Get("X-Signature-Ed25519")
	ts := c.Get("X-Signature-Timestamp")
	if sig == "" || ts == "" {
		h.logger.Warn().Str("request_id", getRequestID(c)).Msg("missing interaction signature headers")
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "invalid signature"})
	}
	if !utf8.Valid(body) {
		h.logger.Warn().Str("request_id", getRequestID(c)).Msg("invalid utf8 interaction body")
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid body"})
	}

	if err := h.discordClient.VerifySignature(ts, sig, body, h.cfg.DiscordPublicKey); err != nil {
		h.logger.Warn().Err(err).Str("request_id", getRequestID(c)).Msg("invalid signature")
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "invalid signature"})
	}

	var interaction domain.Interaction
	if err := json.Unmarshal(body, &interaction); err != nil {
		h.logger.Warn().Err(err).Str("request_id", getRequestID(c)).Msg("invalid interaction body")
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid body"})
	}

	switch interaction.Type {
	case domain.InteractionTypePing:
		return c.JSON(domain.InteractionResponse{Type: domain.InteractionResponsePong})
	case domain.InteractionTypeApplicationCommand:
		return h.handleCommand(c, interaction)
	default:
		return c.Status(fiber.StatusNotImplemented).JSON(fiber.Map{"error": "unsupported interaction type"})
	}
}

func (h *InteractionHandler) handleCommand(c *fiber.Ctx, interaction domain.Interaction) error {
	requestID := getRequestID(c)
	if interaction.Data == nil || interaction.Data.Name == "" {
		h.logger.Warn().Str("request_id", requestID).Msg("interaction missing command name")
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "missing command name"})
	}
	if interaction.GuildID == "" {
		h.logger.Warn().Str("request_id", requestID).Str("command", interaction.Data.Name).Msg("interaction outside guild")
		return c.JSON(domain.InteractionResponse{Type: domain.InteractionResponseChannelMessage, Data: &domain.InteractionMessageData{Content: "This command must be used inside a server.", Flags: domain.MessageFlagEphemeral}})
	}
	if interaction.Token == "" {
		h.logger.Warn().Str("request_id", requestID).Str("command", interaction.Data.Name).Msg("interaction missing token")
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "missing interaction token"})
	}

	discordID, err := parseUserID(interaction)
	if err != nil {
		h.logger.Warn().Err(err).Str("request_id", requestID).Msg("missing actor id")
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "missing actor id"})
	}

	switch interaction.Data.Name {
	case "enter", "enter-giveaway":
		resp, err := h.enterSvc.BuildResponse(discordID)
		if err != nil {
			h.logger.Error().Err(err).Str("request_id", requestID).Msg("enter response failed")
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to build response"})
		}
		return c.JSON(resp)
	case "entrants":
		members, err := h.discordClient.GetMembersByRole(c.Context(), h.cfg.GuildID, h.cfg.GiveawayRoleID)
		if err != nil {
			h.logger.Error().Err(err).Str("request_id", requestID).Msg("list entrants failed")
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to fetch entrants"})
		}
		return c.JSON(domain.InteractionResponse{Type: domain.InteractionResponseChannelMessage, Data: &domain.InteractionMessageData{Content: fmt.Sprintf("Current entrants: %d", len(members)), Flags: domain.MessageFlagEphemeral}})
	case "winner":
		if !hasManageGuild(interaction) {
			return c.JSON(domain.InteractionResponse{Type: domain.InteractionResponseChannelMessage, Data: &domain.InteractionMessageData{Content: "You need Manage Server permission to use /winner.", Flags: domain.MessageFlagEphemeral}})
		}
		count := clamp(h.cfg.WinnerDefaultCount, 1, h.cfg.WinnerMax)
		if rawCount, ok := interaction.Data.IntOption("count"); ok {
			count = clamp(rawCount, 1, h.cfg.WinnerMax)
		}
		if err := h.winnerQueue.Enqueue(c.Context(), worker.WinnerTask{Interaction: interaction, Count: count}); err != nil {
			h.logger.Error().Err(err).Str("request_id", requestID).Msg("winner queue unavailable")
			return c.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{"error": "winner queue unavailable"})
		}
		return c.JSON(domain.InteractionResponse{Type: domain.InteractionResponseDeferredMessage, Data: &domain.InteractionMessageData{Flags: domain.MessageFlagEphemeral}})
	default:
		return c.JSON(domain.InteractionResponse{Type: domain.InteractionResponseChannelMessage, Data: &domain.InteractionMessageData{Content: "Unknown command.", Flags: domain.MessageFlagEphemeral}})
	}
}

func parseUserID(interaction domain.Interaction) (int64, error) {
	var raw string
	if interaction.Member != nil && interaction.Member.User != nil {
		raw = interaction.Member.User.ID
	}
	if raw == "" && interaction.User != nil {
		raw = interaction.User.ID
	}
	if raw == "" {
		return 0, fmt.Errorf("missing user id")
	}
	id, err := strconv.ParseInt(raw, 10, 64)
	if err != nil {
		return 0, err
	}
	return id, nil
}

func hasManageGuild(interaction domain.Interaction) bool {
	if interaction.Member == nil || interaction.Member.Permissions == "" {
		return false
	}
	permRaw := strings.TrimSpace(interaction.Member.Permissions)
	permissions, err := strconv.ParseUint(permRaw, 10, 64)
	if err != nil {
		return false
	}
	return permissions&domain.PermissionManageGuild == domain.PermissionManageGuild
}

func getRequestID(c *fiber.Ctx) string {
	id := c.Locals("request_id")
	if raw, ok := id.(string); ok && raw != "" {
		return raw
	}
	return "unknown"
}

func clamp(value, min, max int) int {
	if value < min {
		return min
	}
	if value > max {
		return max
	}
	return value
}
