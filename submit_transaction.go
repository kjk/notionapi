package notionapi

type submitTransactionRequest struct {
	Operations []*Operation `json:"operations"`
}

// Operation describes a single operation sent
type Operation struct {
	ID      string      `json:"id"`      // id of the block being modified
	Table   string      `json:"table"`   // "block" etc.
	Path    []string    `json:"path"`    // e.g. ["properties", "title"]
	Command string      `json:"command"` // "set", "update"
	Args    interface{} `json:"args"`
}

func (c *Client) SubmitTransaction(ops []*Operation) error {
	req := &submitTransactionRequest{
		Operations: ops,
	}
	// response is empty, as far as I can tell
	var rsp map[string]interface{}
	apiURL := "/api/v3/submitTransaction"
	_, err := doNotionAPI(c, apiURL, req, &rsp)
	return err
}

// this is title for
func buildSetTitleOp(id string, title string) *Operation {
	return &Operation{
		ID:      id,
		Table:   "block",
		Path:    []string{"properties", "title"},
		Command: "set",
		Args: []interface{}{
			[]string{title},
		},
	}
}

/*
// last_edited_time seems to be Unix() * 1000.
// It doesn't matter if we do UTC() or not
func notionTimeNow() int64 {
	return time.Now().Unix() * 1000
}

func buildLastEditedTimeOp(id string) *Operation {
	args := map[string]interface{}{
		"last_edited_time": notionTimeNow(),
	}
	return &Operation{
		ID:      id,
		Table:   "block",
		Path:    []string{},
		Command: "update",
		Args:    args,
	}
}
*/

func buildSetPageFormat(id string, args map[string]interface{}) *Operation {
	return &Operation{
		ID:      id,
		Table:   "block",
		Path:    []string{"format"},
		Command: "update",
		Args:    args,
	}
}

/*
// TODO: add constants for known languages
func buildUpdateCodeBlockLang(id string, lang string) *Operation {
	args := map[string]interface{}{
		"language": []string{lang},
	}
	return &Operation{
		ID:      id,
		Table:   "block",
		Path:    []string{"properties"},
		Command: "update",
		Args:    args,
	}
}
*/
