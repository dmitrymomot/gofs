package storage_test

import (
	"io"
	"os"
	"strings"
	"testing"

	"github.com/dmitrymomot/go-env"
	"github.com/dmitrymomot/gofs/storage"
	"github.com/google/uuid"
	_ "github.com/joho/godotenv/autoload" // Load .env file automatically
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// S3 configuration from .env file
var (
	fileStorageKey            = env.MustString("STORAGE_KEY")
	fileStorageSecret         = env.MustString("STORAGE_SECRET")
	fileStorageEndpoint       = env.MustString("STORAGE_ENDPOINT")
	fileStorageRegion         = env.MustString("STORAGE_REGION")
	fileStorageBucket         = env.MustString("STORAGE_BUCKET")
	fileStorageUrl            = env.MustString("STORAGE_URL")
	fileStorageDisableSsl     = env.GetBool("STORAGE_DISABLE_SSL", false)
	fileStorageForcePathStyle = env.GetBool("STORAGE_FORCE_PATH_STYLE", false)

	// S3 client instnce for tests.
	// It connects to S3-compatible storage.
	s3Client, _ = storage.NewS3Client(storage.Options{
		Key:            fileStorageKey,
		Secret:         fileStorageSecret,
		Endpoint:       fileStorageEndpoint,
		Region:         fileStorageRegion,
		DisableSSL:     fileStorageDisableSsl,
		ForcePathStyle: fileStorageForcePathStyle,
	})

	// Storage interactor instance for tests.
	interactor = storage.New(s3Client, fileStorageBucket, fileStorageUrl)
)

// create test file testdata/text.txt with content "Hello, World!".
func createTestFile() {
	// check if test file exists
	if _, err := os.Stat("testdata/text.txt"); !os.IsNotExist(err) {
		// exit if test file exists
		return
	}

	// check if testdata directory exists
	if _, err := os.Stat("testdata"); os.IsNotExist(err) {
		// create testdata directory
		if err := os.Mkdir("testdata", 0o755); err != nil {
			panic(err)
		}
	}

	// create test file
	file, err := os.Create("testdata/text.txt")
	if err != nil {
		panic(err)
	}
	defer file.Close()

	// write "Hello, World!" to test file
	if _, err := file.WriteString("Hello, World!"); err != nil {
		panic(err)
	}
}

// Test upload and download file to S3-compatible storage.
func TestSimpleFileInteraction(t *testing.T) {
	createTestFile()

	file, err := os.Open("testdata/text.txt")
	require.NoError(t, err)
	defer file.Close()

	fileBytes, err := io.ReadAll(file)
	require.NoError(t, err)
	require.Greater(t, len(fileBytes), 0)

	filepath := strings.Join([]string{
		"testing",
		uuid.New().String(),
		"text.txt",
	}, "/")

	t.Run("Upload", func(t *testing.T) {
		contentType, err := storage.GetFileContentType(file)
		require.NoError(t, err)
		assert.Equal(t, "text/plain", contentType)

		// Upload file to storage.
		require.NoError(t, interactor.Upload(fileBytes, filepath, storage.Public, contentType))
	})

	t.Run("Download", func(t *testing.T) {
		// Download file from storage.
		fl, ct, err := interactor.Download(filepath)
		require.NoError(t, err)
		assert.Equal(t, "text/plain", *ct)
		assert.NotNil(t, fl)
		defer fl.Close()

		data, err := io.ReadAll(fl)
		require.NoError(t, err)
		require.EqualValues(t, "Hello, World!", string(data))
	})

	t.Run("Delete", func(t *testing.T) {
		// Delete file from storage.
		require.NoError(t, interactor.Delete(filepath))

		// Check if file was deleted
		_, _, err := interactor.Download(filepath)
		assert.Error(t, err)
	})
}

// Test multipart upload file to S3-compatible storage.
func TestMultipartUpload(t *testing.T) {
	file, err := os.Open("testdata/image.png")
	require.NoError(t, err)
	defer file.Close()

	fileBytes, err := io.ReadAll(file)
	require.NoError(t, err)
	require.Greater(t, len(fileBytes), 0)

	filepath := strings.Join([]string{
		"testing",
		uuid.New().String(),
		"image.png",
	}, "/")

	t.Run("Upload", func(t *testing.T) {
		contentType, err := storage.GetFileContentTypeByBytes(fileBytes)
		require.NoError(t, err)
		require.Equal(t, "image/png", contentType)

		// Create multipart upload
		uploadID, err := interactor.CreateMultipartUpload(filepath, contentType, storage.Public)
		require.NoError(t, err)
		require.NotEmpty(t, uploadID)
		require.NotNil(t, uploadID)

		maxPartSize := int64(1024 * 1024 * 5) // 5MB
		totalParts, err := storage.GetMaxFileParts(file, maxPartSize)
		require.NoError(t, err)
		require.Greater(t, totalParts, int64(0))
		require.LessOrEqual(t, totalParts, int64(10000))

		var start, currentSize int64
		remaining := len(fileBytes)
		partNum := int64(1)
		completedParts := make([]storage.CompletedPart, 0, totalParts)

		for start = 0; remaining > 0; start += maxPartSize {
			if remaining > int(maxPartSize) {
				currentSize = maxPartSize
			} else {
				currentSize = int64(remaining)
			}

			// Upload file parts
			part, err := interactor.UploadPart(filepath, uploadID, fileBytes[start:start+currentSize], partNum, totalParts)
			require.NoError(t, err)
			require.NotEmpty(t, part)
			require.NotNil(t, part)

			completedParts = append(completedParts, part)

			remaining -= int(currentSize)
			partNum++
		}

		// Complete multipart upload
		require.NoError(t, interactor.CompleteMultipartUpload(filepath, uploadID, completedParts...))
	})

	t.Run("Download", func(t *testing.T) {
		// Download file from storage.
		fl, ct, err := interactor.Download(filepath)
		require.NoError(t, err)
		assert.Equal(t, "image/png", *ct)
		assert.NotNil(t, fl)
		defer fl.Close()

		data, err := io.ReadAll(fl)
		require.NoError(t, err)
		require.EqualValues(t, fileBytes, data)
	})

	t.Run("Delete", func(t *testing.T) {
		// Delete file from storage.
		require.NoError(t, interactor.Delete(filepath))

		// Check if file was deleted
		_, _, err := interactor.Download(filepath)
		assert.Error(t, err)
	})
}
