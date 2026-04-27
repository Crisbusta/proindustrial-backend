package storage

// S3Provider implements Provider for AWS S3 and Cloudflare R2 (S3-compatible).
//
// To activate:
//  1. Add dependency:  go get github.com/aws/aws-sdk-go-v2/...
//  2. Set env vars:    STORAGE_DRIVER=s3
//                      S3_BUCKET, S3_REGION, S3_ENDPOINT (R2 only),
//                      AWS_ACCESS_KEY_ID, AWS_SECRET_ACCESS_KEY
//  3. Uncomment the implementation below and delete this comment block.
//
// Cloudflare R2 endpoint format:
//   https://<account-id>.r2.cloudflarestorage.com
//
// The rest of the codebase does not change — Provider is the contract.

// Uncomment when ready:
//
// import (
// 	"context"
// 	"io"
//
// 	"github.com/aws/aws-sdk-go-v2/aws"
// 	"github.com/aws/aws-sdk-go-v2/config"
// 	"github.com/aws/aws-sdk-go-v2/credentials"
// 	"github.com/aws/aws-sdk-go-v2/service/s3"
// )
//
// type S3Provider struct {
// 	client   *s3.Client
// 	bucket   string
// 	publicBase string // e.g. "https://cdn.proindustrial.cl"
// }
//
// func NewS3Provider(bucket, region, endpoint, accessKey, secretKey, publicBase string) (*S3Provider, error) {
// 	opts := []func(*config.LoadOptions) error{
// 		config.WithRegion(region),
// 		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(accessKey, secretKey, "")),
// 	}
// 	cfg, err := config.LoadDefaultConfig(context.Background(), opts...)
// 	if err != nil {
// 		return nil, err
// 	}
// 	var clientOpts []func(*s3.Options)
// 	if endpoint != "" {
// 		clientOpts = append(clientOpts, func(o *s3.Options) {
// 			o.BaseEndpoint = aws.String(endpoint)
// 			o.UsePathStyle = true
// 		})
// 	}
// 	return &S3Provider{client: s3.NewFromConfig(cfg, clientOpts...), bucket: bucket, publicBase: publicBase}, nil
// }
//
// func (p *S3Provider) Upload(ctx context.Context, key string, r io.Reader, size int64, contentType string) (string, error) {
// 	_, err := p.client.PutObject(ctx, &s3.PutObjectInput{
// 		Bucket:        aws.String(p.bucket),
// 		Key:           aws.String(key),
// 		Body:          r,
// 		ContentLength: aws.Int64(size),
// 		ContentType:   aws.String(contentType),
// 	})
// 	if err != nil {
// 		return "", err
// 	}
// 	return p.PublicURL(key), nil
// }
//
// func (p *S3Provider) Delete(ctx context.Context, key string) error {
// 	_, err := p.client.DeleteObject(ctx, &s3.DeleteObjectInput{
// 		Bucket: aws.String(p.bucket),
// 		Key:    aws.String(key),
// 	})
// 	return err
// }
//
// func (p *S3Provider) PublicURL(key string) string {
// 	return p.publicBase + "/" + key
// }
