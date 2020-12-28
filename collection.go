package notionapi

import (
	"encoding/json"
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
	Version     int                      `json:"version"`
	Name        interface{}              `json:"name"`
	Schema      map[string]*ColumnSchema `json:"schema"`
	Format      *CollectionFormat        `json:"format"`
	ParentID    string                   `json:"parent_id"`
	ParentTable string                   `json:"parent_table"`
	Alive       bool                     `json:"alive"`
	CopiedFrom  string                   `json:"copied_from"`

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

// FormatTable describes format for BlockTable
type FormatTable struct {
	PageSort        []string         `json:"page_sort"`
	TableWrap       bool             `json:"table_wrap"`
	TableProperties []*TableProperty `json:"table_properties"`
}

// CollectionView represents a collection view
type CollectionView struct {
	ID          string          `json:"id"`
	Version     int64           `json:"version"`
	Type        string          `json:"type"` // "table"
	Format      *FormatTable    `json:"format"`
	Name        string          `json:"name"`
	ParentID    string          `json:"parent_id"`
	ParentTable string          `json:"parent_table"`
	Query       json.RawMessage `json:"query"`
	Query2      json.RawMessage `json:"query2"`
	Alive       bool            `json:"alive"`
	PageSort    []string        `json:"page_sort"`
	ShardID     int64           `json:"shard_id"`
	SpaceID     string          `json:"space_id"`

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

	if cv.Format == nil {
		log(c, "buildTableView: page: '%s', missing CollectionView.Format in collection view with id '%s'\n", ToNoDashID(tv.Page.ID), cv.ID)
		return nil
	}

	if collection == nil {
		log(c, "buildTableView: page: '%s', colleciton is nil, collection view id: '%s'\n", ToNoDashID(tv.Page.ID), cv.ID)
		// TODO: maybe should return nil if this is missing in data returned
		// by Notion. If it's a bug in our interpretation, we should fix
		// that instead
		return fmt.Errorf("buildTableView: page: '%s', colleciton is nil, collection view id: '%s'", ToNoDashID(tv.Page.ID), cv.ID)
	}

	if collection.Schema == nil {
		log(c, "buildTableView: page: '%s', missing collection.Schema, collection view id: '%s', collection id: '%s'\n", ToNoDashID(tv.Page.ID), cv.ID, collection.ID)
		// TODO: maybe should return nil if this is missing in data returned
		// by Notion. If it's a bug in our interpretation, we should fix
		// that instead
		return fmt.Errorf("buildTableView: page: '%s', missing collection.Schema, collection view id: '%s', collection id: '%s'", ToNoDashID(tv.Page.ID), cv.ID, collection.ID)
	}

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

	// blockIDs are IDs of page blocks
	// each page represents one table row
	blockIds := res.Result.BlockIDS
	for _, id := range blockIds {
		rec, ok := res.RecordMap.Blocks[id]
		if !ok {
			cvID := tv.CollectionView.ID
			return fmt.Errorf("didn't find block with id '%s' for collection view with id '%s'", id, cvID)
		}
		b := rec.Block
		tr := &TableRow{
			TableView: tv,
			Page:      b,
		}
		tv.Rows = append(tv.Rows, tr)
	}

	// pre-calculate cell content
	for _, tr := range tv.Rows {
		for _, ci := range tv.Columns {
			propName := ci.Property.Property
			v := tr.Page.GetProperty(propName)
			tr.Columns = append(tr.Columns, v)
		}
	}
	return nil
}
