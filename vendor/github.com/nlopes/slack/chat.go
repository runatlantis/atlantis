package slack

import (
	"context"
	"encoding/json"
	"net/url"

	"github.com/nlopes/slack/slackutilsx"
)

const (
	DEFAULT_MESSAGE_USERNAME         = ""
	DEFAULT_MESSAGE_REPLY_BROADCAST  = false
	DEFAULT_MESSAGE_ASUSER           = false
	DEFAULT_MESSAGE_PARSE            = ""
	DEFAULT_MESSAGE_THREAD_TIMESTAMP = ""
	DEFAULT_MESSAGE_LINK_NAMES       = 0
	DEFAULT_MESSAGE_UNFURL_LINKS     = false
	DEFAULT_MESSAGE_UNFURL_MEDIA     = true
	DEFAULT_MESSAGE_ICON_URL         = ""
	DEFAULT_MESSAGE_ICON_EMOJI       = ""
	DEFAULT_MESSAGE_MARKDOWN         = true
	DEFAULT_MESSAGE_ESCAPE_TEXT      = true
)

type chatResponseFull struct {
	Channel          string `json:"channel"`
	Timestamp        string `json:"ts"`         //Regualr message timestamp
	MessageTimeStamp string `json:"message_ts"` //Ephemeral message timestamp
	Text             string `json:"text"`
	SlackResponse
}

// getMessageTimestamp will inspect the `chatResponseFull` to ruturn a timestamp value
// in `chat.postMessage` its under `ts`
// in `chat.postEphemeral` its under `message_ts`
func (c chatResponseFull) getMessageTimestamp() string {
	if len(c.Timestamp) > 0 {
		return c.Timestamp
	}
	return c.MessageTimeStamp
}

// PostMessageParameters contains all the parameters necessary (including the optional ones) for a PostMessage() request
type PostMessageParameters struct {
	Username        string       `json:"username"`
	AsUser          bool         `json:"as_user"`
	Parse           string       `json:"parse"`
	ThreadTimestamp string       `json:"thread_ts"`
	ReplyBroadcast  bool         `json:"reply_broadcast"`
	LinkNames       int          `json:"link_names"`
	Attachments     []Attachment `json:"attachments"`
	UnfurlLinks     bool         `json:"unfurl_links"`
	UnfurlMedia     bool         `json:"unfurl_media"`
	IconURL         string       `json:"icon_url"`
	IconEmoji       string       `json:"icon_emoji"`
	Markdown        bool         `json:"mrkdwn,omitempty"`
	EscapeText      bool         `json:"escape_text"`

	// chat.postEphemeral support
	Channel string `json:"channel"`
	User    string `json:"user"`
}

// NewPostMessageParameters provides an instance of PostMessageParameters with all the sane default values set
func NewPostMessageParameters() PostMessageParameters {
	return PostMessageParameters{
		Username:        DEFAULT_MESSAGE_USERNAME,
		User:            DEFAULT_MESSAGE_USERNAME,
		AsUser:          DEFAULT_MESSAGE_ASUSER,
		Parse:           DEFAULT_MESSAGE_PARSE,
		ThreadTimestamp: DEFAULT_MESSAGE_THREAD_TIMESTAMP,
		LinkNames:       DEFAULT_MESSAGE_LINK_NAMES,
		Attachments:     nil,
		UnfurlLinks:     DEFAULT_MESSAGE_UNFURL_LINKS,
		UnfurlMedia:     DEFAULT_MESSAGE_UNFURL_MEDIA,
		IconURL:         DEFAULT_MESSAGE_ICON_URL,
		IconEmoji:       DEFAULT_MESSAGE_ICON_EMOJI,
		Markdown:        DEFAULT_MESSAGE_MARKDOWN,
		EscapeText:      DEFAULT_MESSAGE_ESCAPE_TEXT,
	}
}

// DeleteMessage deletes a message in a channel
func (api *Client) DeleteMessage(channel, messageTimestamp string) (string, string, error) {
	respChannel, respTimestamp, _, err := api.SendMessageContext(context.Background(), channel, MsgOptionDelete(messageTimestamp))
	return respChannel, respTimestamp, err
}

