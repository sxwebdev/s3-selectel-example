package awss3

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/service/s3/types"
)

// IFiles files interface
type IFiles interface {
	List(ctx context.Context, bucket string) ([]string, error)
	Download(ctx context.Context, bucket, filePath string) ([]byte, error)
	Upload(ctx context.Context, bucket, fileName string, content []byte) (string, error)
	Delete(ctx context.Context, bucket string, filePaths []string) error
	ListBuckets(ctx context.Context) ([]types.Bucket, error)
	CreateBucket(ctx context.Context, name string) error
	BucketExists(ctx context.Context, bucket string) (bool, error)
	DeleteBucket(ctx context.Context, bucket string) error
}
