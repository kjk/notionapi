package notionapi

import (
	"fmt"
	"io"
	"strings"
)

type NotionID struct {
	DashID   string
	NoDashID string
}

func NewNotionID(maybeID string) *NotionID {
	if IsValidDashID(maybeID) {
		return &NotionID{
			DashID:   maybeID,
			NoDashID: ToNoDashID(maybeID),
		}
	}
	if IsValidNoDashID(maybeID) {
		return &NotionID{
			DashID:   ToDashID(maybeID),
			NoDashID: maybeID,
		}
	}
	return nil
}

var (
	dashIDLen   = len("2131b10c-ebf6-4938-a127-7089ff02dbe4")
	noDashIDLen = len("2131b10cebf64938a1277089ff02dbe4")
)

// only hex chars seem to be valid
func isValidNoDashIDChar(c byte) bool {
	switch {
	case c >= '0' && c <= '9':
		return true
	case c >= 'a' && c <= 'f':
		return true
	case c >= 'A' && c <= 'F':
		// currently not used but just in case notion starts using them
		return true
	}
	return false
}

func isValidDashIDChar(c byte) bool {
	if c == '-' {
		return true
	}
	return isValidNoDashIDChar(c)
}

// IsValidDashID returns true if id looks like a valid Notion dash id
func IsValidDashID(id string) bool {
	if len(id) != dashIDLen {
		return false
	}
	if id[8] != '-' ||
		id[13] != '-' ||
		id[18] != '-' ||
		id[23] != '-' {
		return false
	}
	for i := range id {
		if !isValidDashIDChar(id[i]) {
			return false
		}
	}
	return true
}

// IsValidNoDashID returns true if id looks like a valid Notion no dash id
func IsValidNoDashID(id string) bool {
	if len(id) != noDashIDLen {
		return false
	}
	for i := range id {
		if !isValidNoDashIDChar(id[i]) {
			return false
		}
	}
	return true
}

// ToNoDashID converts 2131b10c-ebf6-4938-a127-7089ff02dbe4
// to 2131b10cebf64938a1277089ff02dbe4.
// If not in expected format, we leave it untouched
func ToNoDashID(id string) string {
	s := strings.Replace(id, "-", "", -1)
	if IsValidNoDashID(s) {
		return s
	}
	return ""
}

// ToDashID convert id in format bb760e2dd6794b64b2a903005b21870a
// to bb760e2d-d679-4b64-b2a9-03005b21870a
// If id is not in that format, we leave it untouched.
func ToDashID(id string) string {
	if IsValidDashID(id) {
		return id
	}
	s := strings.Replace(id, "-", "", -1)
	if len(s) != noDashIDLen {
		return id
	}
	res := id[:8] + "-" + id[8:12] + "-" + id[12:16] + "-" + id[16:20] + "-" + id[20:]
	return res
}

func isIDEqual(id1, id2 string) bool {
	id1 = ToNoDashID(id1)
	id2 = ToNoDashID(id2)
	return id1 == id2
}

func isSafeChar(r rune) bool {
	if r >= '0' && r <= '9' {
		return true
	}
	if r >= 'a' && r <= 'z' {
		return true
	}
	if r >= 'A' && r <= 'Z' {
		return true
	}
	return false
}

// SafeName returns a file-system safe name
func SafeName(s string) string {
	var res string
	for _, r := range s {
		if !isSafeChar(r) {
			res += "-"
		} else {
			res += string(r)
		}
	}
	// replace multi-dash with single dash
	for strings.Contains(res, "--") {
		res = strings.Replace(res, "--", "-", -1)
	}
	res = strings.TrimLeft(res, "-")
	res = strings.TrimRight(res, "-")
	return res
}

// ErrPageNotFound is returned by Client.DownloadPage if page
// cannot be found
type ErrPageNotFound struct {
	PageID string
}

func newErrPageNotFound(pageID string) *ErrPageNotFound {
	return &ErrPageNotFound{
		PageID: pageID,
	}
}

// Error return error string
func (e *ErrPageNotFound) Error() string {
	pageID := ToNoDashID(e.PageID)
	return fmt.Sprintf("couldn't retrieve page '%s'", pageID)
}

// IsErrPageNotFound returns true if err is an instance of ErrPageNotFound
func IsErrPageNotFound(err error) bool {
	_, ok := err.(*ErrPageNotFound)
	return ok
}

func closeNoError(c io.Closer) {
	_ = c.Close()
}

// log JSON after pretty printing it
func logJSON(client *Client, js []byte) {
	//client.vlogf("%s\n\n", string(js))
	pp := string(PrettyPrintJS(js))
	client.vlogf("%s\n\n", pp)
}
