package awss3

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/davecgh/go-spew/spew"

	"github.com/tkcrm/modules/pkg/cfg"
	"github.com/tkcrm/modules/pkg/logger"
)

func getConfig() *Config {
	var config Config
	if err := cfg.LoadConfig(&config); err != nil {
		logger.New().Fatalf("could not load configuration: %v", err)
	}

	return &config
}

func newAWS() (IFiles, error) {
	c := getConfig()
	return New(
		logger.New(
			logger.WithConsoleColored(true),
			logger.WithLogFormat(logger.FORMAT_CONSOLE),
		),
		c,
	)
}

func Test_ListBuckets(t *testing.T) {
	s3, err := newAWS()
	if err != nil {
		t.Fatal(err)
	}

	result, err := s3.ListBuckets(context.Background())
	if err != nil {
		t.Fatal(err)
	}

	spew.Dump(result)
}

func Test_List(t *testing.T) {
	s3, err := newAWS()
	if err != nil {
		t.Fatal(err)
	}

	result, err := s3.List(context.Background(), "backups")
	if err != nil {
		t.Fatal(err)
	}

	spew.Dump(result)
}

func Test_Upload(t *testing.T) {
	s3, err := newAWS()
	if err != nil {
		t.Fatal(err)
	}

	f, err := os.ReadFile("/Users/uname/Downloads/voucher_6661575.pdf")
	if err != nil {
		t.Fatal(err)
	}

	if _, err := s3.Upload(context.Background(), "backup.new", "test/voucher_6661575.pdf", f); err != nil {
		t.Fatal(err)
	}
}

func Test_Download(t *testing.T) {
	s3, err := newAWS()
	if err != nil {
		t.Fatal(err)
	}

	res, err := s3.Download(context.Background(), "backup.new", "test/voucher_6661575.pdf")
	if err != nil {
		t.Fatal(err)
	}

	fmt.Println(string(res))
}

func Test_Delete(t *testing.T) {
	s3, err := newAWS()
	if err != nil {
		t.Fatal(err)
	}

	filesToDelete := []string{
		"test/folder/asdf.asdf/photo_2022-12-21 19.58.37.jpeg",
		"photo_2022-12-21 19.58.37.jpeg",
	}

	if err := s3.Delete(context.Background(), "test.bucket", filesToDelete); err != nil {
		t.Fatal(err)
	}
}

func Test_CreateBucket(t *testing.T) {
	s3, err := newAWS()
	if err != nil {
		t.Fatal(err)
	}

	if err := s3.CreateBucket(context.Background(), "test-bucket"); err != nil {
		t.Fatal(err)
	}
}

func Test_BucketExists(t *testing.T) {
	s3, err := newAWS()
	if err != nil {
		t.Fatal(err)
	}

	exists, err := s3.BucketExists(context.Background(), "test-bucket")
	if err != nil {
		t.Fatal(err)
	}

	spew.Dump(exists)
}

func Test_DeleteBucket(t *testing.T) {
	s3, err := newAWS()
	if err != nil {
		t.Fatal(err)
	}

	if err := s3.DeleteBucket(context.Background(), "test-bucket"); err != nil {
		t.Fatal(err)
	}
}
