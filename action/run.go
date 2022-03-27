package main

import (
	"context"
	"fmt"
	"os"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/aws/smithy-go/logging"
)

type s3Client struct {
	ctx    context.Context
	logger logging.Logger
	client *s3.Client
}

func newS3Client() (*s3Client, error) {
	ctx := context.Background()
	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		return nil, fmt.Errorf("error loading default config: %v", err)
	}

	client := s3.NewFromConfig(cfg)

	return &s3Client{
		ctx:    ctx,
		logger: cfg.Logger,
		client: client,
	}, nil
}

func (s3c *s3Client) uploadToAWS(bucket *string, key *string, filePath string) error {
	file, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("error opening file: %v", err)
	}

	defer closeFile(file)

	putObjectInput := &s3.PutObjectInput{
		Bucket: bucket,
		Key: key,
		ACL: types.ObjectCannedACLPublicRead,
		Body: file,
	}

	if _, err = s3c.client.PutObject(s3c.ctx, putObjectInput); err != nil {
		return fmt.Errorf("error putting file into S3 bucket: %v", err)
	}

	s3c.logger.Logf("INFO", "%s was uploaded successfully into S3 bucket", filePath)
	return nil
}

func closeFile(file *os.File) {
	if err := file.Close(); err != nil {
		panic(fmt.Errorf("error closing the file %s: %v", file.Name(), err))
	}
}

func main() {
	s, err := newS3Client()
	if err != nil {
		panic(fmt.Errorf("error creating S3 client: %v", err))
	}

	bucketName := os.Getenv("BUCKET_NAME")
	bucketDirectory := os.Getenv("BUCKET_DIR")
	srcDirectory := os.Getenv("SRC_DIR")
	autoDeploymentFile := os.Getenv("AUTO_DEPLOYMENT_FILE")
	functionZipFile := os.Getenv("FUNCTION_ZIP_FILE")
	autoDeploymentKey := bucketDirectory + "/" + autoDeploymentFile
	functionZipKey := bucketDirectory + "/" + functionZipFile
	autoDeploymentSrcPath := srcDirectory + "/" + autoDeploymentFile
	functionZipSrcPath := srcDirectory + "/" + functionZipFile

	if err = s.uploadToAWS(&bucketName, &autoDeploymentKey, autoDeploymentSrcPath); err != nil {
		panic(fmt.Errorf("error uploading auto deployment file into S3 bucket: %v", err))
	}

	if err = s.uploadToAWS(&bucketName, &functionZipKey, functionZipSrcPath); err != nil {
		panic(fmt.Errorf("error uploading function zip file into S3 bucket: %v", err))
	}
}