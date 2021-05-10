package slack

import (
	"context"
	"errors"
	"net/url"
)

type emojiResponseFull struct {
	Emoji map[string]string `json:"emoji"`
	SlackResponse
}

// GetEmoji retrieves all the emojis
func (api *Client) GetEmoji() (map[string]string, error) {
	return api.GetEmojiContext(context.Background())
}

// GetEmojiContext retrieves all the emojis with a custom context
func (api *Client) GetEmojiContext(ctx context.Context) (map[string]string, error) {
	values := url.Values{
		"token": {api.token},
	}
	response := &emojiResponseFull{}

	err := postSlackMethod(ctx, api.httpclient, "emoji.list", values, response, api.debug)
	if err != nil {
		return nil, err
	}
	if !response.Ok {
		return nil, errors.New(response.Error)
	}
	return response.Emoji, nil
}
