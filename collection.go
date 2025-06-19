package notionapi

import (
	"errors"
	"fmt"
)

const (
	// TODO: those are probably CollectionViewType
	// CollectionViewTypeTable is a table block
	CollectionViewTypeTable = "table"
	// CollectionViewTypeTable is a lists block
	CollectionViewTypeList = "list"
)

// CollectionColumnOption describes options for ColumnTypeMultiSelect
// collection column
type CollectionColumnOption struct {
	Color string `json:"color"`
	ID    string `json:"id"`
	Value string `json:"value"`
}

type FormulaArg struct {
	Name       *string `json:"name,omitempty"`
	ResultType string  `json:"result_type"`
	Type       string  `json:"type"`
	Value      *string `json:"value,omitempty"`
	ValueType  *string `json:"value_type,omitempty"`
}

type ColumnFormula struct {
	Args       []FormulaArg `json:"args"`
	Name       string       `json:"name"`
	Operator   string       `json:"operator"`
	ResultType string       `json:"result_type"`
	Type       string       `json:"type"`
}

// ColumnSchema describes a info of a collection column
type ColumnSchema struct {
	Name string `json:"name"`
	// ColumnTypeTitle etc.
	Type string `json:"type"`

	// for Type == ColumnTypeNumber, e.g. "dollar", "number"
	NumberFormat string `json:"number_format"`

	// For Type == ColumnTypeRollup
	Aggregation        string `json:"aggregation"` // e.g. "unique"
	TargetProperty     string `json:"target_property"`
	RelationProperty   string `json:"relation_property"`
	TargetPropertyType string `json:"target_property_type"`

	// for Type == ColumnTypeRelation
	CollectionID string `json:"collection_id"`
	Property     string `json:"property"`

	// for Type == ColumnTypeFormula
	Formula *ColumnFormula

	Options []*CollectionColumnOption `json:"options"`

	// TODO: would have to set it up from Collection.RawJSON
	//RawJSON map[string]interface{} `json:"-"`
}

// CollectionPageProperty describes properties of a collection
type CollectionPageProperty struct {
	Property string `json:"property"`
	Visible  bool   `json:"visible"`
}

// CollectionFormat describes format of a collection
type CollectionFormat struct {
	CoverPosition  float64                   `json:"collection_cover_position"`
	PageProperties []*CollectionPageProperty `json:"collection_page_properties"`
}

// Collection describes a collection
type Collection struct {
	ID          string                   `json:"id"`
	SpaceId     *string                  `json:"space_id"`
	Version     int                      `json:"version"`
	Name        interface{}              `json:"name"`
	Schema      map[string]*ColumnSchema `json:"schema"`
	Format      *CollectionFormat        `json:"format"`
	ParentID    string                   `json:"parent_id"`
	ParentTable string                   `json:"parent_table"`
	Alive       bool                     `json:"alive"`
	CopiedFrom  string                   `json:"copied_from"`
	Cover       string                   `json:"cover"`
	Description []interface{}            `json:"description"`

	// TODO: are those ever present?
	Type          string   `json:"type"`
	FileIDs       []string `json:"file_ids"`
	Icon          string   `json:"icon"`
	TemplatePages []string `json:"template_pages"`

	// calculated by us
	name    []*TextSpan
	RawJSON map[string]interface{} `json:"-"`
}

// GetName parses Name and returns as a string
func (c *Collection) GetName() string {
	if len(c.name) == 0 {
		if c.Name == nil {
			return ""
		}
		c.name, _ = ParseTextSpans(c.Name)
	}
	return TextSpansToString(c.name)
}

// TableProperty describes property of a table
type TableProperty struct {
	Width    int    `json:"width"`
	Visible  bool   `json:"visible"`
	Property string `json:"property"`
}

type QuerySort struct {
	ID        string `json:"id"`
	Type      string `json:"type"`
	Property  string `json:"property"`
	Direction string `json:"direction"`
}

type QueryAggregate struct {
	ID              string `json:"id"`
	Type            string `json:"type"`
	Property        string `json:"property"`
	ViewType        string `json:"view_type"`
	AggregationType string `json:"aggregation_type"`
}

type QueryAggregation struct {
	Property   string `json:"property"`
	Aggregator string `json:"aggregator"`
}

type Query struct {
	Sort         []QuerySort            `json:"sort"`
	Aggregate    []QueryAggregate       `json:"aggregate"`
	Aggregations []QueryAggregation     `json:"aggregations"`
	Filter       map[string]interface{} `json:"filter"`
}

// FormatTable describes format for BlockTable
type FormatTable struct {
	PageSort        []string         `json:"page_sort"`
	TableWrap       bool             `json:"table_wrap"`
	TableProperties []*TableProperty `json:"table_properties"`
}

