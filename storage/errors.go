package storage

import (
	"errors"
)

// Predefined paackage errors
var (
	ErrMissedUploadID     = errors.New("upload id is missed or empty")
	ErrNoCompletedParts   = errors.New("no completed parts, nothing to upload")
	ErrTotalParts         = errors.New("total parts can be between 1 and 10000")
	ErrPartNum            = errors.New("part number can be between 1 and total parts")
	ErrFileEmpty          = errors.New("file is empty")
	ErrInvalidContentType = errors.New("invalid content type")
	ErrInvalidReader      = errors.New("invalid reader provided or reader is nil")
)
