package main

import (
	"bytes"
	"fmt"
	"net/http"
	"os"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3"
)

type S3Client struct {
	bucketName string
	prefix     string
	client     *s3.S3
	err        error
}

// Setup s3 info
func (s *S3Client) SetupS3(bucketName string, prefix string, client *s3.S3, err error) {
	s.bucketName = bucketName
	s.prefix = prefix
	s.client = client
	s.err = err
}

// Get s3 info
func (s S3Client) GetS3Info() S3Client {
	return s
}

// Initialize s3 client
func NewS3Client(awsConfig *AWSConfig, bucketName string, prefix string) S3Client {
	_s3 := S3Client{}

	awsSession, err := awsConfig.CreateAWSSession()
	if err != nil {
		_s3.err = err
		return _s3.GetS3Info()
	}

	// now use the assumed role to connect to s3
	s3Client := s3.New(awsSession)

	_s3.SetupS3(bucketName, prefix, s3Client, nil)

	return _s3.GetS3Info()
}

func UploadPlanFile(s S3Client, key string, outputFilePath string) error {
	file, err := os.Open(outputFilePath)
	if err != nil {
		return fmt.Errorf("failed to open file %q: %v", outputFilePath, err)
	}

	defer file.Close()

	fileInfo, _ := file.Stat()
	var size int64 = fileInfo.Size()

	buffer := make([]byte, size)
	// read file content to buffer
	file.Read(buffer)
	fileBytes := bytes.NewReader(buffer)
	fileType := http.DetectContentType(buffer)

	keyWithPrefix := fmt.Sprintf("%s/%s", s.prefix, key)

	// Create put object
	params := &s3.PutObjectInput{
		Bucket:        aws.String(s.bucketName),  // required
		Key:           aws.String(keyWithPrefix), // required
		Body:          fileBytes,
		ContentLength: aws.Int64(size),
		ContentType:   aws.String(fileType),
		Metadata: map[string]*string{
			"Key": aws.String("MetadataValue"), //required
		},
	}

	_, err = s.client.PutObject(params)
	if err != nil {
		return fmt.Errorf("failed to upload plan file to S3: %v", err)
	}
	return nil
}
