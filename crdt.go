package notionapi

import (
	"encoding/json"
	"sort"
	"unicode/utf16"
)

// Newer versions of Notion API return text properties (e.g. "title")
// also as crdt_data. It's a CRDT representation of the text. We convert
// it to "properties" format ([["text", [["b"], ["a", "url"]]], ...])
// so that it can be parsed with ParseTextSpans().
//
// The structure is:
//
//	{ "r": "<root node key>", "n": { "<node key>": { "s": { "i": [runs...] }, "c": [children...] } } }
//
// a run of text is:
//
//	{ "t": "t", "i": [<seq>, <offset>], "l": <len>, "c": "text", "b": [annotations...] }
//
// where "i" is the id of the first character (each character has an id
// <seq>, <offset+n>) and "l" is negative for deleted text.
// An annotation is:
//
//	{ "a": [attr...], "s": { "a": "b"|"a", "i": <char id> }, "e": { "a": "b"|"a", "i": <char id> } }
//
// i.e. formatting attribute applied to the range of characters from
// start anchor to end anchor. Anchor "b" means before a character,
// "a" means after. Annotations can span multiple runs.

type crdtRun struct {
	seq     string
	off     int
	n       int // number of utf-16 units (also for deleted runs)
	base    int // position in the resulting text (in utf-16 units)
	deleted bool
	text    []uint16
}

type crdtAnchor struct {
	after bool
	seq   string
	off   int
	// "start" / "end"
	special string
}

type crdtAnnotation struct {
	attr  []interface{}
	start crdtAnchor
	end   crdtAnchor
}

func parseCrdtID(v interface{}) (seq string, off int, ok bool) {
	a, ok := v.([]interface{})
	if !ok || len(a) != 2 {
		return "", 0, false
	}
	seq, ok = a[0].(string)
	if !ok {
		return "", 0, false
	}
	f, ok := a[1].(float64)
	if !ok {
		return "", 0, false
	}
	return seq, int(f), true
}

func parseCrdtAnchor(v interface{}) crdtAnchor {
	var res crdtAnchor
	m, _ := v.(map[string]interface{})
	if m == nil {
		return res
	}
	res.after = m["a"] == "a"
	if s, ok := m["i"].(string); ok {
		res.special = s
		return res
	}
	res.seq, res.off, _ = parseCrdtID(m["i"])
	return res
}

type crdtDoc struct {
	runs        []*crdtRun
	annotations []*crdtAnnotation
	total       int
	// the same annotation is listed in every run it covers
	// so we de-duplicate them by id
	seenAnnotations map[string]bool
}

func (d *crdtDoc) addNode(node map[string]interface{}, nodes map[string]interface{}, visited map[string]bool) {
	s, _ := node["s"].(map[string]interface{})
	items, _ := s["i"].([]interface{})
	for _, it := range items {
		run, _ := it.(map[string]interface{})
		if run["t"] != "t" {
			continue
		}
		seq, off, ok := parseCrdtID(run["i"])
		if !ok {
			continue
		}
		l, _ := run["l"].(float64)
		r := &crdtRun{seq: seq, off: off, base: d.total}
		if text, ok := run["c"].(string); ok && l >= 0 {
			r.text = utf16.Encode([]rune(text))
			r.n = len(r.text)
			d.total += r.n
		} else {
			r.deleted = true
			r.n = int(l)
			if r.n < 0 {
				r.n = -r.n
			}
		}
		d.runs = append(d.runs, r)
		annotations, _ := run["b"].([]interface{})
		for _, a := range annotations {
			am, _ := a.(map[string]interface{})
			attr, ok := am["a"].([]interface{})
			if !ok || len(attr) == 0 {
				continue
			}
			var key string
			if id, ok := am["i"]; ok {
				kd, _ := json.Marshal(id)
				key = string(kd)
			} else {
				kd, _ := json.Marshal([]interface{}{attr, am["s"], am["e"]})
				key = string(kd)
			}
			if d.seenAnnotations[key] {
				continue
			}
			d.seenAnnotations[key] = true
			d.annotations = append(d.annotations, &crdtAnnotation{
				attr:  attr,
				start: parseCrdtAnchor(am["s"]),
				end:   parseCrdtAnchor(am["e"]),
			})
		}
	}
	children, _ := node["c"].([]interface{})
	for _, c := range children {
		switch child := c.(type) {
		case string:
			if visited[child] {
				continue
			}
			visited[child] = true
			if cn, ok := nodes[child].(map[string]interface{}); ok {
				d.addNode(cn, nodes, visited)
			}
		case map[string]interface{}:
			d.addNode(child, nodes, visited)
		}
	}
}