// CollectionView represents a collection view
type CollectionView struct {
	ID          string       `json:"id"`
	Version     int64        `json:"version"`
	Type        string       `json:"type"` // "table"
	Format      *FormatTable `json:"format"`
	Name        string       `json:"name"`
	ParentID    string       `json:"parent_id"`
	ParentTable string       `json:"parent_table"`
	Query       *Query       `json:"query2"`
	Alive       bool         `json:"alive"`
	PageSort    []string     `json:"page_sort"`
	SpaceID     string       `json:"space_id"`

	// set by us
	RawJSON map[string]interface{} `json:"-"`
}

type TableRow struct {
	// TableView that owns this row
	TableView *TableView

	// data for row is stored as properties of a page
	Page *Block

	// values extracted from Page for each column
	Columns [][]*TextSpan
}

// ColumnInfo describes a schema for a given cell (column)
type ColumnInfo struct {
	// TableView that owns this column
	TableView *TableView

	// so that we can access TableRow.Columns[Index]
	Index    int
	Schema   *ColumnSchema
	Property *TableProperty
}

func (c *ColumnInfo) ID() string {
	return c.Property.Property
}

func (c *ColumnInfo) Type() string {
	return c.Schema.Type
}

func (c *ColumnInfo) Name() string {
	if c.Schema == nil {
		return ""
	}
	return c.Schema.Name
}

// TableView represents a view of a table (Notion calls it a Collection View)
// Meant to be a representation that is easier to work with
type TableView struct {
	// original data
	Page           *Page
	CollectionView *CollectionView
	Collection     *Collection

	// easier to work representation we calculate
	Columns []*ColumnInfo
	Rows    []*TableRow

	HasMore  bool
	SizeHint int
	SpaceId  string
}

func (t *TableView) RowCount() int {
	return len(t.Rows)
}

func (t *TableView) ColumnCount() int {
	return len(t.Columns)
}

func (t *TableView) CellContent(row, col int) []*TextSpan {
	return t.Rows[row].Columns[col]
}

// TODO: some tables miss title column in TableProperties
// maybe synthesize it if doesn't exist as a first column
func (c *Client) buildTableView(tv *TableView, res *QueryCollectionResponse) error {
	cv := tv.CollectionView
	collection := tv.Collection

	if collection == nil {
		c.logf("buildTableView: page: '%s', colleciton is nil, collection view id: '%s'\n", ToNoDashID(tv.Page.ID), cv.ID)
		// TODO: maybe should return nil if this is missing in data returned
		// by Notion. If it's a bug in our interpretation, we should fix
		// that instead
		return fmt.Errorf("buildTableView: page: '%s', colleciton is nil, collection view id: '%s'", ToNoDashID(tv.Page.ID), cv.ID)
	}

	if cv.Format != nil && collection.Schema != nil {
		idx := 0
		for _, prop := range cv.Format.TableProperties {
			if !prop.Visible {
				continue
			}
			propName := prop.Property
			schema := collection.Schema[propName]
			ci := &ColumnInfo{
				TableView: tv,

				Index:    idx,
				Property: prop,
				Schema:   schema,
			}
			idx++
			tv.Columns = append(tv.Columns, ci)
		}
	}

	// blockIDs are IDs of page blocks
	// each page represents one table row
	var blockIds []string
	if res.Result.ReducerResults != nil && res.Result.ReducerResults.CollectionGroupResults != nil {
		blockIds = res.Result.ReducerResults.CollectionGroupResults.BlockIds
		tv.HasMore = res.Result.ReducerResults.CollectionGroupResults.HasMore
	}
	for _, id := range blockIds {
		rec, ok := res.RecordMap.Blocks[id]
		if !ok {
			continue
		}
		b := rec.Block
		if b != nil {
			tr := &TableRow{
				TableView: tv,
				Page:      b,
			}
			tv.Rows = append(tv.Rows, tr)
		}
	}

	return nil
}

// FetchTableRows limit is 1000, if limit > 1000, it will get last 1000 rows
func (c *Client) FetchTableRows(tv *TableView, limits ...int) (*TableView, error) {
	if tv == nil {
		return nil, errors.New("tableView is nil")
	}

	req := QueryCollectionRequest{}
	req.Collection.ID = tv.Collection.ID
	req.Collection.SpaceID = tv.SpaceId
	req.CollectionView.ID = tv.CollectionView.ID
	req.CollectionView.SpaceID = tv.SpaceId
	if req.Loader == nil {
		req.Loader = MakeLoaderReducer(tv.CollectionView.Query, limits...)
	}

	var rsp QueryCollectionResponse
	var err error
	apiURL := "/api/v3/queryCollection?src=change_group"
	err = c.doNotionAPI(apiURL, req, &rsp, &rsp.RawJSON)
	if err != nil {
		return nil, err
	}

	if err := ParseRecordMap(rsp.RecordMap); err != nil {
		return nil, err
	}

	tv.Columns = nil // reset columns
	tv.Rows = nil    // reset rows
	tv.SizeHint = rsp.Result.SizeHint

	if err := c.buildTableView(tv, &rsp); err != nil {
		return nil, err
	}

	return tv, nil
}
