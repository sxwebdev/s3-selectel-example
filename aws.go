package awss3

import (
	"bytes"
	"context"
	"fmt"
	"path/filepath"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/feature/s3/manager"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/aws/smithy-go"
	"github.com/pkg/errors"
	"github.com/tkcrm/modules/pkg/logger"
)

// S3 ...
type S3 struct {
	logger     logger.Logger
	cfg        *Config
	svc        *s3.Client
	uploader   *manager.Uploader
	downloader *manager.Downloader
}

// New ...
func New(l logger.Logger, cfg *Config) (*S3, error) {
	s := &S3{
		logger: l,
		cfg:    cfg,
	}

	customEndpointResolver := aws.EndpointResolverWithOptionsFunc(func(service, region string, options ...interface{}) (aws.Endpoint, error) {
		return aws.Endpoint{
			PartitionID:       "selectel",
			URL:               cfg.Endpoint,
			SigningRegion:     cfg.Region,
			HostnameImmutable: true,
		}, nil
	})

	awsCfg, err := config.LoadDefaultConfig(
		context.TODO(),
		config.WithRegion(cfg.Region),
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(cfg.AccessID, cfg.SecretKey, cfg.Token)),
		config.WithEndpointResolverWithOptions(customEndpointResolver),
		config.WithLogger(logger.New(logger.WithConsoleColored(true), logger.WithLogFormat(logger.FORMAT_CONSOLE))),
		config.WithClientLogMode(aws.LogRetries|aws.LogRequest),
	)
	if err != nil {
		return nil, err
	}

	s.svc = s3.NewFromConfig(awsCfg)
	s.uploader = manager.NewUploader(s.svc)
	s.downloader = manager.NewDownloader(s.svc)

	return s, nil
}

// List ...
func (s *S3) List(ctx context.Context, bucket string) ([]string, error) {
	if bucket == "" {
		return nil, ErrEmptyBucket
	}

	res, err := s.svc.ListObjects(ctx, &s3.ListObjectsInput{
		Bucket: aws.String(bucket),
	})
	if err != nil {
		return nil, errors.Wrap(err, "ListObjects error")
	}

	result := make([]string, 0, len(res.Contents))
	for _, value := range res.Contents {
		if value.Key != nil && *value.Key != "" && filepath.Ext(*value.Key) != "" {
			result = append(result, *value.Key)
		}
	}

	return result, nil
}

// Upload file to s3 bucket
func (s *S3) Upload(ctx context.Context, bucket, filePath string, content []byte) (string, error) {
	if bucket == "" {
		return "", ErrEmptyBucket
	}

	if len(content) == 0 {
		return "", errors.New("empty file content")
	}

	result, err := s.uploader.Upload(ctx, &s3.PutObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(filePath),
		Body:   bytes.NewBuffer(content),
	})
	if err != nil {
		return "", fmt.Errorf("failed to upload file: %v", err)
	}

	if result.Key == nil {
		return "", fmt.Errorf("received empty s3 file key")
	}

	return *result.Key, nil
}

// Download file from s3 bucket
func (s *S3) Download(ctx context.Context, bucket, filePath string) ([]byte, error) {
	if bucket == "" {
		return nil, ErrEmptyBucket
	}

	if filePath == "" {
		return nil, errors.New("empty file path")
	}

	buff := manager.NewWriteAtBuffer([]byte{})
	if _, err := s.downloader.Download(ctx, buff, &s3.GetObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(filePath),
	}); err != nil {
		return nil, fmt.Errorf("failed to download file, %v", err)
	}

	return buff.Bytes(), nil
}

// Delete - Delete directory with files from s3 bucket
func (s *S3) Delete(ctx context.Context, bucket string, filePaths []string) error {
	if bucket == "" {
		return ErrEmptyBucket
	}

	if len(filePaths) == 0 {
		return errors.New("empty filePaths array")
	}

	var objectIds []types.ObjectIdentifier
	for _, key := range filePaths {
		objectIds = append(objectIds, types.ObjectIdentifier{Key: aws.String(key)})
	}

	delResp, err := s.svc.DeleteObjects(ctx, &s3.DeleteObjectsInput{
		Bucket: aws.String(bucket),
		Delete: &types.Delete{Objects: objectIds},
	})
	if err != nil {
		return errors.Wrap(err, "DeleteObjects error")
	}

	for _, f := range filePaths {
		var exist bool
		for _, delF := range delResp.Deleted {
			if delF.Key != nil && *delF.Key == f {
				exist = true
				break
			}
		}
		if !exist {
			return fmt.Errorf("file \"%s\" was not deleted", f)
		}
	}

	return nil
}

func (s *S3) ListBuckets(ctx context.Context) ([]types.Bucket, error) {
	result, err := s.svc.ListBuckets(ctx, &s3.ListBucketsInput{})
	if err != nil {
		return nil, fmt.Errorf("couldn't list buckets for your account. Here's why: %v", err)
	}

	return result.Buckets, err
}

func (s *S3) BucketExists(ctx context.Context, bucket string) (bool, error) {
	_, err := s.svc.HeadBucket(ctx, &s3.HeadBucketInput{
		Bucket: aws.String(bucket),
	})
	exists := true
	if err != nil {
		var apiError smithy.APIError
		if errors.As(err, &apiError) {
			switch apiError.(type) {
			case *types.NotFound:
				exists = false
				err = nil
			}
		}
	}

	return exists, err
}

func (s *S3) CreateBucket(ctx context.Context, bucket string) error {
	_, err := s.svc.CreateBucket(ctx, &s3.CreateBucketInput{
		Bucket: aws.String(bucket),
		CreateBucketConfiguration: &types.CreateBucketConfiguration{
			LocationConstraint: types.BucketLocationConstraint(s.cfg.Region),
		},
	})
	if err != nil {
		return fmt.Errorf("couldn't create bucket %v in Region %v. Here's why: %v",
			bucket, s.cfg.Region, err,
		)
	}

	return nil
}

func (s *S3) DeleteBucket(ctx context.Context, bucket string) error {
	if _, err := s.svc.DeleteBucket(ctx, &s3.DeleteBucketInput{
		Bucket: aws.String(bucket),
	}); err != nil {
		return fmt.Errorf("couldn't delete bucket %v. Here's why: %v", bucket, err)
	}

	return nil
}
