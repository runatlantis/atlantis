package slack

import (
	"context"
	"errors"
	"io"
	"net/url"
	"strconv"
	"strings"
)

const (
	// Add here the defaults in the siten
	DEFAULT_FILES_USER    = ""
	DEFAULT_FILES_CHANNEL = ""
	DEFAULT_FILES_TS_FROM = 0
	DEFAULT_FILES_TS_TO   = -1
	DEFAULT_FILES_TYPES   = "all"
	DEFAULT_FILES_COUNT   = 100
	DEFAULT_FILES_PAGE    = 1
)

// File contains all the information for a file
type File struct {
	ID        string   `json:"id"`
	Created   JSONTime `json:"created"`
	Timestamp JSONTime `json:"timestamp"`

	Name              string `json:"name"`
	Title             string `json:"title"`
	Mimetype          string `json:"mimetype"`
	ImageExifRotation int    `json:"image_exif_rotation"`
	Filetype          string `json:"filetype"`
	PrettyType        string `json:"pretty_type"`
	User              string `json:"user"`

	Mode         string `json:"mode"`
	Editable     bool   `json:"editable"`
	IsExternal   bool   `json:"is_external"`
	ExternalType string `json:"external_type"`

	Size int `json:"size"`

	URL                string `json:"url"`          // Deprecated - never set
	URLDownload        string `json:"url_download"` // Deprecated - never set
	URLPrivate         string `json:"url_private"`
	URLPrivateDownload string `json:"url_private_download"`

	OriginalH   int    `json:"original_h"`
	OriginalW   int    `json:"original_w"`
	Thumb64     string `json:"thumb_64"`
	Thumb80     string `json:"thumb_80"`
	Thumb160    string `json:"thumb_160"`
	Thumb360    string `json:"thumb_360"`
	Thumb360Gif string `json:"thumb_360_gif"`
	Thumb360W   int    `json:"thumb_360_w"`
	Thumb360H   int    `json:"thumb_360_h"`
	Thumb480    string `json:"thumb_480"`
	Thumb480W   int    `json:"thumb_480_w"`
	Thumb480H   int    `json:"thumb_480_h"`
	Thumb720    string `json:"thumb_720"`
	Thumb720W   int    `json:"thumb_720_w"`
	Thumb720H   int    `json:"thumb_720_h"`
	Thumb960    string `json:"thumb_960"`
	Thumb960W   int    `json:"thumb_960_w"`
	Thumb960H   int    `json:"thumb_960_h"`
	Thumb1024   string `json:"thumb_1024"`
	Thumb1024W  int    `json:"thumb_1024_w"`
	Thumb1024H  int    `json:"thumb_1024_h"`

	Permalink       string `json:"permalink"`
	PermalinkPublic string `json:"permalink_public"`

	EditLink         string `json:"edit_link"`
	Preview          string `json:"preview"`
	PreviewHighlight string `json:"preview_highlight"`
	Lines            int    `json:"lines"`
	LinesMore        int    `json:"lines_more"`

	IsPublic        bool     `json:"is_public"`
	PublicURLShared bool     `json:"public_url_shared"`
	Channels        []string `json:"channels"`
	Groups          []string `json:"groups"`
	IMs             []string `json:"ims"`
	InitialComment  Comment  `json:"initial_comment"`
	CommentsCount   int      `json:"comments_count"`
	NumStars        int      `json:"num_stars"`
	IsStarred       bool     `json:"is_starred"`
}

// FileUploadParameters contains all the parameters necessary (including the optional ones) for an UploadFile() request.
//
// There are three ways to upload a file. You can either set Content if file is small, set Reader if file is large,
// or provide a local file path in File to upload it from your filesystem.
type FileUploadParameters struct {
	File            string
	Content         string
	Reader          io.Reader
	Filetype        string
	Filename        string
	Title           string
	InitialComment  string
	Channels        []string
	ThreadTimestamp string
}

// GetFilesParameters contains all the parameters necessary (including the optional ones) for a GetFiles() request
type GetFilesParameters struct {
	User          string
	Channel       string
	TimestampFrom JSONTime
	TimestampTo   JSONTime
	Types         string
	Count         int
	Page          int
}

