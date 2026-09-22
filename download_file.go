package notionapi

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
)

// DownloadFileResponse is a result of DownloadFile()
type DownloadFileResponse struct {
	URL           string
	CacheFilePath string
	Data          []byte
	Header        http.Header
	FromCache     bool
}

// GetFileToken returns the value of file_token cookie, needed to
// download files from file.notion.so (e.g. exported pages).
// Requires AuthToken. The result is cached in FileToken.
func (c *Client) GetFileToken() (string, error) {
	if c.FileToken != "" {
		return c.FileToken, nil
	}
	if c.AuthToken == "" {
		return "", errors.New("AuthToken is required to get file token")
	}
	// loadUserContent sets file_token cookie in the response
	uri := notionHost + "/api/v3/loadUserContent"
	req, err := http.NewRequest("POST", uri, strings.NewReader("{}"))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", userAgent)
	req.Header.Set("cookie", "token_v2="+c.AuthToken)
	rsp, err := c.getHTTPClient().Do(req)
	if err != nil {
		return "", err
	}
	defer closeNoError(rsp.Body)
	if rsp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("http POST '%s' failed with status %s", uri, rsp.Status)
	}
	for _, cookie := range rsp.Cookies() {
		if cookie.Name == "file_token" && cookie.Value != "" {
			c.FileToken = cookie.Value
			return c.FileToken, nil
		}
	}
	return "", fmt.Errorf("file_token cookie not found in response from '%s'", uri)
}

// urls like https://file.notion.so/... require file_token cookie
func isNotionFileURL(uri string) bool {
	u, err := url.Parse(uri)
	if err != nil {
		return false
	}
	return strings.HasPrefix(u.Host, "file.notion.")
}

// DownloadURLStream downloads a given url with possibly authenticated client and returns a stream
// The caller is responsible for closing the Response.Body when done
func (c *Client) DownloadURLStream(uri string) (*http.Response, error) {
	req, err := http.NewRequest("GET", uri, nil)
	if err != nil {
		return nil, err
	}
	if c.AuthToken != "" && c.FileToken == "" && isNotionFileURL(uri) {
		// best effort: if this fails, the download will most likely
		// fail with 403 and report that
		_, _ = c.GetFileToken()
	}
	if cookie := c.downloadCookie(); cookie != "" {
		req.Header.Set("cookie", cookie)
	}
	httpClient := c.getHTTPClient()
	resp, err := httpClient.Do(req)
	if err != nil {
		//fmt.Printf("DownloadFile: httpClient.Do() for '%s' failed with '%s'\n", uri, err)
		return nil, err
	}
	if resp.StatusCode >= 400 {
		resp.Body.Close()
		//fmt.Printf("DownloadFile: httpClient.Do() for '%s' failed with '%s'\n", uri, resp.Status)
		return nil, fmt.Errorf("http GET '%s' failed with status %s", uri, resp.Status)
	}

	return resp, nil
}

// downloadCookie returns the cookie header used when downloading files.
// Downloads from file.notion.so (e.g. exported pages) require file_token
// cookie in addition to token_v2
func (c *Client) downloadCookie() string {
	var parts []string
	if c.AuthToken != "" {
		parts = append(parts, "token_v2="+c.AuthToken)
	}
	if c.FileToken != "" {
		parts = append(parts, "file_token="+c.FileToken)
	}
	return strings.Join(parts, "; ")
}