// DeleteMessageContext deletes a message in a channel with a custom context
func (api *Client) DeleteMessageContext(ctx context.Context, channel, messageTimestamp string) (string, string, error) {
	respChannel, respTimestamp, _, err := api.SendMessageContext(ctx, channel, MsgOptionDelete(messageTimestamp))
	return respChannel, respTimestamp, err
}

// PostMessage sends a message to a channel.
// Message is escaped by default according to https://api.slack.com/docs/formatting
// Use http://davestevens.github.io/slack-message-builder/ to help crafting your message.
func (api *Client) PostMessage(channel, text string, params PostMessageParameters) (string, string, error) {
	respChannel, respTimestamp, _, err := api.SendMessageContext(
		context.Background(),
		channel,
		MsgOptionText(text, params.EscapeText),
		MsgOptionAttachments(params.Attachments...),
		MsgOptionPostMessageParameters(params),
	)
	return respChannel, respTimestamp, err
}

// PostMessageContext sends a message to a channel with a custom context
// For more details, see PostMessage documentation
func (api *Client) PostMessageContext(ctx context.Context, channel, text string, params PostMessageParameters) (string, string, error) {
	respChannel, respTimestamp, _, err := api.SendMessageContext(
		ctx,
		channel,
		MsgOptionText(text, params.EscapeText),
		MsgOptionAttachments(params.Attachments...),
		MsgOptionPostMessageParameters(params),
	)
	return respChannel, respTimestamp, err
}

// PostEphemeral sends an ephemeral message to a user in a channel.
// Message is escaped by default according to https://api.slack.com/docs/formatting
// Use http://davestevens.github.io/slack-message-builder/ to help crafting your message.
func (api *Client) PostEphemeral(channelID, userID string, options ...MsgOption) (string, error) {
	return api.PostEphemeralContext(
		context.Background(),
		channelID,
		userID,
		options...,
	)
}

// PostEphemeralContext sends an ephemeal message to a user in a channel with a custom context
// For more details, see PostEphemeral documentation
func (api *Client) PostEphemeralContext(ctx context.Context, channelID, userID string, options ...MsgOption) (timestamp string, err error) {
	_, timestamp, _, err = api.SendMessageContext(ctx, channelID, append(options, MsgOptionPostEphemeral2(userID))...)
	return timestamp, err
}

// UpdateMessage updates a message in a channel
func (api *Client) UpdateMessage(channelID, timestamp, text string) (string, string, string, error) {
	return api.UpdateMessageContext(context.Background(), channelID, timestamp, text)
}

// UpdateMessageContext updates a message in a channel
func (api *Client) UpdateMessageContext(ctx context.Context, channelID, timestamp, text string) (string, string, string, error) {
	return api.SendMessageContext(ctx, channelID, MsgOptionUpdate(timestamp), MsgOptionText(text, true))
}

// SendMessage more flexible method for configuring messages.
func (api *Client) SendMessage(channel string, options ...MsgOption) (string, string, string, error) {
	return api.SendMessageContext(context.Background(), channel, options...)
}

// SendMessageContext more flexible method for configuring messages with a custom context.
func (api *Client) SendMessageContext(ctx context.Context, channelID string, options ...MsgOption) (channel string, timestamp string, text string, err error) {
	var (
		config   sendConfig
		response chatResponseFull
	)

	if config, err = applyMsgOptions(api.token, channelID, options...); err != nil {
		return "", "", "", err
	}

	if err = postForm(ctx, api.httpclient, config.endpoint, config.values, &response, api.debug); err != nil {
		return "", "", "", err
	}

	return response.Channel, response.getMessageTimestamp(), response.Text, response.Err()
}

// UnsafeApplyMsgOptions utility function for debugging/testing chat requests.
// NOTE: USE AT YOUR OWN RISK: No issues relating to the use of this function
// will be supported by the library.
func UnsafeApplyMsgOptions(token, channel string, options ...MsgOption) (string, url.Values, error) {
	config, err := applyMsgOptions(token, channel, options...)
	return config.endpoint, config.values, err
}

