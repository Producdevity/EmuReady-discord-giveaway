package domain

import "strconv"

const (
	InteractionTypePing               = 1
	InteractionTypeApplicationCommand = 2

	InteractionResponsePong            = 1
	InteractionResponseChannelMessage  = 4
	InteractionResponseDeferredMessage = 5

	MessageFlagEphemeral  = 64
	PermissionManageGuild = 0x20
)

type Interaction struct {
	Type    int                `json:"type"`
	Data    *InteractionData   `json:"data"`
	GuildID string             `json:"guild_id"`
	Member  *InteractionMember `json:"member"`
	User    *DiscordUser       `json:"user"`
	Token   string             `json:"token"`
}

type InteractionData struct {
	Name    string              `json:"name"`
	Options []InteractionOption `json:"options"`
}

type InteractionOption struct {
	Name  string `json:"name"`
	Type  int    `json:"type"`
	Value any    `json:"value"`
}

type InteractionResponse struct {
	Type int                     `json:"type"`
	Data *InteractionMessageData `json:"data,omitempty"`
}

type InteractionMessageData struct {
	Content         string           `json:"content,omitempty"`
	Flags           int              `json:"flags,omitempty"`
	Components      []any            `json:"components,omitempty"`
	AllowedMentions *AllowedMentions `json:"allowed_mentions,omitempty"`
}

type WebhookMessageEdit struct {
	Content         string           `json:"content,omitempty"`
	AllowedMentions *AllowedMentions `json:"allowed_mentions,omitempty"`
}

type AllowedMentions struct {
	Parse []string `json:"parse,omitempty"`
}

type InteractionMember struct {
	User        *DiscordUser `json:"user"`
	Permissions string       `json:"permissions"`
	Roles       []string     `json:"roles"`
}

type DiscordUser struct {
	ID string `json:"id"`
}

func (d *InteractionData) IntOption(name string) (int, bool) {
	for _, option := range d.Options {
		if option.Name != name {
			continue
		}
		switch value := option.Value.(type) {
		case float64:
			return int(value), true
		case string:
			parsed, err := strconv.Atoi(value)
			if err == nil {
				return parsed, true
			}
		}
	}
	return 0, false
}
