package rooms

import (
	"errors"
)

var (
	ErrConnNotFound         = errors.New("connection not found")
	ErrDelRoomPermissionDen = errors.New("permission denied")
	ErrRoomNotFound         = errors.New("rooms not found")
	ErrRoomIsExist          = errors.New("rooms is exist")
)
