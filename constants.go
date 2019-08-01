package notionapi

const (
	// PermissionTypeUser describes permissions for a user
	PermissionTypeUser = "user_permission"
	// PermissionTypePublic describes permissions for public
	PermissionTypePublic = "public_permission"
)

// for CollectionColumnInfo.Type
const (
	// ColumnMultiSelect is multi-select column
	ColumnMultiSelect = "multi_select"
	ColumnTypeNumber  = "number"
	ColumnTypeTitle   = "title"
	// TODO: text, select, date, person, Files&Media, checkbox, URL, Email, phone
	// formula, relaion, created time, created by, last edited time, last edited by
)

const (
	// TableSpace represents a Notion workspace
	TableSpace = "space"
	// TableBlock represents a Notion block
	TableBlock = "block"
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
