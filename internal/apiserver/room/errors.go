package room

import (
	"errors"
)

var (
	ErrRoomIsExist = errors.New("room is exist")
)

var (
	ErrUserNotFound = errors.New("user not found")
	ErrRoomNotFound = errors.New("room not found")
	ErrDelRoomPermissionDen = errors.New("permission denied")
)