func applyMsgOptions(token, channel string, options ...MsgOption) (sendConfig, error) {
	config := sendConfig{
		endpoint: SLACK_API + string(chatPostMessage),
		values: url.Values{
			"token":   {token},
			"channel": {channel},
		},
	}

	for _, opt := range options {
		if err := opt(&config); err != nil {
			return config, err
		}
	}

	return config, nil
}

type sendMode string

const (
	chatUpdate        sendMode = "chat.update"
	chatPostMessage   sendMode = "chat.postMessage"
	chatDelete        sendMode = "chat.delete"
	chatPostEphemeral sendMode = "chat.postEphemeral"
	chatMeMessage     sendMode = "chat.meMessage"
)

type sendConfig struct {
	endpoint string
	values   url.Values
}

// MsgOption option provided when sending a message.
type MsgOption func(*sendConfig) error

// MsgOptionPost posts a messages, this is the default.
func MsgOptionPost() MsgOption {
	return func(config *sendConfig) error {
		config.endpoint = SLACK_API + string(chatPostMessage)
		config.values.Del("ts")
		return nil
	}
}

// MsgOptionPostEphemeral - DEPRECATED: use MsgOptionPostEphemeral2
// posts an ephemeral message.
func MsgOptionPostEphemeral() MsgOption {
	return func(config *sendConfig) error {
		config.endpoint = SLACK_API + string(chatPostEphemeral)
		config.values.Del("ts")
		return nil
	}
}

// MsgOptionPostEphemeral2 - posts an ephemeral message to the provided user.
func MsgOptionPostEphemeral2(userID string) MsgOption {
	return func(config *sendConfig) error {
		config.endpoint = SLACK_API + string(chatPostEphemeral)
		MsgOptionUser(userID)(config)
		config.values.Del("ts")

		return nil
	}
}

// MsgOptionMeMessage posts a "me message" type from the calling user
func MsgOptionMeMessage() MsgOption {
	return func(config *sendConfig) error {
		config.endpoint = SLACK_API + string(chatMeMessage)
		return nil
	}
}

// MsgOptionUpdate updates a message based on the timestamp.
func MsgOptionUpdate(timestamp string) MsgOption {
	return func(config *sendConfig) error {
		config.endpoint = SLACK_API + string(chatUpdate)
		config.values.Add("ts", timestamp)
		return nil
	}
}

// MsgOptionDelete deletes a message based on the timestamp.
func MsgOptionDelete(timestamp string) MsgOption {
	return func(config *sendConfig) error {
		config.endpoint = SLACK_API + string(chatDelete)
		config.values.Add("ts", timestamp)
		return nil
	}
}

// MsgOptionAsUser whether or not to send the message as the user.
func MsgOptionAsUser(b bool) MsgOption {
	return func(config *sendConfig) error {
		if b != DEFAULT_MESSAGE_ASUSER {
			config.values.Set("as_user", "true")
		}
		return nil
	}
}

// MsgOptionUser set the user for the message.
func MsgOptionUser(userID string) MsgOption {
	return func(config *sendConfig) error {
		config.values.Set("user", userID)
		return nil
	}
}

// MsgOptionText provide the text for the message, optionally escape the provided
// text.
func MsgOptionText(text string, escape bool) MsgOption {
	return func(config *sendConfig) error {
		if escape {
			text = slackutilsx.EscapeMessage(text)
		}
		config.values.Add("text", text)
		return nil
	}
}

// MsgOptionAttachments provide attachments for the message.
func MsgOptionAttachments(attachments ...Attachment) MsgOption {
	return func(config *sendConfig) error {
		if attachments == nil {
			return nil
		}

		attachments, err := json.Marshal(attachments)
		if err == nil {
			config.values.Set("attachments", string(attachments))
		}
		return err
	}
}

// MsgOptionEnableLinkUnfurl enables link unfurling
func MsgOptionEnableLinkUnfurl() MsgOption {
	return func(config *sendConfig) error {
		config.values.Set("unfurl_links", "true")
		return nil
	}
}

// MsgOptionDisableLinkUnfurl disables link unfurling
func MsgOptionDisableLinkUnfurl() MsgOption {
	return func(config *sendConfig) error {
		config.values.Set("unfurl_links", "false")
		return nil
	}
}

