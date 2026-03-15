package errors

import "errors"

var (
	ErrInvalidCredentials  = errors.New("invalid email or password")
	ErrInvalidRefreshToken = errors.New("invalid or expired refresh token")
	ErrInternalServer      = errors.New("internal server error")
)
