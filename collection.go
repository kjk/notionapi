package notionapi

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

// ColumnSchema describes a info of a collection column
type ColumnSchema struct {
	Name string `json:"name"`
	// ColumnTypeTitle etc.
	Type string `json:"type"`

	// for Type == ColumnTypeNumber, e.g. "dollar", "number"
	NumberFormat string `json:"number_format"`

	// For Type == ColumnTypeRollup
	TargetProperty     string `json:"target_property"`
	RelationProperty   string `json:"relation_property"`
	TargetPropertyType string `json:"target_property_type"`

	// for Type == ColumnTypeRelation
	CollectionID string `json:"collection_id"`
	Property     string `json:"property"`

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
	ID          string       `json:"id"`
	Version     int64        `json:"version"`
	Type        string       `json:"type"` // "table"
	Format      *FormatTable `json:"format"`
	Name        string       `json:"name"`
	ParentID    string       `json:"parent_id"`
	ParentTable string       `json:"parent_table"`
	Query       *Query       `json:"query"`
	Alive       bool         `json:"alive"`
	PageSort    []string     `json:"page_sort"`

	// set by us
	RawJSON map[string]interface{} `json:"-"`
}

// CellSchema describes a schema for a given cell (column)
type CellSchema struct {
	// TODO: implement me
}

// TableColumn represents a single cell in a table
type TableCell struct {
	Parent *TableRow

	Value  []*TextSpan
	Schema *CellSchema
}

type TableRow struct {
	// data for row is stored as properties of a page
	Page *Block

	Columns []*TableCell
}

// TableView represents a view of a table (Notion calls it a Collection View)
// Meant to be a representation that is easier to work with
type TableView struct {
	// this is the raw data from which we build a representation
	// that is nicer to work with
	Page           *Page
	CollectionView *CollectionView
	Collection     *Collection

	// a table is an array of rows
	ColumnHeaders []*TableProperty
	Rows          []*TableRow

	// TODO: temporary
	OriginatingBlock *Block
	CollectionRows   []*Block
}

func (t *TableView) RowCount() int {
	return len(t.Rows)
}

func (t *TableView) ColumnCount() int {
	if len(t.Rows) == 0 {
		return 0
	}
	// we assume each row has the same amount of columns
	return len(t.Rows[0].Columns)
}

func buildTableView(tv *TableView) {
	var cols []*TableProperty
	cv := tv.CollectionView
	//c := tv.Collection
	props := cv.Format.TableProperties
	for _, prop := range props {
		if prop.Visible {
			cols = append(cols, prop)
		}
	}
	tv.ColumnHeaders = cols
	/*
		for _, rowID := range cv.PageSort {

		}*/
}
