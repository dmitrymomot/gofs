package gofs

// DB is the interface for the storage database.
// The database is used to store the status of multipart uploads.
type DB interface {
	// CreateUpload creates a new multipart upload.
	CreateUpload(key string, uploadID string, totalParts int64) error

	// AddPart adds a new part to the multipart upload.
	AddPart(key string, partNumber int64, etag string) error

	// CompleteUpload completes the multipart upload.
	CompleteUpload(key string) error

	// AbortUpload aborts the multipart upload.
	AbortUpload(key string) error

	// GetUploadID returns the upload ID for the given key.
	GetUploadID(key string) (string, error)

	// GetParts returns the parts for the given key.
	GetParts(key string) ([]CompletedPart, error)

	// GetStatus returns the status of the given key.
	GetStatus(key string) (UploadStatus, error)
}

// CompletedPart represents a part of a multipart upload.
type CompletedPart interface {
	PartNumber() int64
	ETag() string
}

// UploadStatus represents the status of a multipart upload.
type UploadStatus interface {
	IsCompleted() bool
	TotalParts() int64
	CompletedPartsNum() int64
}
