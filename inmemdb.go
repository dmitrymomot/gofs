package gofs

import "sync"

type (
	// InMemoryDB is an in-memory implementation of the DB interface.
	inMemoryDB struct {
		sync.RWMutex
		records map[string]inMemoryRecord
	}

	// Represents a record in the database. It has an uploadID string field, an integer totalParts field indicating how many parts the record is split into, and a map from part numbers to inMemoryPart values (parts).
	inMemoryRecord struct {
		uploadID   string
		totalParts int64
		parts      map[int64]inMemoryPart
	}

	// Represents a single part of a larger record. It has a partNumber integer field and an eTag string field.
	inMemoryPart struct {
		partNumber int64
		eTag       string
	}
)

// NewInMemoryDB creates a new in-memory database.
func NewInMemoryDB() DB {
	return &inMemoryDB{
		records: make(map[string]inMemoryRecord),
	}
}

// CreateUpload creates a new upload with the given key (string), uploadID (string) and totalParts (int64).
func (db *inMemoryDB) CreateUpload(key string, uploadID string, totalParts int64) error {
	// Validate inputs
	if key == "" {
		return ErrFileKeyEmpty
	}
	if totalParts <= 0 || totalParts > 10000 {
		return ErrInvalidTotalParts
	}

	db.Lock()
	defer db.Unlock()

	// Check if record already exists for the given key
	if _, ok := db.records[key]; ok {
		return ErrAlreadyExists // Return ErrAlreadyExists indicating that a record already exists for the given key
	}

	// Create a new inMemoryRecord object with the given uploadID and totalParts, and initialize its parts map
	record := inMemoryRecord{
		uploadID:   uploadID,
		totalParts: totalParts,
		parts:      make(map[int64]inMemoryPart),
	}

	db.records[key] = record // Add the new record to the database

	return nil // Return nil error indicating that the operation was successful
}

// AddPart is a method of the inMemoryDB struct that takes in a key (string), a partNumber (int64) and an eTag (string)
// and returns an error.
func (db *inMemoryDB) AddPart(key string, partNumber int64, eTag string) error {
	db.Lock()
	defer db.Unlock()

	record, ok := db.records[key]
	if !ok {
		return ErrNotFound
	}

	record.parts[partNumber] = inMemoryPart{
		partNumber: partNumber,
		eTag:       eTag,
	}

	db.records[key] = record

	return nil
}

// CompleteUpload is a method of the struct inMemoryDB that takes in a key (string) and completes the corresponding upload.
// It returns an error if the operation was unsuccessful, specifically if the key was not found in the records.
func (db *inMemoryDB) CompleteUpload(key string) error {
	// acquire a write lock on the database to protect against concurrent access
	db.Lock()
	defer db.Unlock()

	// check if the given key exists in the records map
	if _, ok := db.records[key]; !ok {
		// return an error indicating that the record was not found
		return ErrNotFound
	}

	// remove the record associated with the given key from the records map
	delete(db.records, key)

	// return a nil error indicating that the operation was successful
	return nil
}

// AbortUpload is a method of the inMemoryDB struct that takes in a key (string) and aborts the corresponding upload.
// It returns an error if the operation was unsuccessful.
func (db *inMemoryDB) AbortUpload(key string) error {
	// acquire a write lock on the database to protect against concurrent access
	db.Lock()
	defer db.Unlock()

	// remove the record associated with the given key from the database
	delete(db.records, key)

	// return a nil error indicating that the operation was successful
	return nil
}

// GetUploadID is a method of the inMemoryDB struct that takes in a key (string)
// and returns the corresponding upload ID (string) and an error.
func (db *inMemoryDB) GetUploadID(key string) (string, error) {
	// acquire a read lock on the database to protect against concurrent access
	db.RLock()
	defer db.RUnlock()

	record, ok := db.records[key]
	if !ok {
		return "", ErrNotFound
	}

	return record.uploadID, nil
}

// GetParts is a method of the inMemoryDB struct that takes in a key (string)
// and returns a slice of CompletedPart interface and an error.
func (db *inMemoryDB) GetParts(key string) ([]CompletedPart, error) {
	db.RLock()
	defer db.RUnlock()

	record, ok := db.records[key]
	if !ok {
		return nil, ErrNotFound
	}

	parts := make([]CompletedPart, 0, len(record.parts))
	for _, part := range record.parts {
		parts = append(parts, part)
	}

	return parts, nil
}

// GetStatus returns the status of the upload.
func (db *inMemoryDB) GetStatus(key string) (UploadStatus, error) {
	db.RLock()
	defer db.RUnlock()

	record, ok := db.records[key]
	if !ok {
		return nil, ErrNotFound
	}

	return record, nil
}

// PartNumber returns the part number.
// Part numbers start at 1.
func (part inMemoryPart) PartNumber() int64 {
	return part.partNumber
}

// ETag returns the ETag of the part.
// The ETag is the MD5 hash of the part.
func (part inMemoryPart) ETag() string {
	return part.eTag
}

// IsCompleted returns true if the upload is completed.
func (record inMemoryRecord) IsCompleted() bool {
	return int64(len(record.parts)) == record.totalParts
}

// TotalParts returns the total number of parts in the upload.
func (record inMemoryRecord) TotalParts() int64 {
	return record.totalParts
}

// Parts returns the parts that have been uploaded.
func (record inMemoryRecord) CompletedParts() []CompletedPart {
	parts := make([]CompletedPart, 0, len(record.parts))
	for _, part := range record.parts {
		parts = append(parts, part)
	}

	return parts
}

// UploadID returns the upload ID.
func (record inMemoryRecord) UploadID() string {
	return record.uploadID
}

// CompletedPartsNum returns the number of completed parts.
func (record inMemoryRecord) CompletedPartsNum() int64 {
	return int64(len(record.parts))
}
