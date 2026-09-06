package api_bucket

import (
	"context"
	"fmt"
	"io"

	"github.com/Johncode248/Rodzic-Chat/logger"
	"github.com/minio/minio-go/v7"
)

func UploadFile(ctx context.Context, bucketName string, objectName string, reader io.Reader, size int64, contentType string) (err error) {
	ctx = logger.Enter(ctx, "api_wasabi.UploadFile", bucketName, objectName, size)
	defer logger.Exit(ctx, "api_wasabi.UploadFile", &err)

	if contentType == "" {
		contentType = "application/octet-stream"
	}

	_, err = ClientBucket.PutObject(ctx, bucketName, objectName, reader, size, minio.PutObjectOptions{
		ContentType: contentType,
	})
	if err != nil {
		return fmt.Errorf("błąd uploadu pliku: %v", err)
	}

	return nil
}
