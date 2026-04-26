package dto

// AuthSignUpResponse represents signup response payload.
type AuthSignUpResponse struct {
	ID        string `json:"id"`
	Email     string `json:"email"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
}

// AuthSignInResponse represents signin response payload.
type AuthSignInResponse struct {
	ID string `json:"id"`
}

// AuthRefreshResponse represents token refresh response payload.
type AuthRefreshResponse struct {
	AccessToken string `json:"access_token"`
	ExpiresIn   int    `json:"expires_in"`
}

// MessageResponse represents a simple message response payload.
type MessageResponse struct {
	Message string `json:"message"`
}

// QueueListResponse represents paginated queue list payload.
type QueueListResponse struct {
	Items      []QueueResponse    `json:"items"`
	Pagination PaginationResponse `json:"pagination"`
}

// UserListResponse represents paginated user list payload.
type UserListResponse struct {
	Items      []UserResponse     `json:"items"`
	Pagination PaginationResponse `json:"pagination"`
}

// QueueSignUpResponse represents queue signup response payload.
type QueueSignUpResponse struct {
	QueueID  string `json:"queue_id"`
	Position int    `json:"position"`
	Status   string `json:"status"`
}
