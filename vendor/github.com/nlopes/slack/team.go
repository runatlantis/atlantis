package slack

import (
	"context"
	"errors"
	"net/url"
	"strconv"
)

const (
	DEFAULT_LOGINS_COUNT = 100
	DEFAULT_LOGINS_PAGE  = 1
)

type TeamResponse struct {
	Team TeamInfo `json:"team"`
	SlackResponse
}

type TeamInfo struct {
	ID          string                 `json:"id"`
	Name        string                 `json:"name"`
	Domain      string                 `json:"domain"`
	EmailDomain string                 `json:"email_domain"`
	Icon        map[string]interface{} `json:"icon"`
}

type LoginResponse struct {
	Logins []Login `json:"logins"`
	Paging `json:"paging"`
	SlackResponse
}

type Login struct {
	UserID    string `json:"user_id"`
	Username  string `json:"username"`
	DateFirst int    `json:"date_first"`
	DateLast  int    `json:"date_last"`
	Count     int    `json:"count"`
	IP        string `json:"ip"`
	UserAgent string `json:"user_agent"`
	ISP       string `json:"isp"`
	Country   string `json:"country"`
	Region    string `json:"region"`
}

type BillableInfoResponse struct {
	BillableInfo map[string]BillingActive `json:"billable_info"`
	SlackResponse
}

type BillingActive struct {
	BillingActive bool `json:"billing_active"`
}

// AccessLogParameters contains all the parameters necessary (including the optional ones) for a GetAccessLogs() request
type AccessLogParameters struct {
	Count int
	Page  int
}

// NewAccessLogParameters provides an instance of AccessLogParameters with all the sane default values set
func NewAccessLogParameters() AccessLogParameters {
	return AccessLogParameters{
		Count: DEFAULT_LOGINS_COUNT,
		Page:  DEFAULT_LOGINS_PAGE,
	}
}

func teamRequest(ctx context.Context, client HTTPRequester, path string, values url.Values, debug bool) (*TeamResponse, error) {
	response := &TeamResponse{}
	err := postSlackMethod(ctx, client, path, values, response, debug)
	if err != nil {
		return nil, err
	}

	if !response.Ok {
		return nil, errors.New(response.Error)
	}

	return response, nil
}

func billableInfoRequest(ctx context.Context, client HTTPRequester, path string, values url.Values, debug bool) (map[string]BillingActive, error) {
	response := &BillableInfoResponse{}
	err := postSlackMethod(ctx, client, path, values, response, debug)
	if err != nil {
		return nil, err
	}

	if !response.Ok {
		return nil, errors.New(response.Error)
	}

	return response.BillableInfo, nil
}

func accessLogsRequest(ctx context.Context, client HTTPRequester, path string, values url.Values, debug bool) (*LoginResponse, error) {
	response := &LoginResponse{}
	err := postSlackMethod(ctx, client, path, values, response, debug)
	if err != nil {
		return nil, err
	}
	if !response.Ok {
		return nil, errors.New(response.Error)
	}
	return response, nil
}

// GetTeamInfo gets the Team Information of the user
func (api *Client) GetTeamInfo() (*TeamInfo, error) {
	return api.GetTeamInfoContext(context.Background())
}

// GetTeamInfoContext gets the Team Information of the user with a custom context
func (api *Client) GetTeamInfoContext(ctx context.Context) (*TeamInfo, error) {
	values := url.Values{
		"token": {api.token},
	}

	response, err := teamRequest(ctx, api.httpclient, "team.info", values, api.debug)
	if err != nil {
		return nil, err
	}
	return &response.Team, nil
}

// GetAccessLogs retrieves a page of logins according to the parameters given
func (api *Client) GetAccessLogs(params AccessLogParameters) ([]Login, *Paging, error) {
	return api.GetAccessLogsContext(context.Background(), params)
}

// GetAccessLogsContext retrieves a page of logins according to the parameters given with a custom context
func (api *Client) GetAccessLogsContext(ctx context.Context, params AccessLogParameters) ([]Login, *Paging, error) {
	values := url.Values{
		"token": {api.token},
	}
	if params.Count != DEFAULT_LOGINS_COUNT {
		values.Add("count", strconv.Itoa(params.Count))
	}
	if params.Page != DEFAULT_LOGINS_PAGE {
		values.Add("page", strconv.Itoa(params.Page))
	}

	response, err := accessLogsRequest(ctx, api.httpclient, "team.accessLogs", values, api.debug)
	if err != nil {
		return nil, nil, err
	}
	return response.Logins, &response.Paging, nil
}

func (api *Client) GetBillableInfo(user string) (map[string]BillingActive, error) {
	return api.GetBillableInfoContext(context.Background(), user)
}

func (api *Client) GetBillableInfoContext(ctx context.Context, user string) (map[string]BillingActive, error) {
	values := url.Values{
		"token": {api.token},
		"user":  {user},
	}

	return billableInfoRequest(ctx, api.httpclient, "team.billableInfo", values, api.debug)
}

// GetBillableInfoForTeam returns the billing_active status of all users on the team.
func (api *Client) GetBillableInfoForTeam() (map[string]BillingActive, error) {
	return api.GetBillableInfoForTeamContext(context.Background())
}

// GetBillableInfoForTeamContext returns the billing_active status of all users on the team with a custom context
func (api *Client) GetBillableInfoForTeamContext(ctx context.Context) (map[string]BillingActive, error) {
	values := url.Values{
		"token": {api.token},
	}

	return billableInfoRequest(ctx, api.httpclient, "team.billableInfo", values, api.debug)
}
