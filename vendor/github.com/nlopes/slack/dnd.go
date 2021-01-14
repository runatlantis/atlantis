package slack

import (
	"context"
	"errors"
	"net/url"
	"strconv"
	"strings"
)

type SnoozeDebug struct {
	SnoozeEndDate string `json:"snooze_end_date"`
}

type SnoozeInfo struct {
	SnoozeEnabled   bool        `json:"snooze_enabled,omitempty"`
	SnoozeEndTime   int         `json:"snooze_endtime,omitempty"`
	SnoozeRemaining int         `json:"snooze_remaining,omitempty"`
	SnoozeDebug     SnoozeDebug `json:"snooze_debug,omitempty"`
}

type DNDStatus struct {
	Enabled            bool `json:"dnd_enabled"`
	NextStartTimestamp int  `json:"next_dnd_start_ts"`
	NextEndTimestamp   int  `json:"next_dnd_end_ts"`
	SnoozeInfo
}

type dndResponseFull struct {
	DNDStatus
	SlackResponse
}

type dndTeamInfoResponse struct {
	Users map[string]DNDStatus `json:"users"`
	SlackResponse
}

func dndRequest(ctx context.Context, client HTTPRequester, path string, values url.Values, debug bool) (*dndResponseFull, error) {
	response := &dndResponseFull{}
	err := postSlackMethod(ctx, client, path, values, response, debug)
	if err != nil {
		return nil, err
	}
	if !response.Ok {
		return nil, errors.New(response.Error)
	}
	return response, nil
}

// EndDND ends the user's scheduled Do Not Disturb session
func (api *Client) EndDND() error {
	return api.EndDNDContext(context.Background())
}

// EndDNDContext ends the user's scheduled Do Not Disturb session with a custom context
func (api *Client) EndDNDContext(ctx context.Context) error {
	values := url.Values{
		"token": {api.token},
	}

	response := &SlackResponse{}

	if err := postSlackMethod(ctx, api.httpclient, "dnd.endDnd", values, response, api.debug); err != nil {
		return err
	}

	return response.Err()
}

// EndSnooze ends the current user's snooze mode
func (api *Client) EndSnooze() (*DNDStatus, error) {
	return api.EndSnoozeContext(context.Background())
}

// EndSnoozeContext ends the current user's snooze mode with a custom context
func (api *Client) EndSnoozeContext(ctx context.Context) (*DNDStatus, error) {
	values := url.Values{
		"token": {api.token},
	}

	response, err := dndRequest(ctx, api.httpclient, "dnd.endSnooze", values, api.debug)
	if err != nil {
		return nil, err
	}
	return &response.DNDStatus, nil
}

// GetDNDInfo provides information about a user's current Do Not Disturb settings.
func (api *Client) GetDNDInfo(user *string) (*DNDStatus, error) {
	return api.GetDNDInfoContext(context.Background(), user)
}

// GetDNDInfoContext provides information about a user's current Do Not Disturb settings with a custom context.
func (api *Client) GetDNDInfoContext(ctx context.Context, user *string) (*DNDStatus, error) {
	values := url.Values{
		"token": {api.token},
	}
	if user != nil {
		values.Set("user", *user)
	}

	response, err := dndRequest(ctx, api.httpclient, "dnd.info", values, api.debug)
	if err != nil {
		return nil, err
	}
	return &response.DNDStatus, nil
}

// GetDNDTeamInfo provides information about a user's current Do Not Disturb settings.
func (api *Client) GetDNDTeamInfo(users []string) (map[string]DNDStatus, error) {
	return api.GetDNDTeamInfoContext(context.Background(), users)
}

// GetDNDTeamInfoContext provides information about a user's current Do Not Disturb settings with a custom context.
func (api *Client) GetDNDTeamInfoContext(ctx context.Context, users []string) (map[string]DNDStatus, error) {
	values := url.Values{
		"token": {api.token},
		"users": {strings.Join(users, ",")},
	}
	response := &dndTeamInfoResponse{}

	if err := postSlackMethod(ctx, api.httpclient, "dnd.teamInfo", values, response, api.debug); err != nil {
		return nil, err
	}
	if !response.Ok {
		return nil, errors.New(response.Error)
	}
	return response.Users, nil
}

// SetSnooze adjusts the snooze duration for a user's Do Not Disturb
// settings. If a snooze session is not already active for the user, invoking
// this method will begin one for the specified duration.
func (api *Client) SetSnooze(minutes int) (*DNDStatus, error) {
	return api.SetSnoozeContext(context.Background(), minutes)
}

// SetSnooze adjusts the snooze duration for a user's Do Not Disturb settings with a custom context.
// For more information see the SetSnooze docs
func (api *Client) SetSnoozeContext(ctx context.Context, minutes int) (*DNDStatus, error) {
	values := url.Values{
		"token":       {api.token},
		"num_minutes": {strconv.Itoa(minutes)},
	}

	response, err := dndRequest(ctx, api.httpclient, "dnd.setSnooze", values, api.debug)
	if err != nil {
		return nil, err
	}
	return &response.DNDStatus, nil
}
