package storage

import (
	"bytes"
	"fmt"
	"io"
	"sort"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/pkg/errors"
)

// Interactor struct
type Interactor struct {
	s3             *s3.S3
	bucket         string
	url            string
	forcePathStyle bool
}

// New is a factory function,
// returns a new instance of the storage interactor
func New(s3Client *s3.S3, bucket, fileEndpoint string) *Interactor {
	return &Interactor{
		s3:             s3Client,
		bucket:         bucket,
		url:            fileEndpoint,
		forcePathStyle: *s3Client.Config.S3ForcePathStyle,
	}
}

// Upload file to the cloud storage
func (i *Interactor) Upload(file []byte, filepath string, acl ACL, contentType string) error {
	input := s3.PutObjectInput{
		Bucket:      aws.String(i.bucket),
		Key:         aws.String(filepath),
		Body:        bytes.NewReader(file),
		ACL:         aws.String(acl.String()),
		ContentType: aws.String(contentType),
	}
	if err := input.Validate(); err != nil {
		return errors.Wrap(err, "storage.upload")
	}

	if _, err := i.s3.PutObject(&input); err != nil {
		return errors.Wrap(err, "storage.upload")
	}

	return nil
}

// Download file from the cloud storage
func (i *Interactor) Download(filepath string) (io.ReadCloser, *string, error) {
	input := &s3.GetObjectInput{
		Bucket: aws.String(i.bucket),
		Key:    aws.String(filepath),
	}
	if err := input.Validate(); err != nil {
		return nil, nil, errors.Wrap(err, "storage.download")
	}

	result, err := i.s3.GetObject(input)
	if err != nil {
		return nil, nil, errors.Wrap(err, "storage.download")
	}

	return result.Body, result.ContentType, nil
}

// Remove file from the cloud storage
func (i *Interactor) Remove(filepath string) error {
	input := &s3.DeleteObjectInput{
		Bucket: aws.String(i.bucket),
		Key:    aws.String(filepath),
	}
	if err := input.Validate(); err != nil {
		return errors.Wrap(err, "storage.remove")
	}

	if _, err := i.s3.DeleteObject(input); err != nil {
		return errors.Wrap(err, "storage.remove")
	}

	return nil
}

// FileURL return public url for a file
func (i *Interactor) FileURL(filepath string) string {
	if i.forcePathStyle {
		return fmt.Sprintf("%s/%s/%s", i.url, i.bucket, filepath)
	}

	return fmt.Sprintf("%s/%s", i.url, filepath)
}

// Create multipart upload
func (i *Interactor) CreateMultipartUpload(filename, contentType string, acl ACL) (string, error) {
	input := &s3.CreateMultipartUploadInput{
		ACL:         aws.String(acl.String()),
		Bucket:      aws.String(i.bucket),
		Key:         aws.String(filename),
		ContentType: aws.String(contentType),
	}
	if err := input.Validate(); err != nil {
		return "", errors.Wrap(err, "storage.createMultipartUpload: invalid params")
	}

	result, err := i.s3.CreateMultipartUpload(input)
	if err != nil {
		return "", errors.Wrap(err, "storage.createMultipartUpload")
	}
	if result.UploadId == nil {
		return "", ErrMissedUploadID
	}

	return *result.UploadId, nil
}

// AbortMultipartUpload aborts a multipart upload.
func (i *Interactor) AbortMultipartUpload(filename, uploadID string) error {
	if uploadID == "" {
		return ErrMissedUploadID
	}

	params := &s3.AbortMultipartUploadInput{
		Bucket:   aws.String(i.bucket),
		Key:      aws.String(filename),
		UploadId: aws.String(uploadID),
	}
	if err := params.Validate(); err != nil {
		return errors.Wrap(err, "storage.abortMultipartUpload: invalid params")
	}

	if _, err := i.s3.AbortMultipartUpload(params); err != nil {
		return errors.Wrap(err, "storage.abortMultipartUpload")
	}

	return nil
}

// CompleteMultipartUpload completes a multipart upload.
func (i *Interactor) CompleteMultipartUpload(filename, uploadID string, completedParts ...*s3.CompletedPart) error {
	if uploadID == "" {
		return ErrMissedUploadID
	}
	if len(completedParts) == 0 {
		return ErrNoCompletedParts
	}

	// Ordering the array based on the PartNumber as each parts could be uploaded in different order!
	sort.Slice(completedParts, func(i, j int) bool {
		return *completedParts[i].PartNumber < *completedParts[j].PartNumber
	})

	params := &s3.CompleteMultipartUploadInput{
		Bucket:          aws.String(i.bucket),
		Key:             aws.String(filename),
		UploadId:        aws.String(uploadID),
		MultipartUpload: &s3.CompletedMultipartUpload{Parts: completedParts},
	}
	if err := params.Validate(); err != nil {
		return errors.Wrap(err, "storage.completeMultipartUpload: invalid params")
	}

	if _, err := i.s3.CompleteMultipartUpload(params); err != nil {
		return errors.Wrap(err, "storage.completeMultipartUpload")
	}

	return nil
}

// Upload uploads a file to S3.
// If partNum is equal to totalParts, the file is considered complete and the
// multipart upload is completed.
func (i *Interactor) UploadPart(filename, uploadID string, data []byte, partNum, totalParts int64) (*s3.CompletedPart, error) {
	if uploadID == "" {
		return nil, ErrMissedUploadID
	}

	if totalParts == 0 || totalParts > 10000 {
		return nil, ErrTotalParts
	}
	if partNum < 1 || partNum > totalParts {
		return nil, ErrPartNum
	}

	params := &s3.UploadPartInput{
		Bucket:     aws.String(i.bucket),
		Key:        aws.String(filename),
		UploadId:   aws.String(uploadID),
		PartNumber: aws.Int64(partNum),
		Body:       bytes.NewReader(data),
	}
	if err := params.Validate(); err != nil {
		return nil, errors.Wrap(err, "storage.uploadPart: invalid params")
	}

	partResp, err := i.s3.UploadPart(params)
	if err != nil {
		return nil, errors.Wrap(err, "storage.uploadPart")
	}

	return &s3.CompletedPart{
		ETag:       aws.String(strings.Trim(*partResp.ETag, "\"")),
		PartNumber: aws.Int64(partNum),
	}, nil
}
