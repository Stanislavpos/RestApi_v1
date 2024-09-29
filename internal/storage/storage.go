package storage

import "errors"

var (
	ErrSongNotFound  = errors.New("song not found")
	ErrSongExist     = errors.New("song exist")
	ErrGroupNotFound = errors.New("group not found")
)
