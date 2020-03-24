package notionapi

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
)

const (
	signedURLPrefix = "https://www.notion.so/signed/"
	s3URLPrefix     = "https://s3-us-west-2.amazonaws.com/secure.notion-static.com/"
	//s3URLPrefixEncoded = "https://s3.us-west-2.amazonaws.com/secure.notion-static.com/"
)

/*
{"urls": [
	{
		"url" : "https://s3-us-west-2.amazonaws.com/secure.notion-static.com/e5661303-82e1-43e4-be8e-662d1598cd53/untitled",
		"permissionRecord": {
			"table":"block",
			"id":"845fd39c-f048-4d41-a0f8-a83851d17afd"
			}
		}
  ]}
*/

type getSignedURLPermissionRecord struct {
	Table string `json:"table"`
	ID    string `json:"id"`
}

type getSignedFileURL struct {
	URL        string                       `json:"url"`
	Permission getSignedURLPermissionRecord `json:"permissionRecord"`
}

type getSignedFileUrlsRequest struct {
	Urls []getSignedFileURL `json:"urls"`
}

// GetSignedFileUrlsResponse is a response of GetSignedFileUrls()
type GetSignedFileUrlsResponse struct {
	SignedUrls []string               `json:"signedUrls"`
	RawJSON    map[string]interface{} `json:"-"`
}

// GetSignedFileUrls executes a raw API call /api/v3/getSignedFileUrls
// For files (e.g. images) stored in Notion we need to get a temporary
// download url (which will be valid for only a short period of time)
func (c *Client) GetSignedFileUrls(urls []string, blockIDs []string) (*GetSignedFileUrlsResponse, error) {
	req := &getSignedFileUrlsRequest{}
	for i, url := range urls {
		perm := getSignedURLPermissionRecord{
			Table: "block",
			ID:    blockIDs[i],
		}
		fu := getSignedFileURL{URL: url, Permission: perm}
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
	URL           string
	CacheFileName string
	Data          []byte
	Header        http.Header
}

// sometimes image url in "source" is not accessible but can
// be accessed when proxied via notion server as
// www.notion.so/image/${source}
// This also allows resizing via ?width=${n} arguments
//
// from: /images/page-cover/met_vincent_van_gogh_cradle.jpg
// =>
// https://www.notion.so/image/https%3A%2F%2Fwww.notion.so%2Fimages%2Fpage-cover%2Fmet_vincent_van_gogh_cradle.jpg?width=3290
func maybeProxyImageURL(uri string) string {
	if strings.HasPrefix(uri, s3URLPrefix) {
		return signedURLPrefix + url.PathEscape(uri)
	}

	// don't proxy external images
	if !strings.Contains(uri, "notion.so") {
		return uri
	}

	if strings.Contains(uri, "//www.notion.so/image/") {
		return uri
	}

	// if the url has https://, it's already in s3.
	// If not, it's only a relative URL (like those for built-in
	// cover pages)
	if !strings.HasPrefix(uri, "https://") {
		uri = "https://www.notion.so" + uri
	}
	return "https://www.notion.so/image/" + url.PathEscape(uri)
}

func (c *Client) maybeSignImageURL(uri string, blockID string) string {
	if !strings.HasPrefix(uri, s3URLPrefix) {
		return maybeProxyImageURL(uri)
	}
	/* notionapi-py does:

	if url.startswith(S3_URL_PREFIX):
		url = SIGNED_URL_PREFIX + quote_plus(url)
		if client:
			url = client.session.head(url).headers.get("Location")
	*/
	rsp, err := c.GetSignedFileUrls([]string{uri}, []string{blockID})
	if err != nil {
		return uri
	}
	return rsp.SignedUrls[0]
}

func (c *Client) downloadFile(uri string) (*DownloadFileResponse, error) {
	req, err := http.NewRequest("GET", uri, nil)
	if err != nil {
		//fmt.Printf("DownloadFile: NewRequest() for '%s' failed with '%s'\n", uri, err)
		return nil, err
	}
	if c.AuthToken != "" {
		req.Header.Set("cookie", fmt.Sprintf("token_v2=%v", c.AuthToken))
	}
	httpClient := c.getHTTPClient()
	resp, err := httpClient.Do(req)
	if err != nil {
		//fmt.Printf("DownloadFile: httpClient.Do() for '%s' failed with '%s'\n", uri, err)
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 400 {
		//fmt.Printf("DownloadFile: httpClient.Do() for '%s' failed with '%s'\n", uri, resp.Status)
		return nil, fmt.Errorf("http GET '%s' failed with status %s", uri, resp.Status)
	}
	var buf bytes.Buffer
	_, err = io.Copy(&buf, resp.Body)
	if err != nil {
		return nil, err
	}
	rsp := &DownloadFileResponse{
		Data:   buf.Bytes(),
		Header: resp.Header,
	}
	return rsp, nil
}

// DownloadFile downloads a file stored in Notion
func (c *Client) DownloadFile(uri string, blockID string) (*DownloadFileResponse, error) {
	//fmt.Printf("DownloadFile: '%s'\n", uri)
	// try proxing through www.notion.so/image/
	if strings.Contains(uri, "s3.us-west-2.amazonaws.com") {
		uri2 := "https://www.notion.so/image/" + url.PathEscape(uri)
		res, err := c.downloadFile(uri2)
		if err == nil {
			return res, nil
		}
	}
	uri2 := c.maybeSignImageURL(uri, blockID)
	return c.downloadFile(uri2)
}
