package notionapi

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"

	"github.com/kjk/common/require"
)

// sort attrs within each span so that their order doesn't matter
func normalizeSpansForTest(v interface{}) string {
	a, _ := v.([]interface{})
	for _, sp := range a {
		s, _ := sp.([]interface{})
		if len(s) > 1 {
			attrs, _ := s[1].([]interface{})
			sort.Slice(attrs, func(i, j int) bool {
				x, _ := json.Marshal(attrs[i])
				y, _ := json.Marshal(attrs[j])
				return string(x) < string(y)
			})
		}
	}
	d, _ := json.Marshal(a)
	return string(d)
}

// for every block in cached test data that has both "properties" and
// "crdt_data", verify that we decode crdt_data to the same value
func TestCrdtMatchesProperties(t *testing.T) {
	files, err := filepath.Glob(filepath.Join("caching_client_testdata", "*"+cacheFileExt))
	require.NoError(t, err)
	require.NotEmpty(t, files)

	nChecked := 0
	for _, path := range files {
		d, err := os.ReadFile(path)
		require.NoError(t, err)
		entries, err := deserializeCacheEntry(d)
		require.NoError(t, err)
		for _, e := range entries {
			var rsp struct {
				RecordMap map[string]json.RawMessage `json:"recordMap"`
			}
			if err := json.Unmarshal(e.Response, &rsp); err != nil {
				continue
			}
			var blocks map[string]struct {
				Value struct {
					Value map[string]interface{} `json:"value"`
				} `json:"value"`
			}
			if err := json.Unmarshal(rsp.RecordMap["block"], &blocks); err != nil {
				continue
			}
			for _, rec := range blocks {
				v := rec.Value.Value
				crdt, _ := v["crdt_data"].(map[string]interface{})
				props, _ := v["properties"].(map[string]interface{})
				for name, cv := range crdt {
					got := crdtToPropertyValue(cv)
					exp, ok := props[name]
					if !ok {
						// no text in properties => no text in crdt
						require.Nil(t, got, "block %s, property %s", v["id"], name)
						continue
					}
					require.Equal(t, normalizeSpansForTest(exp), normalizeSpansForTest(got), "block %s, property %s", v["id"], name)
					nChecked++
				}
			}
		}
	}
	require.True(t, nChecked > 0)
}

func TestCrdtDeletedText(t *testing.T) {
	// title of a page where all the text was deleted
	s := `{"n":{"VSUoTnjRHGfD_Eaz9ToTzg,\"start\",\"end\"":{"c":[],"s":{"i":[{"t":"s"},{"i":["47klL0kpPiOq",1],"l":-7,"o":"start","t":"t"},{"i":["47klL0kpPiOq",9],"l":-13,"o":["47klL0kpPiOq",7],"t":"t"},{"i":["47klL0kpPiOq",8],"l":-1,"o":["47klL0kpPiOq",7],"t":"t"},{"t":"e"}],"l":"","x":"VSUoTnjRHGfD_Eaz9ToTzg"}}},"r":"VSUoTnjRHGfD_Eaz9ToTzg,\"start\",\"end\""}`
	var v interface{}
	require.NoError(t, json.NewDecoder(strings.NewReader(s)).Decode(&v))
	require.Nil(t, crdtToPropertyValue(v))
}

func TestCrdtFormatting(t *testing.T) {
	// "Mention a page that is not a sub-page: ‣ " where ‣ is a page mention
	s := `{"n":{"JeMwNGvXpq31O70ZI1olOw,\"start\",\"end\"":{"c":[],"s":{"i":[{"t":"s"},{"c":"Mention a page that is not a sub-page: ","i":["_uKUoUDxuaXD",1],"l":39,"o":"start","t":"t"},{"a":[],"b":[{"a":["p","4c6a54c6-8b3e-4ea2-af9c-faabcc88d58d","bc202e06-6caa-4e3f-81eb-f226ab5deef7"],"e":{"a":"a","i":["_uKUoUDxuaXD",40]},"i":["_uKUoUDxuaXD",42],"l":"","s":{"a":"b","i":["_uKUoUDxuaXD",40]},"t":"a","x":"JeMwNGvXpq31O70ZI1olOw"}],"c":"‣","i":["_uKUoUDxuaXD",40],"l":1,"o":["_uKUoUDxuaXD",39],"t":"t"},{"c":" ","i":["_uKUoUDxuaXD",41],"l":1,"o":["_uKUoUDxuaXD",40],"t":"t"},{"t":"e"}],"l":"","x":"JeMwNGvXpq31O70ZI1olOw"}}},"r":"JeMwNGvXpq31O70ZI1olOw,\"start\",\"end\""}`
	var v interface{}
	require.NoError(t, json.NewDecoder(strings.NewReader(s)).Decode(&v))
	got := crdtToPropertyValue(v)
	exp := `[["Mention a page that is not a sub-page: "],["‣",[["p","4c6a54c6-8b3e-4ea2-af9c-faabcc88d58d","bc202e06-6caa-4e3f-81eb-f226ab5deef7"]]],[" "]]`
	require.Equal(t, exp, normalizeSpansForTest(got))

	// GetProperty() falls back to crdt_data when "properties" is missing
	b := &Block{RawJSON: map[string]interface{}{"crdt_data": map[string]interface{}{"title": v}}}
	spans := b.GetTitle()
	require.Equal(t, 3, len(spans))
	require.Equal(t, "Mention a page that is not a sub-page: ", spans[0].Text)
	require.Equal(t, TextSpanSpecial, spans[1].Text)
	require.Equal(t, AttrPage, spans[1].Attrs[0][0])
}
