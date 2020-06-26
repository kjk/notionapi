package notionapi

const (
	// PermissionTypeUser describes permissions for a user
	PermissionTypeUser = "user_permission"
	// PermissionTypePublic describes permissions for public
	PermissionTypePublic = "public_permission"
)

// for Schema.Type
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
	// those are Record.Type and determine the type of Record.Value
	TableSpace          = "space"
	TableActivity       = "activity"
	TableBlock          = "block"
	TableUser           = "notion_user"
	TableCollection     = "collection"
	TableCollectionView = "collection_view"
	TableComment        = "comment"
	TableDiscussion     = "discussion"
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
