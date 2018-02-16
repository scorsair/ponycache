package main

import (
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/go-redis/redis"
	"os"
	"path"
	"strings"
)

func initAwsConfig(c *Config) *aws.Config {
	creds := credentials.NewStaticCredentials(
		c.Storage.Access_key,
		c.Storage.Secret_key,
		c.Storage.Token,
	)

	_, err := creds.Get()
	if err != nil {
		fmt.Printf("Bad credentials: %s\n", err)
		os.Exit(1)
	}

	cfg := aws.NewConfig().WithRegion(c.Storage.Region).WithCredentials(creds).WithEndpoint(c.Storage.Endpoint)

	// return s3.New(session.New(), cfg)
	return cfg
}

func upload(e Entry, c *Config, s *aws.Config, r *redis.Client) bool {
	if !isFileExist(e.value) {
		fmt.Printf("File %s doesnt exist!", e.value)
		return false
	}
	e.ttl = 0
	if e.sha1 == "" {
		fmt.Println("Whereis sha1 motherfucker?!")
		return false
	}

	basename := path.Base(e.value)
	sess, err := session.NewSession(s)
	s3_key := fmt.Sprintf("%s/%s/file", e.id, e.key)

	// Create an uploader with the session and default options
	uploader := s3manager.NewUploader(sess)
	file, err := os.OpenFile(e.value, os.O_RDONLY, 0444)

	if err != nil {
		fmt.Printf("Failed to open file %s, %v", basename, err)
		return false
	}

	// Upload the file to S3.
	result, err := uploader.Upload(&s3manager.UploadInput{
		Bucket: aws.String(c.Storage.Bucket),
		Key:    aws.String(s3_key),
		Body:   file,
	})
	if err != nil {
		fmt.Errorf("Failed to upload file, %v", err)
		return false
	}
	setEnv(r, e)
	fmt.Printf("file uploaded to, %s\n", result.Location)
	return true
}

func fillKeysFromBucket(e Entry, c *Config, s *aws.Config, r *redis.Client) []Entry {
	var entries []Entry
	sess, _ := session.NewSession(s)
	svc := s3.New(sess)
	prefix := fmt.Sprintf("%s/", e.id)
	newparams := &s3.ListObjectsInput{
		Bucket: aws.String(c.Storage.Bucket),
		Prefix: aws.String(prefix),
	}
	newresp, _ := svc.ListObjects(newparams)
	for _, key := range newresp.Contents {
		resp := strings.SplitN(*key.Key, "/", 3)
		if *key.Size != 0 {
			tmp := Entry{id: e.id, key: resp[1]}
			fillEntry(&tmp)
			entries = append(entries, tmp)
		}

	}
	return entries
}

func download(e Entry, c *Config, s *aws.Config, r *redis.Client) bool {
	var basename string
	if isFileExist(e.value) {
		fmt.Printf("File %s already exist!\n", e.value)
		return false
	}
	e.ttl = 0
	if e.sha1 == "" {
		fmt.Println("Whereis sha1 motherfucker?!")
		return false
	}
	// fmt.Printf("%+v", e)
	if e.value != "" {
		basename = path.Base(e.value)
	} else {
		basename, _ = getEnv(r, e.sha1)
	}

	if isFileExist(basename) {
		fmt.Printf("File %s already exist!\n", basename)
		return false
	}

	sess, err := session.NewSession(s)

	downloader := s3manager.NewDownloader(sess)

	fl, err := os.Create(basename)
	defer fl.Close()
	if err != nil {
		fmt.Errorf("Failed to create file %q, %v", basename, err)
	}

	s3_key := fmt.Sprintf("%s/%s/file", e.id, e.key)

	n, err := downloader.Download(fl, &s3.GetObjectInput{
		Bucket: aws.String(c.Storage.Bucket),
		Key:    aws.String(s3_key),
	})

	if err != nil {
		fmt.Errorf("failed to download file, %v", err)
	}
	fmt.Printf("%s file downloaded, %d bytes\n", basename, n)
	return true
}
