package storage

import "github.com/aws/aws-sdk-go/service/s3"

// Predefined ACL permissions
const (
	Public  ACL = s3.ObjectCannedACLPublicRead
	Private ACL = s3.ObjectCannedACLPrivate
)

// ACL permission
type ACL string

// String returns the string representation of the ACL.
// This is required to satisfy the Stringer interface.
func (a ACL) String() string {
	return string(a)
}
