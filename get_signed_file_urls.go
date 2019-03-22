package notionapi

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"strings"
)

type getSignedFileUrlsRequest struct {
	Urls []getSignedFileURL `json:"urls"`
}

type getSignedFileURL struct {
	URL string `json:"url"`
}

type GetSignedFileUrlsResponse struct {
	SignedUrls []string `json:"signedUrls"`
	RawJSON    []byte   `json:"-"`
}

// GetSignedFileUrls executes a raw API call /api/v3/getSignedFileUrls
// For files (e.g. images) stored in Notion we need to get a temporary
// download url (which will be valid for only a short period of time)
func (c *Client) GetSignedFileUrls(urls []string) (*GetSignedFileUrlsResponse, error) {
	req := &getSignedFileUrlsRequest{}
	for _, url := range urls {
		fu := getSignedFileURL{URL: url}
		req.Urls = append(req.Urls, fu)
	}

	apiURL := "/api/v3/getSignedFileUrls"
	var rsp GetSignedFileUrlsResponse
	var err error
	rsp.RawJSON, err = doNotionAPI(c, apiURL, req, &rsp)
	if err != nil {
		return nil, err
	}
	return &rsp, nil
}

// DownloadFileResponse is a result of DownloadFile()
type DownloadFileResponse struct {
	Data    []byte
	Headers http.Header
}

// see if the url is for a file stored in notion
func isNotionURL(uri string) bool {
	return strings.HasPrefix(uri, "https://www.notion.so/")
}

func (c *Client) DownloadFile(uri string) (*DownloadFileResponse, error) {
	if isNotionURL(uri) {
		rsp, err := c.GetSignedFileUrls([]string{uri})
		if err != nil {
			// TODO: can it be that it returns an error because the url
			// doesn't need signed url
			return nil, err
		}
		uri = rsp.SignedUrls[0]
	}

	req, err := http.NewRequest("GET", uri, nil)
	if err != nil {
		return nil, err
	}
	if c.AuthToken != "" {
		req.Header.Set("cookie", fmt.Sprintf("token_v2=%v", c.AuthToken))
	}
	httpClient := c.getHTTPClient()
	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("http GET '%s' failed with status %s", uri, resp.Status)
	}
	var buf bytes.Buffer
	_, err = io.Copy(&buf, resp.Body)
	if err != nil {
		return nil, err
	}
	rsp := &DownloadFileResponse{
		Data:    buf.Bytes(),
		Headers: resp.Header,
	}
	return rsp, nil
}
