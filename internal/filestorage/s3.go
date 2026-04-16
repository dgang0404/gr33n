package filestorage

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"net/url"
	"strings"
	"time"

	v4 "github.com/aws/aws-sdk-go-v2/aws/signer/v4"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

type s3API interface {
	PutObject(ctx context.Context, params *s3.PutObjectInput, optFns ...func(*s3.Options)) (*s3.PutObjectOutput, error)
	GetObject(ctx context.Context, params *s3.GetObjectInput, optFns ...func(*s3.Options)) (*s3.GetObjectOutput, error)
	DeleteObject(ctx context.Context, params *s3.DeleteObjectInput, optFns ...func(*s3.Options)) (*s3.DeleteObjectOutput, error)
}

type s3PresignAPI interface {
	PresignGetObject(ctx context.Context, params *s3.GetObjectInput, optFns ...func(*s3.PresignOptions)) (*v4.PresignedHTTPRequest, error)
}

// S3 stores blobs in an S3-compatible bucket.
type S3 struct {
	client  s3API
	presign s3PresignAPI
	bucket  string
	prefix  string
}

func NewS3(ctx context.Context, cfg Config) (*S3, error) {
	if cfg.S3Bucket == "" {
		return nil, errors.New("S3_BUCKET is required when FILE_STORAGE_BACKEND=s3")
	}
	region := cfg.S3Region
	if region == "" {
		region = "auto"
	}
	loadOpts := []func(*awsconfig.LoadOptions) error{
		awsconfig.WithRegion(region),
	}
	if cfg.S3AccessKeyID != "" || cfg.S3SecretKey != "" {
		if cfg.S3AccessKeyID == "" || cfg.S3SecretKey == "" {
			return nil, errors.New("S3_ACCESS_KEY_ID and S3_SECRET_ACCESS_KEY must both be set")
		}
		loadOpts = append(loadOpts, awsconfig.WithCredentialsProvider(
			credentials.NewStaticCredentialsProvider(cfg.S3AccessKeyID, cfg.S3SecretKey, ""),
		))
	}
	awsCfg, err := awsconfig.LoadDefaultConfig(ctx, loadOpts...)
	if err != nil {
		return nil, fmt.Errorf("load s3 config: %w", err)
	}
	var baseEndpoint string
	if cfg.S3Endpoint != "" {
		baseEndpoint, err = normalizeEndpoint(cfg.S3Endpoint, cfg.S3DisableHTTPS)
		if err != nil {
			return nil, fmt.Errorf("normalize S3_ENDPOINT: %w", err)
		}
	}

	opts := func(o *s3.Options) {
		o.UsePathStyle = cfg.S3UsePathStyle
		if baseEndpoint != "" {
			o.BaseEndpoint = &baseEndpoint
		}
	}
	return &S3{
		client:  s3.NewFromConfig(awsCfg, opts),
		presign: s3.NewPresignClient(s3.NewFromConfig(awsCfg, opts)),
		bucket:  cfg.S3Bucket,
		prefix:  cfg.S3Prefix,
	}, nil
}

func (s *S3) Backend() string { return "s3" }

func (s *S3) Put(ctx context.Context, key string, r io.Reader, maxBytes int64) (int64, error) {
	objectKey := s.objectKey(key)
	lr := &io.LimitedReader{R: r, N: maxBytes + 1}
	data, err := io.ReadAll(lr)
	if err != nil {
		return 0, err
	}
	if int64(len(data)) > maxBytes {
		return 0, fmt.Errorf("file too large")
	}
	_, err = s.client.PutObject(ctx, &s3.PutObjectInput{
		Bucket:        &s.bucket,
		Key:           &objectKey,
		Body:          bytes.NewReader(data),
		ContentLength: int64Ptr(int64(len(data))),
	})
	if err != nil {
		return 0, err
	}
	return int64(len(data)), nil
}

func (s *S3) Open(ctx context.Context, key string) (io.ReadCloser, error) {
	objectKey := s.objectKey(key)
	out, err := s.client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: &s.bucket,
		Key:    &objectKey,
	})
	if err != nil {
		return nil, err
	}
	return out.Body, nil
}

func (s *S3) Delete(ctx context.Context, key string) error {
	objectKey := s.objectKey(key)
	_, err := s.client.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: &s.bucket,
		Key:    &objectKey,
	})
	return err
}

func (s *S3) DownloadURL(ctx context.Context, key, filename, mime string, ttl time.Duration) (string, error) {
	objectKey := s.objectKey(key)
	params := &s3.GetObjectInput{
		Bucket: &s.bucket,
		Key:    &objectKey,
	}
	if mime != "" {
		params.ResponseContentType = &mime
	}
	if filename != "" {
		disposition := fmt.Sprintf(`inline; filename="%s"`, strings.ReplaceAll(filename, `"`, ""))
		params.ResponseContentDisposition = &disposition
	}
	req, err := s.presign.PresignGetObject(ctx, params, func(o *s3.PresignOptions) {
		o.Expires = ttl
	})
	if err != nil {
		return "", err
	}
	return req.URL, nil
}

func (s *S3) objectKey(key string) string {
	key = strings.TrimLeft(key, "/")
	if s.prefix == "" {
		return key
	}
	return strings.Trim(s.prefix, "/") + "/" + key
}

func normalizeEndpoint(raw string, disableHTTPS bool) (string, error) {
	if strings.TrimSpace(raw) == "" {
		return "", nil
	}
	if !strings.Contains(raw, "://") {
		if disableHTTPS {
			raw = "http://" + raw
		} else {
			raw = "https://" + raw
		}
	}
	u, err := url.Parse(raw)
	if err != nil {
		return "", err
	}
	return u.String(), nil
}

func int64Ptr(v int64) *int64 { return &v }
