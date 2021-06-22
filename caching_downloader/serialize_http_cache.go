package caching_downloader

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"

	"github.com/kjk/caching_http_client"
	"github.com/kjk/siser"
	"github.com/tidwall/pretty"
)

const (
	recCacheName = "httpcache-v1"
)

var prettyOpts = pretty.Options{
	Width:  80,
	Prefix: "",
	Indent: "  ",
	// sorting keys only slightly slower
	SortKeys: true,
}

// pretty-print if valid JSON. If not, return unchanged
// about 4x faster than naive version using json.Unmarshal() + json.Marshal()
func ppJSON(js []byte) []byte {
	if !json.Valid(js) {
		return js
	}
	return pretty.PrettyOptions(js, &prettyOpts)
}
func recGetKey(r *siser.Record, key string, pErr *error) string {
	if *pErr != nil {
		return ""
	}
	v, ok := r.Get(key)
	if !ok {
		*pErr = fmt.Errorf("didn't find key '%s'", key)
	}
	return v
}

func recGetKeyBytes(r *siser.Record, key string, pErr *error) []byte {
	return []byte(recGetKey(r, key, pErr))
}

func serializeHTTPCache(c *caching_http_client.Cache) ([]byte, error) {
	if len(c.CachedRequests) == 0 {
		return []byte{}, nil
	}
	buf := bytes.NewBuffer(nil)
	w := siser.NewWriter(buf)
	w.NoTimestamp = true
	var r siser.Record
	for _, rr := range c.CachedRequests {
		r.Reset()
		body := ppJSON(rr.Body)
		response := ppJSON(rr.Response)
		r.Write("Method", rr.Method)
		r.Write("URL", rr.URL)
		r.Write("Body", string(body))
		r.Write("Response", string(response))
		r.Name = recCacheName
		_, err := w.WriteRecord(&r)
		if err != nil {
			return nil, err
		}
	}
	d := buf.Bytes()
	return d, nil
}

func deserializeHTTPCache(d []byte) (*caching_http_client.Cache, error) {
	res := &caching_http_client.Cache{}
	br := bufio.NewReader(bytes.NewBuffer(d))
	r := siser.NewReader(br)
	r.NoTimestamp = true
	var err error
	for r.ReadNextRecord() {
		if r.Name != recCacheName {
			return nil, fmt.Errorf("unexpected record type '%s', wanted '%s'", r.Name, recCacheName)
		}
		rr := &caching_http_client.RequestResponse{}
		rr.Method = recGetKey(r.Record, "Method", &err)
		rr.URL = recGetKey(r.Record, "URL", &err)
		rr.Body = recGetKeyBytes(r.Record, "Body", &err)
		rr.Response = recGetKeyBytes(r.Record, "Response", &err)
		res.Add(rr)
	}
	if r.Err() != nil {
		return nil, r.Err()
	}
	return res, nil
}
