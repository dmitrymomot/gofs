package storage

import (
	"bytes"
	"fmt"
	"io"
	"sort"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/pkg/errors"
)

type (
	// Interactor struct
	Interactor struct {
		s3             *s3.S3
		bucket         string
		fileEndpoint   string
		forcePathStyle bool
	}

	// CompletedPart represents a part of a multipart upload.
	CompletedPart interface {
		PartNumber() int64
		ETag() string
	}

	completedPart struct {
		partNumber int64
		etag       string
	}
)

// PartNumber returns the part number.
func (p *completedPart) PartNumber() int64 {
	return p.partNumber
}

// ETag returns the ETag.
func (p *completedPart) ETag() string {
	return p.etag
}

// New is a factory function,
// returns a new instance of the storage interactor
func New(s3Client *s3.S3, bucket, fileEndpoint string) *Interactor {
	return &Interactor{
		s3:             s3Client,
		bucket:         bucket,
		fileEndpoint:   fileEndpoint,
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

// Delete file from the cloud storage
func (i *Interactor) Delete(filepath string) error {
	input := &s3.DeleteObjectInput{
		Bucket: aws.String(i.bucket),
		Key:    aws.String(filepath),
	}
	if err := input.Validate(); err != nil {
		return errors.Wrap(err, "storage.delete")
	}

	if _, err := i.s3.DeleteObject(input); err != nil {
		return errors.Wrap(err, "storage.delete")
	}

	return nil
}

// FileURL return public url for a file
func (i *Interactor) FileURL(filepath string) string {
	if i.forcePathStyle {
		return fmt.Sprintf("%s/%s/%s", i.fileEndpoint, i.bucket, filepath)
	}

	return fmt.Sprintf("%s/%s", i.fileEndpoint, filepath)
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
func (i *Interactor) CompleteMultipartUpload(filename, uploadID string, completedParts ...CompletedPart) error {
	if uploadID == "" {
		return ErrMissedUploadID
	}
	if len(completedParts) == 0 {
		return ErrNoCompletedParts
	}

	// Ordering the array based on the PartNumber as each parts could be uploaded in different order!
	sort.Slice(completedParts, func(i, j int) bool {
		return completedParts[i].PartNumber() < completedParts[j].PartNumber()
	})

	// Converting the CompletedPart to s3.CompletedPart
	parts := make([]*s3.CompletedPart, len(completedParts))
	for i, part := range completedParts {
		parts[i] = &s3.CompletedPart{
			ETag:       aws.String(part.ETag()),
			PartNumber: aws.Int64(part.PartNumber()),
		}
	}

	params := &s3.CompleteMultipartUploadInput{
		Bucket:          aws.String(i.bucket),
		Key:             aws.String(filename),
		UploadId:        aws.String(uploadID),
		MultipartUpload: &s3.CompletedMultipartUpload{Parts: parts},
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
func (i *Interactor) UploadPart(filename, uploadID string, data []byte, partNum, totalParts int64) (CompletedPart, error) {
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

	return &completedPart{
		etag:       *partResp.ETag,
		partNumber: partNum,
	}, nil
}
