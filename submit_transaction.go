package notionapi

import "time"

// Command Types
const (
	CommandSet        = "set"
	CommandUpdate     = "update"
	CommandListAfter  = "listAfter"
	CommandListRemove = "listRemove"
)

type submitTransactionRequest struct {
	Operations []*Operation `json:"operations"`
}

// Operation describes a single operation sent
type Operation struct {
	ID      string      `json:"id"`      // id of the block being modified
	Table   string      `json:"table"`   // "block" etc.
	Path    []string    `json:"path"`    // e.g. ["properties", "title"]
	Command string      `json:"command"` // "set", "update", "listAfter"
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

// Now returns now in micro seconds as expected by the notion API
func Now() int64 {
	return time.Now().Unix() * 1000
}

// buildOp creates an Operation for this block
func (b *Block) buildOp(command string, path []string, args interface{}) *Operation {
	return &Operation{
		ID:      b.ID,
		Table:   "block",
		Path:    path,
		Command: command,
		Args:    args,
	}
}

// SetTitleOp creates an Operation to set the title property
func (b *Block) SetTitleOp(title string) *Operation {
	return b.buildOp(CommandSet, []string{"properties", "title"}, [][]string{{title}})
}

// TODO: Generalize this for the other fields
// UpdatePropertiesOp creates an op to update the block's properties
func (b *Block) UpdatePropertiesOp(source string) *Operation {
	return b.buildOp(CommandUpdate, []string{"properties"}, map[string]interface{}{
		"source": [][]string{{source}},
	})
}

// TODO: Make this work somehow for all of Block's fields
// UpdateOp creates an operation to update the block
func (b *Block) UpdateOp(block *Block) *Operation {
	params := map[string]interface{}{}
	if block.Type != "" {
		params["type"] = block.Type
	}
	if block.LastEditedTime != 0 {
		params["last_edited_time"] = block.LastEditedTime
	}
	if block.LastEditedBy != "" {
		params["last_edited_by"] = block.LastEditedBy
	}
	return b.buildOp(CommandUpdate, []string{}, params)
}

// TODO: Make the input more strict
// UpdateFormatOp creates an operation to update the block's format
func (b *Block) UpdateFormatOp(params interface{}) *Operation {
	return b.buildOp(CommandUpdate, []string{"format"}, params)
}

// ListAfterContentOp creates an operation to list a child block block after another one
// if afterID is empty the block will be listed as the last one
func (b *Block) ListAfterContentOp(id, afterID string) *Operation {
	args := map[string]string{
		"id": id,
	}
	if afterID != "" {
		args["after"] = afterID
	}
	return b.buildOp(CommandListAfter, []string{"content"}, args)
}

// ListRemoveContentOp creates an operation to remove a record from the block
func (b *Block) ListRemoveContentOp(id string) *Operation {
	return b.buildOp(CommandListRemove, []string{"content"}, map[string]string{
		"id": id,
	})
}

// ListAfterFileIDsOp creates an operation to set the file ID
func (b *Block) ListAfterFileIDsOp(fileID string) *Operation {
	return b.buildOp(CommandListAfter, []string{"file_ids"}, map[string]string{
		"id": fileID,
	})
}

/*
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
