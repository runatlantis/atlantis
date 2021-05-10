package slack

import (
	"context"
	"errors"
	"fmt"
	"net/url"
)

type adminResponse struct {
	OK    bool   `json:"ok"`
	Error string `json:"error"`
}

func adminRequest(ctx context.Context, client HTTPRequester, method string, teamName string, values url.Values, debug bool) (*adminResponse, error) {
	adminResponse := &adminResponse{}
	err := parseAdminResponse(ctx, client, method, teamName, values, adminResponse, debug)
	if err != nil {
		return nil, err
	}

	if !adminResponse.OK {
		return nil, errors.New(adminResponse.Error)
	}

	return adminResponse, nil
}

// DisableUser disabled a user account, given a user ID
func (api *Client) DisableUser(teamName string, uid string) error {
	return api.DisableUserContext(context.Background(), teamName, uid)
}

// DisableUserContext disabled a user account, given a user ID with a custom context
func (api *Client) DisableUserContext(ctx context.Context, teamName string, uid string) error {
	values := url.Values{
		"user":       {uid},
		"token":      {api.token},
		"set_active": {"true"},
		"_attempts":  {"1"},
	}

	_, err := adminRequest(ctx, api.httpclient, "setInactive", teamName, values, api.debug)
	if err != nil {
		return fmt.Errorf("Failed to disable user with id '%s': %s", uid, err)
	}

	return nil
}

// InviteGuest invites a user to Slack as a single-channel guest
func (api *Client) InviteGuest(teamName, channel, firstName, lastName, emailAddress string) error {
	return api.InviteGuestContext(context.Background(), teamName, channel, firstName, lastName, emailAddress)
}

// InviteGuestContext invites a user to Slack as a single-channel guest with a custom context
func (api *Client) InviteGuestContext(ctx context.Context, teamName, channel, firstName, lastName, emailAddress string) error {
	values := url.Values{
		"email":            {emailAddress},
		"channels":         {channel},
		"first_name":       {firstName},
		"last_name":        {lastName},
		"ultra_restricted": {"1"},
		"token":            {api.token},
		"resend":           {"true"},
		"set_active":       {"true"},
		"_attempts":        {"1"},
	}

	_, err := adminRequest(ctx, api.httpclient, "invite", teamName, values, api.debug)
	if err != nil {
		return fmt.Errorf("Failed to invite single-channel guest: %s", err)
	}

	return nil
}

// InviteRestricted invites a user to Slack as a restricted account
func (api *Client) InviteRestricted(teamName, channel, firstName, lastName, emailAddress string) error {
	return api.InviteRestrictedContext(context.Background(), teamName, channel, firstName, lastName, emailAddress)
}

// InviteRestrictedContext invites a user to Slack as a restricted account with a custom context
func (api *Client) InviteRestrictedContext(ctx context.Context, teamName, channel, firstName, lastName, emailAddress string) error {
	values := url.Values{
		"email":      {emailAddress},
		"channels":   {channel},
		"first_name": {firstName},
		"last_name":  {lastName},
		"restricted": {"1"},
		"token":      {api.token},
		"resend":     {"true"},
		"set_active": {"true"},
		"_attempts":  {"1"},
	}

	_, err := adminRequest(ctx, api.httpclient, "invite", teamName, values, api.debug)
	if err != nil {
		return fmt.Errorf("Failed to restricted account: %s", err)
	}

	return nil
}

// InviteToTeam invites a user to a Slack team
func (api *Client) InviteToTeam(teamName, firstName, lastName, emailAddress string) error {
	return api.InviteToTeamContext(context.Background(), teamName, firstName, lastName, emailAddress)
}

// InviteToTeamContext invites a user to a Slack team with a custom context
func (api *Client) InviteToTeamContext(ctx context.Context, teamName, firstName, lastName, emailAddress string) error {
	values := url.Values{
		"email":      {emailAddress},
		"first_name": {firstName},
		"last_name":  {lastName},
		"token":      {api.token},
		"set_active": {"true"},
		"_attempts":  {"1"},
	}

	_, err := adminRequest(ctx, api.httpclient, "invite", teamName, values, api.debug)
	if err != nil {
		return fmt.Errorf("Failed to invite to team: %s", err)
	}

	return nil
}

// SetRegular enables the specified user
func (api *Client) SetRegular(teamName, user string) error {
	return api.SetRegularContext(context.Background(), teamName, user)
}

// SetRegularContext enables the specified user with a custom context
func (api *Client) SetRegularContext(ctx context.Context, teamName, user string) error {
	values := url.Values{
		"user":       {user},
		"token":      {api.token},
		"set_active": {"true"},
		"_attempts":  {"1"},
	}

	_, err := adminRequest(ctx, api.httpclient, "setRegular", teamName, values, api.debug)
	if err != nil {
		return fmt.Errorf("Failed to change the user (%s) to a regular user: %s", user, err)
	}

	return nil
}

// SendSSOBindingEmail sends an SSO binding email to the specified user
func (api *Client) SendSSOBindingEmail(teamName, user string) error {
	return api.SendSSOBindingEmailContext(context.Background(), teamName, user)
}

// SendSSOBindingEmailContext sends an SSO binding email to the specified user with a custom context
func (api *Client) SendSSOBindingEmailContext(ctx context.Context, teamName, user string) error {
	values := url.Values{
		"user":       {user},
		"token":      {api.token},
		"set_active": {"true"},
		"_attempts":  {"1"},
	}

	_, err := adminRequest(ctx, api.httpclient, "sendSSOBind", teamName, values, api.debug)
	if err != nil {
		return fmt.Errorf("Failed to send SSO binding email for user (%s): %s", user, err)
	}

	return nil
}

// SetUltraRestricted converts a user into a single-channel guest
func (api *Client) SetUltraRestricted(teamName, uid, channel string) error {
	return api.SetUltraRestrictedContext(context.Background(), teamName, uid, channel)
}

// SetUltraRestrictedContext converts a user into a single-channel guest with a custom context
func (api *Client) SetUltraRestrictedContext(ctx context.Context, teamName, uid, channel string) error {
	values := url.Values{
		"user":       {uid},
		"channel":    {channel},
		"token":      {api.token},
		"set_active": {"true"},
		"_attempts":  {"1"},
	}

	_, err := adminRequest(ctx, api.httpclient, "setUltraRestricted", teamName, values, api.debug)
	if err != nil {
		return fmt.Errorf("Failed to ultra-restrict account: %s", err)
	}

	return nil
}

// SetRestricted converts a user into a restricted account
func (api *Client) SetRestricted(teamName, uid string) error {
	return api.SetRestrictedContext(context.Background(), teamName, uid)
}

// SetRestrictedContext converts a user into a restricted account with a custom context
func (api *Client) SetRestrictedContext(ctx context.Context, teamName, uid string) error {
	values := url.Values{
		"user":       {uid},
		"token":      {api.token},
		"set_active": {"true"},
		"_attempts":  {"1"},
	}

	_, err := adminRequest(ctx, api.httpclient, "setRestricted", teamName, values, api.debug)
	if err != nil {
		return fmt.Errorf("Failed to restrict account: %s", err)
	}

	return nil
}
