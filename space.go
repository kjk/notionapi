package notionapi

// SpacePermissions represents permissions for space
type SpacePermissions struct {
	Role   string `json:"role"`
	Type   string `json:"type"` // e.g. "user_permission"
	UserID string `json:"user_id"`
}

// SpacePermissionGroups represesnts group permissions for space
type SpacePermissionGroups struct {
	ID      string   `json:"id"`
	Name    string   `json:"name"`
	UserIds []string `json:"user_ids,omitempty"`
}

// Space describes Notion workspace.
type Space struct {
	ID                  string                  `json:"id"`
	Version             float64                 `json:"version"`
	Name                string                  `json:"name"`
	Domain              string                  `json:"domain"`
	Permissions         []*SpacePermissions     `json:"permissions,omitempty"`
	PermissionGroups    []SpacePermissionGroups `json:"permission_groups"`
	Icon                string                  `json:"icon"`
	EmailDomains        []string                `json:"email_domains"`
	BetaEnabled         bool                    `json:"beta_enabled"`
	Pages               []string                `json:"pages,omitempty"`
	DisablePublicAccess bool                    `json:"disable_public_access"`
	DisableGuests       bool                    `json:"disable_guests"`
	DisableMoveToSpace  bool                    `json:"disable_move_to_space"`
	DisableExport       bool                    `json:"disable_export"`
	CreatedBy           string                  `json:"created_by"`
	CreatedTime         int64                   `json:"created_time"`
	LastEditedBy        string                  `json:"last_edited_by"`
	LastEditedTime      int64                   `json:"last_edited_time"`

	RawJSON map[string]interface{} `json:"-"`
}