// resolve anchor to a position in the resulting text
func (d *crdtDoc) resolve(a crdtAnchor) (int, bool) {
	switch a.special {
	case "start":
		return 0, true
	case "end":
		return d.total, true
	}
	for _, r := range d.runs {
		if r.seq != a.seq || a.off < r.off || a.off >= r.off+r.n {
			continue
		}
		if r.deleted {
			return r.base, true
		}
		pos := r.base + (a.off - r.off)
		if a.after {
			pos++
		}
		return pos, true
	}
	return 0, false
}

func (d *crdtDoc) toPropertyValue() interface{} {
	if d.total == 0 {
		return nil
	}
	text := make([]uint16, 0, d.total)
	for _, r := range d.runs {
		if !r.deleted {
			text = append(text, r.text...)
		}
	}
	// attributes for each utf-16 unit
	attrs := make([][]interface{}, d.total)
	for _, an := range d.annotations {
		start, ok1 := d.resolve(an.start)
		end, ok2 := d.resolve(an.end)
		if !ok1 || !ok2 {
			continue
		}
		for i := start; i < end && i < d.total; i++ {
			if i >= 0 {
				attrs[i] = append(attrs[i], an.attr)
			}
		}
	}
	// merge consecutive units with the same attributes into spans
	var res []interface{}
	start := 0
	for i := 1; i <= d.total; i++ {
		if i < d.total && attrsEqual(attrs[i], attrs[start]) {
			continue
		}
		span := []interface{}{string(utf16.Decode(text[start:i]))}
		if len(attrs[start]) > 0 {
			span = append(span, attrs[start])
		}
		res = append(res, span)
		start = i
	}
	return res
}

func attrsEqual(a, b []interface{}) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		aj, _ := json.Marshal(a[i])
		bj, _ := json.Marshal(b[i])
		if string(aj) != string(bj) {
			return false
		}
	}
	return true
}

// crdtToPropertyValue converts a crdt_data field to the format used by
// "properties" i.e. [["text", [["b"], ["a", "url"]]], ...].
// Returns nil if there is no text.
func crdtToPropertyValue(v interface{}) interface{} {
	m, ok := v.(map[string]interface{})
	if !ok {
		return nil
	}
	nodes, _ := m["n"].(map[string]interface{})
	if len(nodes) == 0 {
		return nil
	}
	d := &crdtDoc{seenAnnotations: map[string]bool{}}
	visited := map[string]bool{}
	rootKey, _ := m["r"].(string)
	if root, ok := nodes[rootKey].(map[string]interface{}); ok {
		visited[rootKey] = true
		d.addNode(root, nodes, visited)
	} else {
		// no root: process all nodes in a stable order
		keys := make([]string, 0, len(nodes))
		for k := range nodes {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		for _, k := range keys {
			if nm, ok := nodes[k].(map[string]interface{}); ok && !visited[k] {
				visited[k] = true
				d.addNode(nm, nodes, visited)
			}
		}
	}
	return d.toPropertyValue()
}

// getCrdtProperty returns value of a text property from crdt_data
// in "properties" format, or nil if not present
func (b *Block) getCrdtProperty(name string) interface{} {
	if b == nil || b.RawJSON == nil {
		return nil
	}
	cd, _ := b.RawJSON["crdt_data"].(map[string]interface{})
	if cd == nil {
		return nil
	}
	v, ok := cd[name]
	if !ok {
		return nil
	}
	return crdtToPropertyValue(v)
}
