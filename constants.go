package notionapi

const (
	// PermissionTypeUser describes permissions for a user
	PermissionTypeUser = "user_permission"
	// PermissionTypePublic describes permissions for public
	PermissionTypePublic = "public_permission"
)

// for CollectionColumnInfo.Type
const (
	ColumnTypeCheckbox       = "checkbox"
	ColumnTypeCreatedBy      = "created_by"
	ColumnTypeCreatedTime    = "created_time"
	ColumnTypeDate           = "date"
	ColumnTypeEmail          = "email"
	ColumnTypeFile           = "file"
	ColumnTypeForumula       = "formula"
	ColumnTypeLastEditedBy   = "last_edited_by"
	ColumnTypeLastEditedTime = "last_edited_time"
	ColumnTypeMultiSelect    = "multi_select"
	ColumnTypeNumber         = "number"
	ColumnTypePerson         = "person"
	ColumnTypePhoneNumber    = "phone_number"
	ColumnTypeRelation       = "relation"
	ColumnTypeRollup         = "rollup"
	ColumnTypeSelect         = "select"
	ColumnTypeText           = "text"
	ColumnTypeTitle          = "title"
	ColumnTypeURL            = "url"
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
