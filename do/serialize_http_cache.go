package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"

	"github.com/kjk/notionapi"
	"github.com/kjk/siser"
)

const (
	recCacheName = "httpcache-v1"
)

// pretty-print if valid JSON. If not, return unchanged
func ppJSON(js []byte) []byte {
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

func serializeHTTPCache(c *notionapi.HTTPCache) ([]byte, error) {
	if len(c.CachedRequests) == 0 {
		return []byte{}, nil
	}
	buf := bytes.NewBuffer(nil)
	w := siser.NewWriter(buf)
	var r siser.Record
	for _, rr := range c.CachedRequests {
		r.Reset()
		body := ppJSON(rr.Body)
		response := ppJSON(rr.Response)
		hdr, err := json.MarshalIndent(rr.Header, "", "  ")
		if err != nil {
			return nil, err
		}
		r.Append("Method", rr.Method)
		r.Append("URL", rr.URL)
		r.Append("Body", string(body))
		r.Append("Response", string(response))
		r.Append("HeadersJSON", string(hdr))
		r.Name = recCacheName
		_, err = w.WriteRecord(&r)
		if err != nil {
			return nil, err
		}
	}
	d := buf.Bytes()
	return d, nil
}

func deserializeHTTPCache(d []byte) (*notionapi.HTTPCache, error) {
	res := &notionapi.HTTPCache{}
	br := bufio.NewReader(bytes.NewBuffer(d))
	r := siser.NewReader(br)
	var err error
	for r.ReadNextRecord() {
		if r.Name != recCacheName {
			return nil, fmt.Errorf("unexpected record type '%s', wanted '%s'", r.Name, recCacheName)
		}
		rr := notionapi.RequestResponse{}
		rr.Method = recGetKey(r.Record, "Method", &err)
		rr.URL = recGetKey(r.Record, "URL", &err)
		rr.Body = recGetKeyBytes(r.Record, "Body", &err)
		rr.Response = recGetKeyBytes(r.Record, "Response", &err)
		hdrsJSON := recGetKeyBytes(r.Record, "HeadersJSON", &err)
		if err != nil {
			return nil, err
		}
		err = json.Unmarshal(hdrsJSON, &rr.Header)
		if err != nil {
			return nil, err
		}
	}
	if r.Err() != nil {
		return nil, r.Err()
	}
	return res, nil
}
