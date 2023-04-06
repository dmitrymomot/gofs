package storage

import (
	"io"
	"strings"

	"github.com/gabriel-vasile/mimetype"
	"github.com/pkg/errors"
)

// GetFileContentType returns the content type of a file.
func GetFileContentType(input io.Reader) (string, error) {
	if input == nil {
		return "", errors.Wrap(ErrInvalidReader, "storage.GetFileContentType")
	}

	mtype, err := mimetype.DetectReader(input)
	if err != nil {
		return "", errors.Wrap(err, "storage.GetFileContentType")
	}

	parts := strings.Split(mtype.String(), ";")
	return parts[0], nil
}

// GetFileContentTypeByBytes returns the content type of a file.
func GetFileContentTypeByBytes(input []byte) (string, error) {
	if input == nil {
		return "", errors.Wrap(ErrInvalidReader, "storage.GetFileContentTypeByBytes")
	}

	mtype := mimetype.Detect(input)
	parts := strings.Split(mtype.String(), ";")
	return parts[0], nil
}

// Get max file parts can be if the file is split into parts with the given part size.
// The max file parts is 10000.
// file io.ReadSeeker: the file to be uploaded.
// partSize int64: the part size.
func GetMaxFileParts(file io.ReadSeeker, partSize int64) (int64, error) {
	if file == nil {
		return 0, errors.Wrap(ErrInvalidReader, "storage.GetMaxFileParts")
	}

	// Get the file size.
	fileSize, err := GetFileSize(file)
	if err != nil {
		return 0, errors.Wrap(err, "storage.GetMaxFileParts")
	}

	// Get the max file parts.
	maxFileParts := int64(fileSize / partSize)
	if fileSize%partSize != 0 {
		maxFileParts++
	}

	return maxFileParts, nil
}

// Get the file size.
// file io.ReadSeeker: the file to be uploaded.
func GetFileSize(file io.ReadSeeker) (int64, error) {
	if file == nil {
		return 0, errors.Wrap(ErrInvalidReader, "storage.GetFileSize")
	}

	// Get the file size.
	fileSize, err := file.Seek(0, io.SeekEnd)
	if err != nil {
		return 0, errors.Wrap(err, "storage.GetFileSize")
	}

	// Reset the read position to the beginning.
	if _, err := file.Seek(0, io.SeekStart); err != nil {
		return 0, errors.Wrap(err, "storage.GetFileSize")
	}

	return fileSize, nil
}

// Get the file extension.
// fileName string: the file name.
func GetFileExtension(fileName string) string {
	if fileName == "" {
		return ""
	}

	// Get the file extension.
	parts := strings.Split(fileName, ".")
	if len(parts) == 0 {
		return ""
	}

	return parts[len(parts)-1]
}

// Get the file name without extension.
// fileName string: the file name.
func GetFileNameWithoutExtension(fileName string) string {
	if fileName == "" {
		return ""
	}

	// Get the file extension.
	parts := strings.Split(fileName, ".")
	if len(parts) == 0 {
		return ""
	}

	return strings.Join(parts[:len(parts)-1], ".")
}
