package rooms

import (
	"errors"
)

var (
	ErrConnNotFound         = errors.New("connection not found")
	ErrDelRoomPermissionDen = errors.New("permission denied; only user that create room can delete them")
	ErrRoomNotFound         = errors.New("room not found")
	ErrRoomIsExist          = errors.New("room is exist")
)
