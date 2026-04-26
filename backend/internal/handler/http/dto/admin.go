package dto

// AdminUpdateUserRequest represents admin request for partial user update.
type AdminUpdateUserRequest struct {
	Role      *string `json:"role"`
	IsBlocked *bool   `json:"is_blocked"`
	GroupID   *string `json:"group_id"`
}

// AdminUpsertGroupRequest represents admin request for create/update group.
type AdminUpsertGroupRequest struct {
	Name string `json:"name" binding:"required"`
}
