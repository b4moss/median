package s3store

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"path"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/b4moss/median/go/storage"
)

const Driver = "s3"

var (
	ErrInvalidConfig  = errors.New("s3: invalid config")
	ErrNotPresignable = errors.New("s3: cannot presign")
)

type Config struct {
	Bucket         string
	Region         string
	Endpoint       string
	ForcePathStyle bool
	Prefix         string
	AccessKey      string
	SecretKey      string
	PresignTTL     time.Duration
}

type API interface {
	PutObject(ctx context.Context, params *s3.PutObjectInput, optFns ...func(*s3.Options)) (*s3.PutObjectOutput, error)
	GetObject(ctx context.Context, params *s3.GetObjectInput, optFns ...func(*s3.Options)) (*s3.GetObjectOutput, error)
	DeleteObject(ctx context.Context, params *s3.DeleteObjectInput, optFns ...func(*s3.Options)) (*s3.DeleteObjectOutput, error)
	HeadObject(ctx context.Context, params *s3.HeadObjectInput, optFns ...func(*s3.Options)) (*s3.HeadObjectOutput, error)
}

type Presigner interface {
	PresignGET(ctx context.Context, bucket, key string, ttl time.Duration) (string, error)
}

type FS struct {
	cfg       Config
	client    API
	presigner Presigner
}

func NewFromClient(cfg Config, client API, presigner Presigner) (*FS, error) {
	if strings.TrimSpace(cfg.Bucket) == "" {
		return nil, fmt.Errorf("%w: bucket required", ErrInvalidConfig)
	}
	if cfg.PresignTTL == 0 {
		cfg.PresignTTL = time.Hour
	}
	if client == nil {
		return nil, fmt.Errorf("%w: client required", ErrInvalidConfig)
	}
	return &FS{cfg: cfg, client: client, presigner: presigner}, nil
}

func New(ctx context.Context, cfg Config) (*FS, error) {
	if strings.TrimSpace(cfg.Bucket) == "" {
		return nil, fmt.Errorf("%w: bucket required", ErrInvalidConfig)
	}
	if cfg.Region == "" {
		cfg.Region = "us-east-1"
	}
	if cfg.PresignTTL == 0 {
		cfg.PresignTTL = time.Hour
	}
	awsCfg := aws.Config{Region: cfg.Region}
	if cfg.AccessKey != "" {
		awsCfg.Credentials = credentials.NewStaticCredentialsProvider(cfg.AccessKey, cfg.SecretKey, "")
	}
	client := s3.NewFromConfig(awsCfg, func(o *s3.Options) {
		if cfg.Endpoint != "" {
			o.BaseEndpoint = aws.String(cfg.Endpoint)
		}
		o.UsePathStyle = cfg.ForcePathStyle
	})
	ps := s3.NewPresignClient(client)
	return NewFromClient(cfg, client, sdkPresigner{ps})
}

type sdkPresigner struct {
	c *s3.PresignClient
}

func (p sdkPresigner) PresignGET(ctx context.Context, bucket, key string, ttl time.Duration) (string, error) {
	out, err := p.c.PresignGetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	}, s3.WithPresignExpires(ttl))
	if err != nil {
		return "", err
	}
	return out.URL, nil
}

func (f *FS) objectKey(relPath string) (string, error) {
	clean, err := cleanRel(relPath)
	if err != nil {
		return "", err
	}
	if f.cfg.Prefix == "" {
		return clean, nil
	}
	return path.Join(strings.Trim(f.cfg.Prefix, "/"), clean), nil
}

func cleanRel(relPath string) (string, error) {
	if relPath == "" {
		return "", storage.ErrInvalidPath
	}
	p := strings.ReplaceAll(relPath, "\\", "/")
	if strings.HasPrefix(p, "/") {
		return "", storage.ErrInvalidPath
	}
	for _, part := range strings.Split(p, "/") {
		if part == ".." {
			return "", storage.ErrInvalidPath
		}
	}
	return p, nil
}

func (f *FS) Put(ctx context.Context, relPath string, r io.Reader, size int64) error {
	key, err := f.objectKey(relPath)
	if err != nil {
		return err
	}
	body := r
	if size >= 0 {
		data, err := io.ReadAll(io.LimitReader(r, size))
		if err != nil {
			return err
		}
		if int64(len(data)) != size {
			return io.ErrUnexpectedEOF
		}
		body = bytes.NewReader(data)
	}
	_, err = f.client.PutObject(ctx, &s3.PutObjectInput{
		Bucket: aws.String(f.cfg.Bucket),
		Key:    aws.String(key),
		Body:   body,
	})
	return err
}

func (f *FS) Get(ctx context.Context, relPath string) (io.ReadCloser, int64, error) {
	key, err := f.objectKey(relPath)
	if err != nil {
		return nil, 0, err
	}
	out, err := f.client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(f.cfg.Bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		if isNotFound(err) {
			return nil, 0, storage.ErrNotFound
		}
		return nil, 0, err
	}
	var size int64
	if out.ContentLength != nil {
		size = *out.ContentLength
	}
	return out.Body, size, nil
}

func (f *FS) Delete(ctx context.Context, relPath string) error {
	key, err := f.objectKey(relPath)
	if err != nil {
		return err
	}
	_, err = f.client.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: aws.String(f.cfg.Bucket),
		Key:    aws.String(key),
	})
	return err
}

func (f *FS) Exists(ctx context.Context, relPath string) (bool, error) {
	key, err := f.objectKey(relPath)
	if err != nil {
		return false, err
	}
	_, err = f.client.HeadObject(ctx, &s3.HeadObjectInput{
		Bucket: aws.String(f.cfg.Bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		if isNotFound(err) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

func (f *FS) PresignGet(ctx context.Context, relPath string, ttl time.Duration) (string, error) {
	if f.presigner == nil {
		return "", ErrNotPresignable
	}
	if ttl <= 0 {
		ttl = f.cfg.PresignTTL
	}
	key, err := f.objectKey(relPath)
	if err != nil {
		return "", err
	}
	return f.presigner.PresignGET(ctx, f.cfg.Bucket, key, ttl)
}

func isNotFound(err error) bool {
	if err == nil {
		return false
	}
	s := err.Error()
	return strings.Contains(s, "NotFound") || strings.Contains(s, "404") || strings.Contains(s, "NoSuchKey")
}
