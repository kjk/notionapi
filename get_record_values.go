package notionapi

import (
	"encoding/json"
	"fmt"
)

// /api/v3/getRecordValues request
type getRecordValuesRequest struct {
	Requests []RecordRequest `json:"requests"`
}

// RecordRequest represents argument to GetRecordValues
type RecordRequest struct {
	Table string `json:"table"`
	ID    string `json:"id"`
}

// Record represents a polymorphic record
type Record struct {
	// fields returned by the server
	Role string `json:"role"`
	// polymorphic value of the record, which we decode into Block, Space etc.
	Value json.RawMessage `json:"value"`

	// fields set from Value based on type
	ID    string `json:"-"`
	Table string `json:"-"`

	Activity       *Activity       `json:"-"`
	Block          *Block          `json:"-"`
	Space          *Space          `json:"-"`
	User           *User           `json:"-"`
	Collection     *Collection     `json:"-"`
	CollectionView *CollectionView `json:"-"`
	Comment        *Comment        `json:"-"`
	Discussion     *Discussion     `json:"-"`
	// TODO: add more types
}

// GetRecordValuesResponse represents response to /api/v3/getRecordValues api
// Note: it depends on Table type in request
type GetRecordValuesResponse struct {
	Results []*Record              `json:"results"`
	RawJSON map[string]interface{} `json:"-"`
}

// table is not always present in Record returned by the server
// so must be provided based on what was asked
func parseRecord(table string, r *Record) error {
	// it's ok if some records don't return a value
	if len(r.Value) == 0 {
		return nil
	}
	if r.Table == "" {
		r.Table = table
	} else {
		// TODO: probably never happens
		panicIf(r.Table != table)
	}

	// set Block/Space etc. based on TableView type
	var pRawJSON *map[string]interface{}
	var obj interface{}
	switch table {
	case TableActivity:
		r.Activity = &Activity{}
		obj = r.Activity
		pRawJSON = &r.Activity.RawJSON
	case TableBlock:
		r.Block = &Block{}
		obj = r.Block
		pRawJSON = &r.Block.RawJSON
	case TableUser:
		r.User = &User{}
		obj = r.User
		pRawJSON = &r.User.RawJSON
	case TableSpace:
		r.Space = &Space{}
		obj = r.Space
		pRawJSON = &r.Space.RawJSON
	case TableCollection:
		r.Collection = &Collection{}
		obj = r.Collection
		pRawJSON = &r.Collection.RawJSON
	case TableCollectionView:
		r.CollectionView = &CollectionView{}
		obj = r.CollectionView
		pRawJSON = &r.CollectionView.RawJSON
	case TableDiscussion:
		r.Discussion = &Discussion{}
		obj = r.Discussion
		pRawJSON = &r.Discussion.RawJSON
	case TableComment:
		r.Comment = &Comment{}
		obj = r.Comment
		pRawJSON = &r.Comment.RawJSON
	}
	if obj == nil {
		return fmt.Errorf("unsupported table '%s'", r.Table)
	}
	if false {
		if table == TableCollectionView {
			s := string(r.Value)
			fmt.Printf("collection_view json:\n%s\n\n", s)
		}
	}
	if err := json.Unmarshal(r.Value, pRawJSON); err != nil {
		return err
	}
	id := (*pRawJSON)["id"]
	if id != nil {
		r.ID = id.(string)
	}
	if err := json.Unmarshal(r.Value, &obj); err != nil {
		return err
	}
	return nil
}

// GetBlockRecords executes a raw API call /api/v3/getRecordValues
// to get records for blocks with given ids
func (c *Client) GetBlockRecords(ids []string) (*GetRecordValuesResponse, error) {
	records := make([]RecordRequest, len(ids))
	for pos, id := range ids {
		dashID := ToDashID(id)
		if !IsValidDashID(dashID) {
			return nil, fmt.Errorf("'%s' is not a valid notion id", id)
		}
		records[pos].Table = TableBlock
		records[pos].ID = dashID
	}
	return c.GetRecordValues(records)
}

// GetRecordValues executes a raw API call /api/v3/getRecordValues
func (c *Client) GetRecordValues(records []RecordRequest) (*GetRecordValuesResponse, error) {
	req := &getRecordValuesRequest{
		Requests: records,
	}

	apiURL := "/api/v3/getRecordValues"
	var rsp GetRecordValuesResponse
	var err error
	if rsp.RawJSON, err = doNotionAPI(c, apiURL, req, &rsp); err != nil {
		return nil, err
	}

	for idx, r := range rsp.Results {
		table := records[idx].Table
		err = parseRecord(table, r)
		if err != nil {
			return nil, err
		}
	}

	return &rsp, nil
}