type fileResponseFull struct {
	File     `json:"file"`
	Paging   `json:"paging"`
	Comments []Comment `json:"comments"`
	Files    []File    `json:"files"`

	SlackResponse
}

// NewGetFilesParameters provides an instance of GetFilesParameters with all the sane default values set
func NewGetFilesParameters() GetFilesParameters {
	return GetFilesParameters{
		User:          DEFAULT_FILES_USER,
		Channel:       DEFAULT_FILES_CHANNEL,
		TimestampFrom: DEFAULT_FILES_TS_FROM,
		TimestampTo:   DEFAULT_FILES_TS_TO,
		Types:         DEFAULT_FILES_TYPES,
		Count:         DEFAULT_FILES_COUNT,
		Page:          DEFAULT_FILES_PAGE,
	}
}

func fileRequest(ctx context.Context, client HTTPRequester, path string, values url.Values, debug bool) (*fileResponseFull, error) {
	response := &fileResponseFull{}
	err := postForm(ctx, client, SLACK_API+path, values, response, debug)
	if err != nil {
		return nil, err
	}
	if !response.Ok {
		return nil, errors.New(response.Error)
	}
	return response, nil
}

// GetFileInfo retrieves a file and related comments
func (api *Client) GetFileInfo(fileID string, count, page int) (*File, []Comment, *Paging, error) {
	return api.GetFileInfoContext(context.Background(), fileID, count, page)
}

// GetFileInfoContext retrieves a file and related comments with a custom context
func (api *Client) GetFileInfoContext(ctx context.Context, fileID string, count, page int) (*File, []Comment, *Paging, error) {
	values := url.Values{
		"token": {api.token},
		"file":  {fileID},
		"count": {strconv.Itoa(count)},
		"page":  {strconv.Itoa(page)},
	}

	response, err := fileRequest(ctx, api.httpclient, "files.info", values, api.debug)
	if err != nil {
		return nil, nil, nil, err
	}
	return &response.File, response.Comments, &response.Paging, nil
}

// GetFiles retrieves all files according to the parameters given
func (api *Client) GetFiles(params GetFilesParameters) ([]File, *Paging, error) {
	return api.GetFilesContext(context.Background(), params)
}

// GetFilesContext retrieves all files according to the parameters given with a custom context
func (api *Client) GetFilesContext(ctx context.Context, params GetFilesParameters) ([]File, *Paging, error) {
	values := url.Values{
		"token": {api.token},
	}
	if params.User != DEFAULT_FILES_USER {
		values.Add("user", params.User)
	}
	if params.Channel != DEFAULT_FILES_CHANNEL {
		values.Add("channel", params.Channel)
	}
	if params.TimestampFrom != DEFAULT_FILES_TS_FROM {
		values.Add("ts_from", strconv.FormatInt(int64(params.TimestampFrom), 10))
	}
	if params.TimestampTo != DEFAULT_FILES_TS_TO {
		values.Add("ts_to", strconv.FormatInt(int64(params.TimestampTo), 10))
	}
	if params.Types != DEFAULT_FILES_TYPES {
		values.Add("types", params.Types)
	}
	if params.Count != DEFAULT_FILES_COUNT {
		values.Add("count", strconv.Itoa(params.Count))
	}
	if params.Page != DEFAULT_FILES_PAGE {
		values.Add("page", strconv.Itoa(params.Page))
	}

	response, err := fileRequest(ctx, api.httpclient, "files.list", values, api.debug)
	if err != nil {
		return nil, nil, err
	}
	return response.Files, &response.Paging, nil
}

// UploadFile uploads a file
func (api *Client) UploadFile(params FileUploadParameters) (file *File, err error) {
	return api.UploadFileContext(context.Background(), params)
}

