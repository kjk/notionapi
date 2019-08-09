package notionapi

import (
	"bytes"
	"encoding/json"
	"io"
	"io/ioutil"
	"net/http"
	"time"
)

type RequestResponse struct {
	Method string `json:"method"`
	URL    string `json:"url"`
	Body   []byte `json:"body"`

	Response []byte      `json:"response"`
	Header   http.Header `json:"header"`
}

type HTTPCache struct {
	CachedRequests       []*RequestResponse `json:"cached_requests"`
	RequestsFromCache    int                `json:"-"`
	RequestsNotFromCache int                `json:"-"`
}

func NewHTTPCache() *HTTPCache {
	return &HTTPCache{}
}

func (c *HTTPCache) Add(rr *RequestResponse) {
	c.CachedRequests = append(c.CachedRequests, rr)
}

// closeableBuffer adds Close() error method to bytes.Buffer
// to satisfy io.ReadCloser interface
type closeableBuffer struct {
	*bytes.Buffer
}

func (b *closeableBuffer) Close() error {
	return nil
}

func readAndReplaceReadCloser(pBody *io.ReadCloser) ([]byte, error) {
	// have to read the body from r and put it back
	var err error
	body := *pBody
	d, err := ioutil.ReadAll(body)
	if err != nil {
		return nil, err
	}
	buf := &closeableBuffer{bytes.NewBuffer(d)}
	*pBody = buf
	return d, nil
}

// pretty-print if valid JSON. If not, return unchanged
func ppJSON2(js []byte) []byte {
	var m map[string]interface{}
	err := json.Unmarshal(js, &m)
	if err != nil {
		return js
	}
	d, err := json.MarshalIndent(m, "", "  ")
	if err != nil {
		return js
	}
	return d
}

func isBodySame(r *http.Request, rr *RequestResponse, cachedBody *[]byte) (bool, error) {
	// only POST request takes body
	if r.Method != http.MethodPost {
		return true, nil
	}
	if r.Body == nil && len(rr.Body) == 0 {
		return true, nil
	}

	d := *cachedBody
	if d == nil {
		var err error
		d, err = readAndReplaceReadCloser(&r.Body)
		if err != nil {
			return false, err
		}
		if d == nil {
			*cachedBody = []byte{}
		} else {
			d = ppJSON2(d)
			*cachedBody = d
		}
	}
	rrBody := ppJSON2(rr.Body)
	return bytes.Equal(d, rrBody), nil
}

func isCachedRequest(r *http.Request, rr *RequestResponse, cachedBody *[]byte) (bool, error) {
	if rr.Method != r.Method {
		return false, nil
	}
	uri1 := rr.URL
	uri2 := r.URL.String()
	if uri1 != uri2 {
		return false, nil
	}
	return isBodySame(r, rr, cachedBody)
}

func (c *HTTPCache) findCachedResponse(r *http.Request, cachedBody *[]byte) (*RequestResponse, error) {
	for _, rr := range c.CachedRequests {
		same, err := isCachedRequest(r, rr, cachedBody)
		if err != nil {
			return nil, err
		}
		if same {
			return rr, nil
		}
	}
	return nil, nil
}

type cachingTransport struct {
	cache     *HTTPCache
	transport http.RoundTripper
}

func (t *cachingTransport) cachedRoundTrip(r *http.Request, cachedRequestBody []byte) (*http.Response, error) {
	transport := t.transport
	if transport == nil {
		transport = http.DefaultTransport
	}
	if cachedRequestBody == nil {
		var err error
		cachedRequestBody, err = readAndReplaceReadCloser(&r.Body)
		if err != nil {
			return nil, err
		}
	}
	rsp, err := transport.RoundTrip(r)
	if err != nil {
		return rsp, err
	}

	// only cache 200 responses
	if rsp.StatusCode != 200 {
		return rsp, nil
	}

	d, err := readAndReplaceReadCloser(&rsp.Body)
	if err != nil {
		return nil, err
	}

	rr := &RequestResponse{
		Method: r.Method,
		URL:    r.URL.String(),
		Body:   cachedRequestBody,

		Response: d,
		Header:   rsp.Header,
	}
	t.cache.Add(rr)
	t.cache.RequestsNotFromCache++
	return rsp, nil
}

func (t *cachingTransport) RoundTrip(r *http.Request) (*http.Response, error) {
	var cachedRequestBody []byte
	rr, err := t.cache.findCachedResponse(r, &cachedRequestBody)
	if err != nil {
		t.cache.RequestsFromCache++
		return nil, err
	}

	if rr == nil {
		return t.cachedRoundTrip(r, cachedRequestBody)
	}

	d := rr.Response
	rsp := &http.Response{
		Status:        "200",
		StatusCode:    200,
		Header:        rr.Header,
		Body:          &closeableBuffer{bytes.NewBuffer(d)},
		ContentLength: int64(len(d)),
	}
	return rsp, nil
}

func NewCachingHTTPClient(cache *HTTPCache) *http.Client {
	c := *http.DefaultClient
	c.Timeout = time.Second * 30
	origTransport := c.Transport
	c.Transport = &cachingTransport{
		cache:     cache,
		transport: origTransport,
	}
	return &c
}
