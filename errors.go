package gofs

import "errors"

// Predefined errors.
var (
	ErrNotFound          = errors.New("not found")
	ErrAlreadyExists     = errors.New("already exists")
	ErrFileKeyEmpty      = errors.New("file uploading key cannot be empty")
	ErrInvalidTotalParts = errors.New("total parts must be greater than zero and not more than 10000")
)