// UploadFileContext uploads a file and setting a custom context
func (api *Client) UploadFileContext(ctx context.Context, params FileUploadParameters) (file *File, err error) {
	// Test if user token is valid. This helps because client.Do doesn't like this for some reason. XXX: More
	// investigation needed, but for now this will do.
	_, err = api.AuthTest()
	if err != nil {
		return nil, err
	}
	response := &fileResponseFull{}
	values := url.Values{
		"token": {api.token},
	}
	if params.Filetype != "" {
		values.Add("filetype", params.Filetype)
	}
	if params.Filename != "" {
		values.Add("filename", params.Filename)
	}
	if params.Title != "" {
		values.Add("title", params.Title)
	}
	if params.InitialComment != "" {
		values.Add("initial_comment", params.InitialComment)
	}
	if params.ThreadTimestamp != "" {
		values.Add("thread_ts", params.ThreadTimestamp)
	}
	if len(params.Channels) != 0 {
		values.Add("channels", strings.Join(params.Channels, ","))
	}
	if params.Content != "" {
		values.Add("content", params.Content)
		err = postForm(ctx, api.httpclient, SLACK_API+"files.upload", values, response, api.debug)
	} else if params.File != "" {
		err = postLocalWithMultipartResponse(ctx, api.httpclient, "files.upload", params.File, "file", values, response, api.debug)
	} else if params.Reader != nil {
		err = postWithMultipartResponse(ctx, api.httpclient, "files.upload", params.Filename, "file", values, params.Reader, response, api.debug)
	}
	if err != nil {
		return nil, err
	}
	if !response.Ok {
		return nil, errors.New(response.Error)
	}
	return &response.File, nil
}

// DeleteFileComment deletes a file's comment
func (api *Client) DeleteFileComment(commentID, fileID string) error {
	return api.DeleteFileCommentContext(context.Background(), fileID, commentID)
}

// DeleteFileCommentContext deletes a file's comment with a custom context
func (api *Client) DeleteFileCommentContext(ctx context.Context, fileID, commentID string) (err error) {
	if fileID == "" || commentID == "" {
		return errors.New("received empty parameters")
	}

	values := url.Values{
		"token": {api.token},
		"file":  {fileID},
		"id":    {commentID},
	}
	_, err = fileRequest(ctx, api.httpclient, "files.comments.delete", values, api.debug)
	return err
}

// DeleteFile deletes a file
func (api *Client) DeleteFile(fileID string) error {
	return api.DeleteFileContext(context.Background(), fileID)
}

// DeleteFileContext deletes a file with a custom context
func (api *Client) DeleteFileContext(ctx context.Context, fileID string) (err error) {
	values := url.Values{
		"token": {api.token},
		"file":  {fileID},
	}

	_, err = fileRequest(ctx, api.httpclient, "files.delete", values, api.debug)
	return err
}

// RevokeFilePublicURL disables public/external sharing for a file
func (api *Client) RevokeFilePublicURL(fileID string) (*File, error) {
	return api.RevokeFilePublicURLContext(context.Background(), fileID)
}

// RevokeFilePublicURLContext disables public/external sharing for a file with a custom context
func (api *Client) RevokeFilePublicURLContext(ctx context.Context, fileID string) (*File, error) {
	values := url.Values{
		"token": {api.token},
		"file":  {fileID},
	}

	response, err := fileRequest(ctx, api.httpclient, "files.revokePublicURL", values, api.debug)
	if err != nil {
		return nil, err
	}
	return &response.File, nil
}

// ShareFilePublicURL enabled public/external sharing for a file
func (api *Client) ShareFilePublicURL(fileID string) (*File, []Comment, *Paging, error) {
	return api.ShareFilePublicURLContext(context.Background(), fileID)
}

// ShareFilePublicURLContext enabled public/external sharing for a file with a custom context
func (api *Client) ShareFilePublicURLContext(ctx context.Context, fileID string) (*File, []Comment, *Paging, error) {
	values := url.Values{
		"token": {api.token},
		"file":  {fileID},
	}

	response, err := fileRequest(ctx, api.httpclient, "files.sharedPublicURL", values, api.debug)
	if err != nil {
		return nil, nil, nil, err
	}
	return &response.File, response.Comments, &response.Paging, nil
}