// DownloadURL downloads a given url with possibly authenticated client
func (c *Client) DownloadURL(uri string) (*DownloadFileResponse, error) {
	resp, err := c.DownloadURLStream(uri)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

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

const (
	notionImageProxy = "https://www.notion.so/image/"
	s3FileURLPrefix  = "https://s3-us-west-2.amazonaws.com/secure.notion-static.com/"
)

// sometimes image url in "source" is not accessible but can
// be accessed when proxied via notion server as
// www.notion.so/image/${source}?table=${parentTable}&id=${blockID}
// This also allows resizing via ?width=${n} arguments
func maybeProxyImageURL(uri string, block *Block) string {

	if strings.HasPrefix(uri, "https://cdn.dutchcowboys.nl/uploads") {
		return uri
	}
	if strings.HasPrefix(uri, "https://images.unsplash.com") {
		return uri
	}

	// TODO: not sure about this one anymore
	if strings.HasPrefix(uri, "https://www.notion.so/images/") {
		return uri
	}

	// from: /images/page-cover/met_vincent_van_gogh_cradle.jpg
	// =>
	// https://www.notion.so/image/https%3A%2F%2Fwww.notion.so%2Fimages%2Fpage-cover%2Fmet_vincent_van_gogh_cradle.jpg?width=3290
	if strings.HasPrefix(uri, "/images/page-cover/") {
		return "https://www.notion.so" + uri
	}

	if block == nil {
		return uri
	}
	blockID := block.ID
	parentTable := block.ParentTable
	spaceID := block.SpaceID

	if strings.HasPrefix(uri, notionImageProxy) {
		uri = uri + "?id=" + blockID + "&table=" + parentTable + "&spaceId=" + spaceID
		return uri
	} else if strings.HasPrefix(uri, "attachment:") {
		uri = notionImageProxy + url.QueryEscape(uri) + "?id=" + blockID + "&table=" + parentTable + "&spaceId=" + spaceID
		return uri
	}

	if !strings.Contains(uri, s3FileURLPrefix) {
		return uri
	}

	uri = notionImageProxy + url.PathEscape(uri) + "?id=" + blockID + "&table=" + parentTable + "&spaceId=" + spaceID
	return uri
}

// DownloadFile downloads a file stored in Notion referenced
// by a block with a given id and of a given block with a given
// parent table (data present in Block)
func (c *Client) DownloadFile(uri string, block *Block) (*DownloadFileResponse, error) {
	// first try downloading proxied url
	uri2 := maybeProxyImageURL(uri, block)
	res, err := c.DownloadURL(uri2)
	if err != nil && uri2 != uri {
		// otherwise just try your luck with original URL
		res, err = c.DownloadURL(uri)
	}
	if err != nil {
		rsp, err2 := c.GetSignedURLs([]string{uri}, block)
		if err2 != nil {
			return nil, err
		}
		if len(rsp.SignedURLS) == 0 {
			return nil, err
		}
		uri3 := rsp.SignedURLS[0]
		res, err = c.DownloadURL(uri3)
	}
	return res, err
}

// DownloadFileStream downloads a file stored in Notion and returns a stream for streaming operations
// The caller is responsible for closing the Response.Body when done
func (c *Client) DownloadFileStream(uri string, block *Block) (*http.Response, error) {
	// first try downloading proxied url
	uri2 := maybeProxyImageURL(uri, block)
	res, err := c.DownloadURLStream(uri2)
	if err != nil {
		rsp, err2 := c.GetSignedURLs([]string{uri}, block)
		if err2 != nil {
			return nil, err
		}
		if len(rsp.SignedURLS) == 0 {
			return nil, err
		}
		uri3 := rsp.SignedURLS[0]
		res, err = c.DownloadURLStream(uri3)
	}
	return res, err
}

// DownloadAttachmentStream downloads an attachment file stored in Notion and returns a stream for streaming operations
// The caller is responsible for closing the Response.Body when done
func (c *Client) DownloadAttachmentStream(uid string, block *Block) (*http.Response, error) {
	return c.DownloadURLStream(c.GetAttachmentURL(uid, block))
}

// GetAttachmentURL returns the URL for an attachment file stored in Notion referenced by a block
func (c *Client) GetAttachmentURL(uid string, block *Block) string {
	if uid == "" || block == nil {
		return ""
	}
	blockNew := *block
	blockNew.ParentTable = "block" // attachments are always in block table

	return maybeProxyImageURL(uid, &blockNew)
}
