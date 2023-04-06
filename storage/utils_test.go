package storage_test

import (
	"bytes"
	"os"
	"testing"

	"github.com/dmitrymomot/gofs/storage"
	"github.com/stretchr/testify/assert"
)

func TestGetFileContentType(t *testing.T) {
	t.Run("Test Case 1 - Invalid Reader", func(t *testing.T) {
		contentType, err := storage.GetFileContentType(nil)
		assert.Error(t, err)
		assert.Equal(t, "", contentType)
	})

	t.Run("Test Case 2 - from string", func(t *testing.T) {
		buffer := []byte("this is a sample text and not a file")
		contentType, err := storage.GetFileContentType(bytes.NewReader(buffer))
		assert.NoError(t, err)
		assert.Equal(t, "text/plain", contentType)
	})

	t.Run("Test Case 3 - Valid file content type", func(t *testing.T) {
		file, err := os.Open("testdata/image.png")
		assert.NoError(t, err)
		defer file.Close()

		contentType, err := storage.GetFileContentType(file)
		assert.NoError(t, err)
		assert.Equal(t, "image/png", contentType)
	})
}

func TestGetMaxFileParts(t *testing.T) {
	t.Run("Test Case 1 - Invalid Reader", func(t *testing.T) {
		maxParts, err := storage.GetMaxFileParts(nil, 5*1024*1024)
		assert.Error(t, err)
		assert.Equal(t, 0, maxParts)
	})

	t.Run("Test Case 2 - Valid file content type", func(t *testing.T) {
		file, err := os.Open("testdata/image.png")
		assert.NoError(t, err)
		defer file.Close()

		maxParts, err := storage.GetMaxFileParts(file, 1024*1024)
		assert.NoError(t, err)
		assert.Equal(t, 1, maxParts)

		maxParts, err = storage.GetMaxFileParts(file, 1024)
		assert.NoError(t, err)
		assert.Equal(t, 6, maxParts)
	})
}
