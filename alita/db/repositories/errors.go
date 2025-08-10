package repositories

import "errors"

// Repository validation errors
var (
	// User repository errors
	ErrUserNil           = errors.New("USER_NIL")
	ErrUserIDZero        = errors.New("USER_ID_ZERO")
	ErrUsernameEmpty     = errors.New("USERNAME_EMPTY")
	ErrUserNotFound      = errors.New("USER_NOT_FOUND")
	ErrUserAtIndexNil    = errors.New("USER_AT_INDEX_NIL")
	ErrUserAtIndexZeroID = errors.New("USER_AT_INDEX_ZERO_ID")

	// Chat repository errors
	ErrChatNil      = errors.New("CHAT_NIL")
	ErrChatIDZero   = errors.New("CHAT_ID_ZERO")
	ErrChatNotFound = errors.New("CHAT_NOT_FOUND")
)