// MsgOptionDisableMediaUnfurl disables media unfurling.
func MsgOptionDisableMediaUnfurl() MsgOption {
	return func(config *sendConfig) error {
		config.values.Set("unfurl_media", "false")
		return nil
	}
}

// MsgOptionDisableMarkdown disables markdown.
func MsgOptionDisableMarkdown() MsgOption {
	return func(config *sendConfig) error {
		config.values.Set("mrkdwn", "false")
		return nil
	}
}

// MsgOptionTS sets the thread TS of the message to enable creating or replying to a thread
func MsgOptionTS(ts string) MsgOption {
	return func(config *sendConfig) error {
		config.values.Set("thread_ts", ts)
		return nil
	}
}

// MsgOptionBroadcast sets reply_broadcast to true
func MsgOptionBroadcast() MsgOption {
	return func(config *sendConfig) error {
		config.values.Set("reply_broadcast", "true")
		return nil
	}
}

// this function combines multiple options into a single option.
func MsgOptionCompose(options ...MsgOption) MsgOption {
	return func(c *sendConfig) error {
		for _, opt := range options {
			if err := opt(c); err != nil {
				return err
			}
		}
		return nil
	}
}

func MsgOptionParse(b bool) MsgOption {
	return func(c *sendConfig) error {
		var v string
		if b {
			v = "1"
		} else {
			v = "0"
		}
		c.values.Set("parse", v)
		return nil
	}
}

// UnsafeMsgOptionEndpoint deliver the message to the specified endpoint.
// NOTE: USE AT YOUR OWN RISK: No issues relating to the use of this Option
// will be supported by the library, it is subject to change without notice that
// may result in compilation errors or runtime behaviour changes.
func UnsafeMsgOptionEndpoint(endpoint string, update func(url.Values)) MsgOption {
	return func(config *sendConfig) error {
		config.endpoint = endpoint
		update(config.values)
		return nil
	}
}

// MsgOptionPostMessageParameters maintain backwards compatibility.
func MsgOptionPostMessageParameters(params PostMessageParameters) MsgOption {
	return func(config *sendConfig) error {
		if params.Username != DEFAULT_MESSAGE_USERNAME {
			config.values.Set("username", params.Username)
		}

		// chat.postEphemeral support
		if params.User != DEFAULT_MESSAGE_USERNAME {
			config.values.Set("user", params.User)
		}

		// never generates an error.
		MsgOptionAsUser(params.AsUser)(config)

		if params.Parse != DEFAULT_MESSAGE_PARSE {
			config.values.Set("parse", params.Parse)
		}
		if params.LinkNames != DEFAULT_MESSAGE_LINK_NAMES {
			config.values.Set("link_names", "1")
		}

		if params.UnfurlLinks != DEFAULT_MESSAGE_UNFURL_LINKS {
			config.values.Set("unfurl_links", "true")
		}

		// I want to send a message with explicit `as_user` `true` and `unfurl_links` `false` in request.
		// Because setting `as_user` to `true` will change the default value for `unfurl_links` to `true` on Slack API side.
		if params.AsUser != DEFAULT_MESSAGE_ASUSER && params.UnfurlLinks == DEFAULT_MESSAGE_UNFURL_LINKS {
			config.values.Set("unfurl_links", "false")
		}
		if params.UnfurlMedia != DEFAULT_MESSAGE_UNFURL_MEDIA {
			config.values.Set("unfurl_media", "false")
		}
		if params.IconURL != DEFAULT_MESSAGE_ICON_URL {
			config.values.Set("icon_url", params.IconURL)
		}
		if params.IconEmoji != DEFAULT_MESSAGE_ICON_EMOJI {
			config.values.Set("icon_emoji", params.IconEmoji)
		}
		if params.Markdown != DEFAULT_MESSAGE_MARKDOWN {
			config.values.Set("mrkdwn", "false")
		}

		if params.ThreadTimestamp != DEFAULT_MESSAGE_THREAD_TIMESTAMP {
			config.values.Set("thread_ts", params.ThreadTimestamp)
		}
		if params.ReplyBroadcast != DEFAULT_MESSAGE_REPLY_BROADCAST {
			config.values.Set("reply_broadcast", "true")
		}

		return nil
	}
}
