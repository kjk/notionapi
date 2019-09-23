package notionapi

const (
	// PermissionTypeUser describes permissions for a user
	PermissionTypeUser = "user_permission"
	// PermissionTypePublic describes permissions for public
	PermissionTypePublic = "public_permission"
)

// for CollectionColumnInfo.Type
const (
	ColumnTypeMultiSelect = "multi_select"
	ColumnTypeCreatedTime = "created_time"
	ColumnTypeNumber      = "number"
	ColumnTypeTitle       = "title"
	ColumnTypeURL         = "url"
	ColumnTypeSelect      = "select"
	ColumnTypeCheckbox    = "checkbox"
	ColumnTypeRelation    = "relation"
	ColumnTypeRollup      = "rollup"
	// TODO: text, date, person, Files&Media, Email, phone
	// formula, time, created by, last edited time, last edited by
)

const (
	// TableSpace represents a Notion workspace
	TableSpace = "space"
	// TableBlock represents a Notion block
	TableBlock = "block"
	// TableUser represents a Notion user
	TableUser = "notion_user"
	// TableCollection represents a Notion collection
	TableCollection = "collection"
)

const (
	// RoleReader represents a reader
	RoleReader = "reader"
	// RoleEditor represents an editor
	RoleEditor = "editor"
)

const (
	// DateTypeDate represents a date in Date.Type
	DateTypeDate = "date"
	// DateTypeDateTime represents a datetime in Date.Type
	DateTypeDateTime = "datetime"
)
