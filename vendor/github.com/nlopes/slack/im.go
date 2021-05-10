package slack

import (
	"context"
	"errors"
	"net/url"
	"strconv"
)

type imChannel struct {
	ID string `json:"id"`
}

type imResponseFull struct {
	NoOp          bool      `json:"no_op"`
	AlreadyClosed bool      `json:"already_closed"`
	AlreadyOpen   bool      `json:"already_open"`
	Channel       imChannel `json:"channel"`
	IMs           []IM      `json:"ims"`
	History
	SlackResponse
}

// IM contains information related to the Direct Message channel
type IM struct {
	conversation
	IsIM          bool   `json:"is_im"`
	User          string `json:"user"`
	IsUserDeleted bool   `json:"is_user_deleted"`
}

func imRequest(ctx context.Context, client HTTPRequester, path string, values url.Values, debug bool) (*imResponseFull, error) {
	response := &imResponseFull{}
	err := postSlackMethod(ctx, client, path, values, response, debug)
	if err != nil {
		return nil, err
	}
	if !response.Ok {
		return nil, errors.New(response.Error)
	}
	return response, nil
}

// CloseIMChannel closes the direct message channel
func (api *Client) CloseIMChannel(channel string) (bool, bool, error) {
	return api.CloseIMChannelContext(context.Background(), channel)
}

// CloseIMChannelContext closes the direct message channel with a custom context
func (api *Client) CloseIMChannelContext(ctx context.Context, channel string) (bool, bool, error) {
	values := url.Values{
		"token":   {api.token},
		"channel": {channel},
	}

	response, err := imRequest(ctx, api.httpclient, "im.close", values, api.debug)
	if err != nil {
		return false, false, err
	}
	return response.NoOp, response.AlreadyClosed, nil
}

// OpenIMChannel opens a direct message channel to the user provided as argument
// Returns some status and the channel ID
func (api *Client) OpenIMChannel(user string) (bool, bool, string, error) {
	return api.OpenIMChannelContext(context.Background(), user)
}

// OpenIMChannelContext opens a direct message channel to the user provided as argument with a custom context
// Returns some status and the channel ID
func (api *Client) OpenIMChannelContext(ctx context.Context, user string) (bool, bool, string, error) {
	values := url.Values{
		"token": {api.token},
		"user":  {user},
	}

	response, err := imRequest(ctx, api.httpclient, "im.open", values, api.debug)
	if err != nil {
		return false, false, "", err
	}
	return response.NoOp, response.AlreadyOpen, response.Channel.ID, nil
}

// MarkIMChannel sets the read mark of a direct message channel to a specific point
func (api *Client) MarkIMChannel(channel, ts string) (err error) {
	return api.MarkIMChannelContext(context.Background(), channel, ts)
}

// MarkIMChannelContext sets the read mark of a direct message channel to a specific point with a custom context
func (api *Client) MarkIMChannelContext(ctx context.Context, channel, ts string) error {
	values := url.Values{
		"token":   {api.token},
		"channel": {channel},
		"ts":      {ts},
	}

	_, err := imRequest(ctx, api.httpclient, "im.mark", values, api.debug)
	return err
}

// GetIMHistory retrieves the direct message channel history
func (api *Client) GetIMHistory(channel string, params HistoryParameters) (*History, error) {
	return api.GetIMHistoryContext(context.Background(), channel, params)
}

// GetIMHistoryContext retrieves the direct message channel history with a custom context
func (api *Client) GetIMHistoryContext(ctx context.Context, channel string, params HistoryParameters) (*History, error) {
	values := url.Values{
		"token":   {api.token},
		"channel": {channel},
	}
	if params.Latest != DEFAULT_HISTORY_LATEST {
		values.Add("latest", params.Latest)
	}
	if params.Oldest != DEFAULT_HISTORY_OLDEST {
		values.Add("oldest", params.Oldest)
	}
	if params.Count != DEFAULT_HISTORY_COUNT {
		values.Add("count", strconv.Itoa(params.Count))
	}
	if params.Inclusive != DEFAULT_HISTORY_INCLUSIVE {
		if params.Inclusive {
			values.Add("inclusive", "1")
		} else {
			values.Add("inclusive", "0")
		}
	}
	if params.Unreads != DEFAULT_HISTORY_UNREADS {
		if params.Unreads {
			values.Add("unreads", "1")
		} else {
			values.Add("unreads", "0")
		}
	}

	response, err := imRequest(ctx, api.httpclient, "im.history", values, api.debug)
	if err != nil {
		return nil, err
	}
	return &response.History, nil
}

// GetIMChannels returns the list of direct message channels
func (api *Client) GetIMChannels() ([]IM, error) {
	return api.GetIMChannelsContext(context.Background())
}

// GetIMChannelsContext returns the list of direct message channels with a custom context
func (api *Client) GetIMChannelsContext(ctx context.Context) ([]IM, error) {
	values := url.Values{
		"token": {api.token},
	}

	response, err := imRequest(ctx, api.httpclient, "im.list", values, api.debug)
	if err != nil {
		return nil, err
	}
	return response.IMs, nil
}
